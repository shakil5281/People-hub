package models

import "time"

type SystemLog struct {
	ID         string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Level      string    `json:"level" gorm:"type:varchar(20);not null;index"`
	Source     string    `json:"source" gorm:"type:varchar(50);not null;index"`
	Message    string    `json:"message" gorm:"type:text;not null"`
	StackTrace *string   `json:"stack_trace,omitempty" gorm:"type:text"`
	UserID     *string   `json:"user_id" gorm:"type:uuid;index"`
	CompanyID  *string   `json:"company_id" gorm:"type:uuid;index"`
	IPAddress  *string   `json:"ip_address" gorm:"type:inet"`
	UserAgent  *string   `json:"user_agent"`
	Method     *string   `json:"method" gorm:"type:varchar(10)"`
	Path       *string   `json:"path" gorm:"type:text"`
	StatusCode *int      `json:"status_code"`
	Latency    *int64    `json:"latency"`
	CreatedAt  time.Time `json:"created_at" gorm:"index"`

	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Company *Company `json:"company,omitempty" gorm:"foreignKey:CompanyID"`
}
