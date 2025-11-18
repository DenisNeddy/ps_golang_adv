package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// RandomNumberRespose структура для JSON ответа

type RandomNumberResponse struct {
	Number int `json:"number"`
}

func randomNumberHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем только GET ЗАПРОСЫ

	if r.Method != http.MethodGet {
		http.Error(w, "Method not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Создаем новый генератор с случайным seed
	source := rand.NewSource(time.Now().UnixNano())

	generator := rand.New(source)

	// Генерируем случайное число от 1 до 6
	randomNum := generator.Intn(6) + 1

	// Создаем ответ

	response := RandomNumberResponse{
		Number: randomNum,
	}

	// Устанавливаем заголовок для JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Кодируем отправляем JSON ответ
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

}
func main() {

	// Создаем кастомный мультиплексор
	mux := http.NewServeMux()

	// Регистрируем обработчик
	mux.HandleFunc("/random", randomNumberHandler)

	// Настраиваем сервер

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Запускаем сервер

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("GET /random to get a random number from 1 to 6")

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
