package order

// Запрос на создание заказа

type CreateOrderRequest struct {
	ProductIDs []uint `json:"product_ids"`
}

// Ответ с данными заказа

type OrderResponse struct {
	ID        uint             `json:"id"`
	UserID    uint             `json:"user_id"`
	Status    string           `json:"status"`
	Products  []ProductInOrder `json:"products"`
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
}

// Упрощенная информация продукте в заказе

type ProductInOrder struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
