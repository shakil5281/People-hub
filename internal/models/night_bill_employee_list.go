package models

import (
	"time"

	"gorm.io/gorm"
)

// NightBillEmployeeList is a configuration table that stores
// which employees are eligible for night bills and their bill type/rate settings.
type NightBillEmployeeList struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID   string         `json:"company_id" gorm:"type:uuid;not null;index"`
	EmployeeID  string         `json:"employee_id" gorm:"type:varchar(50);not null"`
	BillType    string         `json:"bill_type" gorm:"type:varchar(20);not null;default:fixed"` // fixed | hourly
	FixedAmount float64        `json:"fixed_amount" gorm:"type:decimal(12,2);default:0"`         // used when bill_type = fixed
	HourlyRate  float64        `json:"hourly_rate" gorm:"type:decimal(12,2);default:0"`          // used when bill_type = hourly
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	Remarks     string         `json:"remarks" gorm:"type:varchar(255)"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy   *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy   *string        `json:"updated_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
