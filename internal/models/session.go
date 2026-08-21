package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string         `json:"user_id" gorm:"type:uuid;not null"`
	IPAddress  string         `json:"ip_address" gorm:"type:inet"`
	UserAgent  string         `json:"user_agent"`
	Browser    string         `json:"browser" gorm:"type:varchar(100)"`
	OS         string         `json:"os" gorm:"type:varchar(100)"`
	LastActive time.Time      `json:"last_active"`
	ExpiresAt  time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}
