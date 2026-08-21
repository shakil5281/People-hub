package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordReset struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string         `json:"user_id" gorm:"type:uuid;not null"`
	TokenHash string         `json:"-" gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time     `json:"used_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
