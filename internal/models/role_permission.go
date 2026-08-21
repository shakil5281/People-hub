package models

import "time"

type RolePermission struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoleID       string    `json:"role_id" gorm:"type:uuid;not null"`
	PermissionID string    `json:"permission_id" gorm:"type:uuid;not null"`
	CreatedAt    time.Time `json:"created_at"`

	Role       Role       `json:"role" gorm:"foreignKey:RoleID"`
	Permission Permission `json:"permission" gorm:"foreignKey:PermissionID"`
}
