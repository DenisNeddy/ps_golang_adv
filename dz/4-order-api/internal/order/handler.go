package order

import (
	"4-order-api/internal/middleware"
	"4-order-api/models"
	"encoding/json"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// POST /order - создание заказа

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {

	// 1. Получаем user_id из JWT токена (middleware уже проверил токен)
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// 2. Декодируем JSON с массивом product_ids

	var req CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 3. Валидация: проверяем что массив не пустой

	if len(req.ProductIDs) == 0 {
		http.Error(w, "Product IDs are required", http.StatusBadRequest)
		return
	}

	// 4. Проверяем существование всех продуктов

	var products []models.Product

	result := h.db.Where("id IN ?", req.ProductIDs).Find(&products)

	if result.Error != nil {
		http.Error(w, "Error fetching products", http.StatusInternalServerError)
		return
	}

	// 5. Проверяем что все продукты найдены
	if len(products) != len(req.ProductIDs) {
		http.Error(w, "Some products not found", http.StatusNotFound)
		return
	}

	// 6. Создаём заказ
	order := models.Order{
		UserID:   userID,
		Products: products,
		Status:   "pending",
	}

	if err := h.db.Create(&order).Error; err != nil {
		http.Error(w, "Error creating order", http.StatusInternalServerError)
		return
	}

	// 7. Загружаем связанные данные для ответа

	h.db.Preload("Products").First(&order, order.ID)

	// 8. Формируем ответ

	response := h.buildOrderResponse(order)

	w.Header().Set("Contetn-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GET /order/{id} - получение заказа по ID

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем user_id из JWT

	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// 2. Получаем ID заказа из URL
	idStr := r.PathValue("id")
	orderID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// 3. Ищем заказ с провекой принадлежности пользователя
	var order models.Order

	result := h.db.Preload("Products").Where("id = ? AND user_id = ?", orderID, userID).First(&order)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Error fetching order", http.StatusInternalServerError)
		return
	}

	// 4. Формирум ответ

	response := h.buildOrderResponse(order)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//GET /my-orders - получение всех заказов пользователя

func (h *Handler) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем user_id из JWT

	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)

	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// 2. Получаем все заказы пользователя

	var orders []models.Order

	result := h.db.Preload("Products").Where("user_id = ?", userID).Find(&orders)

	if result.Error != nil {
		http.Error(w, "Error fetching orders", http.StatusInternalServerError)
		return
	}

	// 3. Формируем массиво ответов

	responses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = h.buildOrderResponse(order)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// Вспомогательная функция для формирования ответа

func (h *Handler) buildOrderResponse(order models.Order) OrderResponse {
	products := make([]ProductInOrder, len(order.Products))

	for i, p := range order.Products {
		products[i] = ProductInOrder{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}

	return OrderResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		Status:    order.Status,
		Products:  products,
		CreatedAt: order.CreatedAt.Format("2006-01-02 15:04:06"),
		UpdatedAt: order.UpdatedAt.Format("2006-01-02 15:04:06"),
	}

}
