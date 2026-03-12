package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Phone  string  `gorm:"uniqueIndex;not null" json:"phone"`
	Orders []Order `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}
