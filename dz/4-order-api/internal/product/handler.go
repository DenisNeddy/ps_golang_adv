package product

import (
	"4-order-api/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// CreateProduct создает новый продукт

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	// Декодируем JSON из тела запроса
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Создаем модель для БД

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Images:      pq.StringArray(req.Images),
	}

	// Сохраняем в БД

	err = h.db.Create(&product).Error
	if err != nil {
		http.Error(w, "Error creating product", http.StatusInternalServerError)
		return
	}

	// Формируем ответ

	response := ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Images:      product.Images,
		CreatedAt:   product.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetProduct получает продукт по ID

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из URL

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Ищем продукты в БД
	var product models.Product

	err = h.db.First(&product, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
	}

	// Фомируем ответ

	response := ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Images:      product.Images,
		CreatedAt:   product.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Поиск по id

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)

	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Проверяем существование продукта

	var product models.Product

	err = h.db.First(&product, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Error fetching product", http.StatusInternalServerError)
		return
	}

	// Декодируем новые данные

	var req UpdateProductRequest

	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Images = pq.StringArray(req.Images)

	// Сохраняем изменения

	err = h.db.Save(&product).Error
	if err != nil {
		http.Error(w, "Error updating product", http.StatusInternalServerError)
		return
	}

	// Формируем ответ

	response := ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CreatedAt:   product.CreatedAt.Format("2006-01-02 15:0:05"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02 15:0:05"),
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)

}

// DeleteProduct удаляет продукт(soft delete)

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Получаем id из url

	IdStr := r.PathValue("id")
	id, err := strconv.ParseUint(IdStr, 10, 32)

	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Проверяем существование продукта
	var product models.Product

	err = h.db.First(&product, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching product", http.StatusInternalServerError)
		return
	}

	// Sodr delete (благодаря gorm.Model)

	err = h.db.Delete(&product).Error

	if err != nil {
		http.Error(w, "Error deleting product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
