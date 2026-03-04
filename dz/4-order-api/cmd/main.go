package main

import (
	"4-order-api/config"
	"4-order-api/database"
	"4-order-api/internal/auth"
	"4-order-api/internal/middleware"
	"4-order-api/internal/product"
	"log"
	"net/http"
)

func main() {

	// 1. Загрузка конйигурации

	cfg := config.NewConfig()

	// 2. Подключение к базе данных

	db, err := database.Connect(cfg)
	authHandler := auth.NewHandler(db, cfg)

	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	// 3.Выполнение миграций

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	loggerMiddleware := middleware.NewLogger()

	// 4. Создаем handler для продуктов

	productHandler := product.NewHandler(db)

	// 5. Настройка HTTP сервера (пока пустой)

	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret)

	// Публичные эндпоинты (без авторизации)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/send", authHandler.SendSMS)
	mux.HandleFunc("POST /auth/verify", authHandler.VerifyCode)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	protectedMux := http.NewServeMux()

	protectedMux.HandleFunc("POST /products", productHandler.CreateProduct)
	protectedMux.HandleFunc("GET /products/{id}", productHandler.GetProduct)
	protectedMux.HandleFunc("PUT /products/{id}", productHandler.UpdateProduct)
	protectedMux.HandleFunc("DELETE /products/{id}", productHandler.DeleteProduct)

	protectedMux.Handle("/products/", jwtMiddleware.Authenticate(mux))

	log.Printf("Server starting on %s", cfg.ServerPort)
	handler := loggerMiddleware.Middleware(mux)
	if err := http.ListenAndServe(cfg.ServerPort, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
