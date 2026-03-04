package auth

import (
	"4-order-api/config"
	"4-order-api/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"crypto/rand"
	"math/big"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	config *config.Config
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:     db,
		config: cfg,
	}
}

func generateSecureCode() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()) + 1000, nil
}

func (h *Handler) SendSMS(w http.ResponseWriter, r *http.Request) {

	// 1. Декодировать JSON с номером телефона

	var req SendSMSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 2. Валидировать формат телефона (начинается с 8/+7, 11 цифр)

	phoneRegex := regexp.MustCompile(`^[78]\d{10}$`)

	if !phoneRegex.MatchString(req.Phone) {
		http.Error(w, "Invalid phone format. Use 89990009900", http.StatusBadRequest)
		return
	}

	// 3. Проверка на спам (не более 3 SMS за последние 10 минут)

	var recentSessions int64

	tenMinutesAgo := time.Now().Add(-10 * time.Minute)

	h.db.Model(&models.Session{}).Where("phone = ? AND created_at > ?", req.Phone, tenMinutesAgo).Count(&recentSessions)

	if recentSessions >= 3 {
		http.Error(w, "Too many requests. Try again later", http.StatusTooManyRequests)
		return
	}

	// 4. Создаем или находим пользователя
	var user models.User

	result := h.db.Where("phone = ?", req.Phone).FirstOrCreate(&user, models.User{Phone: req.Phone})
	if result.Error != nil {
		log.Printf("Error creating user: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 5. Генерируем уникальный sessionID

	sessionID := uuid.New().String()

	// 6. Генерируем 4-значный код (1000-9999)
	code, err := generateSecureCode()

	if err != nil {
		log.Printf("Error generating code: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 7. Создаем сессию в БД (действительна 5 минут)

	session := models.Session{
		SessionID: sessionID,
		Phone:     req.Phone,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	if err := h.db.Create(&session).Error; err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 8. "Отправляем SMS" (в реальности - логируем в консоль)
	log.Printf("📱 SMS для %s: Ваш код подтверждения: %d (действителен 5 минут)", req.Phone, code)

	// 9. Возвращаем sessionID клиенту

	response := SendSMSResponse{
		SessionID: sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func (h *Handler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	// 1. Декодируем JSON запрос
	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 2. Ищем сессию в БД
	var session models.Session

	result := h.db.Where("session_id = ?", req.SessionID).First(&session)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}
		log.Printf("Database error: %v", result.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 3. Проверяем валидность сессии

	if session.Used {
		http.Error(w, "Code already used", http.StatusUnauthorized)
		return
	}

	if time.Now().After(session.ExpiresAt) {
		http.Error(w, "Code expired", http.StatusUnauthorized)
		return
	}

	if session.Code != req.Code {
		http.Error(w, "Invalid code", http.StatusUnauthorized)
		return
	}

	// 4. Помечаем сессию как использованную

	session.Used = true

	if err := h.db.Save(&session).Error; err != nil {
		log.Printf("Error updating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 5. Находим пользователя

	var user models.User

	if err := h.db.Where("phone = ?", session.Phone).First(&user).Error; err != nil {
		log.Printf("User not found: %v", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// 6. Генерируем JWT токен

	token, err := h.generateJWT(user.ID, user.Phone)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		http.Error(w, "Internal sever error", http.StatusInternalServerError)
		return
	}

	// 7. Возвращаем токен

	response := VerifyCodeResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("✅ Пользователь %s успешно авторизован", user.Phone)

}

func (h *Handler) generateJWT(userID uint, phone string) (string, error) {
	// Создаём claims (полезная нагрузка токена)
	claims := jwt.MapClaims{
		"user_id": userID,
		"phone":   phone,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Токен действителен 24 часа
		"iat":     time.Now().Unix(),                     // Время создания
	}

	// Создаём токен с алгоритмом HMAC SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен секретным ключом
	tokenString, err := token.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
