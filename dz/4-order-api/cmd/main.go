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

	// 1. Загрузка конфигурации
	cfg := config.NewConfig()

	// 2. Подключение к базе данных
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	// 3. Выполнение миграций
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	// 4. Создаём handlers
	authHandler := auth.NewHandler(db, cfg)
	productHandler := product.NewHandler(db)

	// 5. Создаём middleware
	loggerMiddleware := middleware.NewLogger()
	jwtMiddleware := middleware.NewJWTMiddleware(cfg.JWTSecret)

	// 6. Создаём ОДИН мультиплексор для всех маршрутов
	mux := http.NewServeMux()

	// 7. Регистрируем публичные эндпоинты (без авторизации)
	mux.HandleFunc("POST /auth/send", authHandler.SendSMS)
	mux.HandleFunc("POST /auth/verify", authHandler.VerifyCode)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 8. Регистрируем защищённые эндпоинты продуктов через middleware
	mux.Handle("POST /products", jwtMiddleware.Authenticate(http.HandlerFunc(productHandler.CreateProduct)))
	mux.Handle("GET /products/{id}", jwtMiddleware.Authenticate(http.HandlerFunc(productHandler.GetProduct)))
	mux.Handle("PUT /products/{id}", jwtMiddleware.Authenticate(http.HandlerFunc(productHandler.UpdateProduct)))
	mux.Handle("DELETE /products/{id}", jwtMiddleware.Authenticate(http.HandlerFunc(productHandler.DeleteProduct)))

	// 9. Запускаем сервер с middleware логирования
	log.Printf("Server starting on %s", cfg.ServerPort)
	handler := loggerMiddleware.Middleware(mux)
	if err := http.ListenAndServe(cfg.ServerPort, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
