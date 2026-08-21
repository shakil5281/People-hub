package models

import "time"

type AuditLog struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     *string   `json:"user_id" gorm:"type:uuid"`
	CompanyID  *string   `json:"company_id" gorm:"type:uuid"`
	Action     string    `json:"action" gorm:"type:varchar(100);not null"`
	Resource   string    `json:"resource" gorm:"type:varchar(100);not null"`
	ResourceID string    `json:"resource_id" gorm:"type:varchar(255)"`
	OldValue   []byte    `json:"old_value" gorm:"type:jsonb"`
	NewValue   []byte    `json:"new_value" gorm:"type:jsonb"`
	IPAddress  string    `json:"ip_address" gorm:"type:inet"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`

	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Company *Company `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}
