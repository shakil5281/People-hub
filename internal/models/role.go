package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   *string        `json:"company_id" gorm:"type:uuid"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null"`
	Description string         `json:"description"`
	IsSystem    bool           `json:"is_system" gorm:"default:false"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy   *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy   *string        `json:"updated_by" gorm:"type:uuid"`

	Users       []User       `json:"users" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}
