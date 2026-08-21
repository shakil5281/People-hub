package models

import "time"

type Permission struct {
	ID          string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Resource    string    `json:"resource" gorm:"type:varchar(50);not null"`
	Action      string    `json:"action" gorm:"type:varchar(50);not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`

	Roles []Role `json:"roles" gorm:"many2many:role_permissions;"`
}
