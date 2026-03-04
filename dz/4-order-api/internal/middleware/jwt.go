package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type JWTMiddleware struct {
	secret string
}

func NewJWTMiddleware(secret string) *JWTMiddleware {
	return &JWTMiddleware{secret: secret}
}

// contextKey тип для ключей контекста (избегаем коллизий)
type contextKey string

const (
	// UserIDKey ключ для хранения user_id в контексте
	UserIDKey contextKey = "user_id"
	// PhoneKey ключ для хранения phone в контексте
	PhoneKey contextKey = "phone"
)

func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Извлекаем заголовок Authorization

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondError(w, "Missing Authorization header", http.StatusUnauthorized)
		}

		// 2. Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.respondError(w, "Invalid Authorization header format. Use: Bearer <token>", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// 3. Парсим и валидируем JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.secret), nil
		})

		if err != nil {
			m.respondError(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// 4. Извлекаем claims и сохраняем в контекст
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Извлекаем user_id (может быть float64 из-за JSON)
			userID, ok := claims["user_id"].(float64)
			if !ok {
				m.respondError(w, "Invalid token claims: missing user_id", http.StatusUnauthorized)
				return
			}

			// Извлекаем phone
			phone, ok := claims["phone"].(string)
			if !ok {
				m.respondError(w, "Invalid token claims: missing phone", http.StatusUnauthorized)
				return
			}

			// Сохраняем данные в контекст запроса
			ctx := context.WithValue(r.Context(), UserIDKey, uint(userID))
			ctx = context.WithValue(ctx, PhoneKey, phone)

			// Вызываем следующий обработчик с обновлённым контекстом
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			// 5. Токен невалиден
			m.respondError(w, "Invalid or expired token", http.StatusUnauthorized)
		}

	})
}

func (m *JWTMiddleware) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
