package models

import (
	"time"

	"gorm.io/gorm"
)

type EidBonus struct {
	ID         string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  string `json:"company_id" gorm:"type:uuid;not null;uniqueIndex:idx_eid_bonus_emp"`
	EmployeeID string `json:"employee_id" gorm:"type:varchar(50);not null;uniqueIndex:idx_eid_bonus_emp"`
	Year       int    `json:"year" gorm:"not null;uniqueIndex:idx_eid_bonus_emp"`
	BonusType  string `json:"bonus_type" gorm:"type:varchar(20);not null"`

	GrossSalary float64 `json:"gross_salary" gorm:"type:decimal(12,2);default:0"`
	BonusAmount float64 `json:"bonus_amount" gorm:"type:decimal(12,2);default:0"`

	Status    string         `json:"status" gorm:"type:varchar(20);default:processed"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy *string        `json:"created_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
