package models

import "time"

type PasswordHistory struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string    `json:"user_id" gorm:"type:uuid;not null"`
	PasswordHash string    `json:"-" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}
