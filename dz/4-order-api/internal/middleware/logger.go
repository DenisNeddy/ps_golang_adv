package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger структура для middleware логирования
type Logger struct {
	logger *logrus.Logger
}

// NewLogger создаёт новый Logger middleware
func NewLogger() *Logger {
	// Создаём новый логгер
	logger := logrus.New()

	// Устанавливаем формат вывода в JSON
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Устанавливаем уровень логирования (Info - стандартный уровень)
	logger.SetLevel(logrus.InfoLevel)

	return &Logger{
		logger: logger,
	}
}

// Middleware метод для логирования HTTP запросов
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Запоминаем время начала обработки запроса
		start := time.Now()

		// Логируем информацию о входящем запросе
		l.logger.WithFields(logrus.Fields{
			"method": r.Method,     // HTTP метод (GET, POST, PUT, DELETE)
			"path":   r.URL.Path,   // Путь запроса (/products, /health)
			"remote": r.RemoteAddr, // IP адрес клиента
		}).Info("Incoming request") // Сообщение о входящем запросе

		// Вызываем следующий обработчик в цепочке
		next.ServeHTTP(w, r)

		// Вычисляем время обработки запроса
		duration := time.Since(start)

		// Логируем информацию о завершении обработки
		l.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": duration.String(), // Время обработки в читаемом формате
		}).Info("Request completed") // Сообщение о завершении
	})
}
