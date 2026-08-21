package models

import (
	"time"

	"gorm.io/gorm"
)

type EmailVerification struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    *string        `json:"user_id" gorm:"type:uuid"`
	Email     string         `json:"email" gorm:"type:varchar(255);not null"`
	TokenHash string         `json:"-" gorm:"type:varchar(255);uniqueIndex;not null"`
	Purpose   string         `json:"purpose" gorm:"type:varchar(50);not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time     `json:"used_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
