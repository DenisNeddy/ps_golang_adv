package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserID   uint      `gorm:"not null" json:"user_id"`
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Products []Product `gorm:"many2many:order_products;" json:"products,omitempty"`
	Status   string    `gorm:"default:'pending'" json:"status"` // pending, completed, cancelled
}
