package models

import "time"

type UserRole struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null"`
	RoleID    string    `json:"role_id" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy *string   `json:"created_by" gorm:"type:uuid"`

	User User `json:"user" gorm:"foreignKey:UserID"`
	Role Role `json:"role" gorm:"foreignKey:RoleID"`
}
