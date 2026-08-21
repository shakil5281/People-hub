package models

import (
	"time"

	"gorm.io/gorm"
)

type Holiday struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   string         `json:"company_id" gorm:"type:uuid;not null"`
	Name        string         `json:"name" gorm:"type:varchar(255);not null"`
	Date        string         `json:"date" gorm:"type:date;not null"`
	FromDate    *string        `json:"from_date" gorm:"type:date"`
	ToDate      *string        `json:"to_date" gorm:"type:date"`
	WeekendDate *string        `json:"weekend_date" gorm:"type:date"`
	Type        string         `json:"type" gorm:"type:varchar(50);not null;default:government"`
	Description string         `json:"description" gorm:"type:text"`
	Status      string         `json:"status" gorm:"type:varchar(20);default:active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy   *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy   *string        `json:"updated_by" gorm:"type:uuid"`

	Company Company `json:"company" gorm:"foreignKey:CompanyID"`
}
