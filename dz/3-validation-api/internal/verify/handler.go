package handler

import (
	"crypto/rand"   // Для генерации криптографически безопасных случайных чисел
	"encoding/hex"  // Для преобразования байтов в hex-строку
	"encoding/json" // Для работы с JSON
	"fmt"           // Для форматирования строк
	"net/http"      // Для работы с HTTP
	"net/smtp"      // Для отправки email через SMTP

	"3-validation-api/config" // Импорт нашего пакета config

	"github.com/jordan-wright/email" // Импорт внешней библиотеки для email
)

// Определение структуры VerifyHandler
type VerifyHandler struct {
	// Поле cfg хранит указатель на конфигурацию
	cfg *config.Config
}

// Конструктор для VerifyHandler
func NewVerifyHandler(cfg *config.Config) *VerifyHandler {
	// Возвращает указатель на новый VerifyHandler с переданной конфигурацией
	return &VerifyHandler{cfg: cfg}
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

	// Объявление анонимной структуры для парсинга JSON запроса
	var req struct {
		// Поле Email для получения email из JSON
		Email string `json:"email"`
	}

	// Парсинг JSON тела запроса в структуру req

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Если ошибка парсинага - возвращаем ошибку 400
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Генерация хеша для verifaication ссылки

	hash, err := generateHash()

	if err != nil {
		// Если ошибка генерации - возвращаем ошибку 500
		http.Error(w, "Error generating hash", http.StatusInternalServerError)
		return
	}

	// Отправка verification email
	if err := h.sendEmail(req.Email, hash); err != nil {
		// Если ошибка отправки - возвращаем ошибку 500
		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	// Кодирование успешного ответа в JSON и отправка клиенту
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Verification email sent", // Сообщение об успехе
		"hash":    hash,                      // Сгенерированный хеш
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

	e.Text = []byte(fmt.Sprintf("Verify your email: http://localhost:8080/verify/%s", hash))

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

	// Кодирование успешного ответа в JSON и отправка клиенту

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified", // Сообщение подтверждении
		"hash":    hash,             // Хеш который был подвержден
	})
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
