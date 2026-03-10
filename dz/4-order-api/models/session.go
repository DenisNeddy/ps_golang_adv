package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	SessionID string    `gorm:"uniqueIndex;not null" json:"session_id"`
	Phone     string    `gorm:"not null" json:"phone"`
	Code      int       `gorm:"not null" json:"code"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
}
