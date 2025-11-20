package handler

import (
	"crypto/rand"   // Для генерации криптографически безопасных случайных чисел
	"encoding/hex"  // Для преобразования байтов в hex-строку
	"encoding/json" // Для работы с JSON
	"fmt"           // Для форматирования строк
	"net/http"      // Для работы с HTTP
	"net/mail"
	"net/smtp" // Для отправки email через SMTP

	"3-validation-api/config" // Импорт нашего пакета config
	"3-validation-api/storage"

	"github.com/jordan-wright/email" // Импорт внешней библиотеки для email
)

// Определение структуры VerifyHandler
type VerifyHandler struct {
	// Поле cfg хранит указатель на конфигурацию
	cfg *config.Config

	storage *storage.Storage
}

// Конструктор для VerifyHandler
func NewVerifyHandler(cfg *config.Config, storage *storage.Storage) *VerifyHandler {
	// Возвращает указатель на новый VerifyHandler с переданной конфигурацией
	return &VerifyHandler{
		cfg:     cfg,
		storage: storage,
	}
}

// SendRequest структуры для запроса отправки verification email

type SendRequest struct {
	Email string `json:"email"`
}

// VerifyResponse структура для ответа проверки hash
type VerifyResponse struct {
	Valid bool `json:"valid"`
}

// Обработчик для отправки verification email
func (h *VerifyHandler) Send(w http.ResponseWriter, r *http.Request) {
	// Проверка что метод запроса именно POST
	if r.Method != http.MethodPost {
		// Если метод не POST - возвращаем ошибку 405
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		// Выход из функции
		return
	}

	// Декодирование JSON тела запроса

	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация email адреса

	if _, err := mail.ParseAddress(req.Email); err != nil {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	// Генерация уникального hash

	hash, err := generateHash()
	if err != nil {
		http.Error(w, "Error generation hash", http.StatusInternalServerError)
		return
	}

	// Сохранение данных в хранилище и JSON файл
	if err := h.storage.Save(req.Email, hash); err != nil {
		http.Error(w, "Error saving verification data", http.StatusInternalServerError)
		return
	}

	// Отправка verification email
	if err := h.sendEmail(req.Email, hash); err != nil {
		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	// Успешный ответ

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Verififcation email sent",
		"hash":    hash,
	})

}

// Вспомогательный метод для отправки email

func (h *VerifyHandler) sendEmail(to, hash string) error {
	// Создание нового объекта email
	e := email.NewEmail()
	// Установка адреса отправилтеля из конфигурации
	e.From = h.cfg.Email
	// Установка адреса получателя
	e.To = []string{to}
	// Установка темы письма
	e.Subject = "Email Verification"

	//  Установка текстового содержимого письма с verification ссылкой

	// Формирование verification ссылки(порт 8081 согласно ТЗ)

	verificationURL := fmt.Sprintf("http://localhost:8081/verify/%s", hash)

	e.Text = []byte(fmt.Sprintf("Verify your email: http://localhost:8080/verify/%s", verificationURL))

	// Создание SMTP аутентификации

	auth := smtp.PlainAuth("", h.cfg.Email, h.cfg.Password, "smtp.gmail.com")
	// Отправка email через SMTP сервер

	return e.Send(h.cfg.Address, auth)
}

//Обработчик для подтверждения email по хешу

func (h *VerifyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	// Проверка что метод запроса именно GET

	if r.Method != http.MethodGet {
		// Если метод не GET - возвращаем ошибку 405
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получение значения хеша из URL пути (часть {hash})
	hash := r.PathValue("hash")

	if hash == "" {
		http.Error(w, "Hash is required", http.StatusBadRequest)
		return
	}

	// Проверка существования hash в хранинилще
	_, exists := h.storage.Get(hash)
	response := VerifyResponse{Valid: exists}

	// Если hash найден - удаляем запись

	if exists {
		h.storage.Delete(hash)
	}

	// Возвращем результат проверки

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Функция для генерации случайного хеша

func generateHash() (string, error) {
	// Создание байтового среза длиной 16 бат (128 бит)
	bytes := make([]byte, 16)
	//Заполнение среза случайными байтами

	if _, err := rand.Read(bytes); err != nil {
		// Если ошибка - возвращаем пустую строку и ошибку
		return "", err
	}

	// Преобразование байтов в hex-стоку и возварт
	return hex.EncodeToString(bytes), nil
}
