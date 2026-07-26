package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                  string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email               string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash        string         `json:"-" gorm:"type:varchar(255);not null"`
	Name                string         `json:"name" gorm:"type:varchar(255);not null"`
	Status              string         `json:"status" gorm:"type:varchar(20);default:pending"`
	MFAEnabled          bool           `json:"mfa_enabled" gorm:"default:false"`
	MFASecret           string         `json:"-" gorm:"type:varchar(255)"`
	FailedAttempts      int            `json:"failed_attempts" gorm:"default:0"`
	LockedAt            *time.Time     `json:"locked_at"`
	LockedBy            *string        `json:"locked_by" gorm:"type:uuid"`
	LastLoginAt         *time.Time     `json:"last_login_at"`
	LastLoginIP         *string        `json:"last_login_ip" gorm:"type:inet"`
	ForcePasswordChange bool           `json:"force_password_change" gorm:"default:false"`
	EmailVerifiedAt     *time.Time     `json:"email_verified_at"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy           *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy           *string        `json:"updated_by" gorm:"type:uuid"`
	DeletedBy           *string        `json:"deleted_by" gorm:"type:uuid"`

	Roles       []Role       `json:"roles" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:user_roles;join_foreignKey:UserID;joinReferences:RoleID;"`
}
