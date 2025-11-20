package main

import (
	"3-validation-api/config"                  // Наш пакет конфигурации
	handler "3-validation-api/internal/verify" // Наш пакет обработчиков
	"fmt"                                      // Для вывода в консоль
	"net/http"                                 // Для HTTP сервера
)

func main() {
	// Создание конфигурации с жестко закодироваными значениями

	cfg := &config.Config{
		Email:    "your-email@gmail.com", // Email отправителя
		Password: "your-app-password",    // Пароль приложения
		Address:  "smtp.gmail.com:587",   // SMTP адрес и порт
	}

	// Создание обработчика verify с передачей конфигурации

	verifyHandler := handler.NewVerifyHandler(cfg)

	// Создаем мультиплексора маршрутов

	mux := http.NewServeMux()

	// Регистрация обработчика для POST /send
	mux.HandleFunc("POST /send", verifyHandler.Send)
	// Регистрация обработчика для GET /verify/{hash}
	mux.HandleFunc("GET /verify/{hash}", verifyHandler.Verify)

	// Вывод сообщения о запуске сервера
	fmt.Println("Server starting on :8080")
	// Запуск HTTP сервера на порту 8080
	http.ListenAndServe(":8080", mux)

}
