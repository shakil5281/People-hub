package models

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Title     string         `json:"title" gorm:"type:varchar(255);not null"`
	Message   string         `json:"message" gorm:"type:text;not null"`
	Type      string         `json:"type" gorm:"type:varchar(20);not null;default:info"`
	IsRead    bool           `json:"is_read" gorm:"default:false;index"`
	ReadAt    *time.Time     `json:"read_at"`
	Metadata  *string        `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *string        `json:"updated_by" gorm:"type:uuid"`

	User *User `json:"user" gorm:"foreignKey:UserID"`
}
