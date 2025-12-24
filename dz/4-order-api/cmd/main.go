package main

import (
	"4-order-api/config"
	"4-order-api/database"
	"log"
	"net/http"
)

func main() {

	// 1. Загрузка конйигурации

	cfg := config.NewConfig()

	// 2. Подключение к базе данных

	db, err := database.Connect(cfg)

	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	// 3.Выполнение миграций

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	// 4. Настройка HTTP сервера (пока пустой)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Server starting on %s", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
