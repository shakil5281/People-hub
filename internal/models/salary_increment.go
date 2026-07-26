package models

import (
	"time"

	"gorm.io/gorm"
)

type SalaryIncrement struct {
	ID         string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  string `json:"company_id" gorm:"type:uuid;not null"`
	EmployeeID string `json:"employee_id" gorm:"type:varchar(50);not null"`

	PreviousGross   float64 `json:"previous_gross" gorm:"type:decimal(12,2);default:0"`
	PreviousBasic   float64 `json:"previous_basic" gorm:"type:decimal(12,2);default:0"`
	PreviousHouse   float64 `json:"previous_house_rent" gorm:"type:decimal(12,2);default:0"`
	PreviousMedical float64 `json:"previous_medical" gorm:"type:decimal(12,2);default:0"`

	IncrementAmount float64 `json:"increment_amount" gorm:"type:decimal(12,2);not null"`
	NewGross        float64 `json:"new_gross" gorm:"type:decimal(12,2);default:0"`
	NewBasic        float64 `json:"new_basic" gorm:"type:decimal(12,2);default:0"`
	NewHouse        float64 `json:"new_house_rent" gorm:"type:decimal(12,2);default:0"`
	NewMedical      float64 `json:"new_medical" gorm:"type:decimal(12,2);default:0"`

	EffectiveDate string `json:"effective_date" gorm:"type:date;not null"`
	Status        string `json:"status" gorm:"type:varchar(20);default:pending"`
	Remarks       string `json:"remarks" gorm:"type:text"`

	ApprovedBy      *string    `json:"approved_by" gorm:"type:uuid"`
	ApprovedAt      *time.Time `json:"approved_at"`
	RejectedBy      *string    `json:"rejected_by" gorm:"type:uuid"`
	RejectedAt      *time.Time `json:"rejected_at"`
	RejectionReason string     `json:"rejection_reason" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy string         `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *string        `json:"updated_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
