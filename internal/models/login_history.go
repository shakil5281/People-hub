package models

import "time"

type LoginHistory struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    *string   `json:"user_id" gorm:"type:uuid"`
	Email     string    `json:"email" gorm:"type:varchar(255);not null"`
	Status    string    `json:"status" gorm:"type:varchar(20);not null"`
	IPAddress string    `json:"ip_address" gorm:"type:inet"`
	UserAgent string    `json:"user_agent"`
	Browser   string    `json:"browser" gorm:"type:varchar(100)"`
	OS        string    `json:"os" gorm:"type:varchar(100)"`
	CreatedAt time.Time `json:"created_at"`

	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
