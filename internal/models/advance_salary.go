package models

import (
	"time"
	"gorm.io/gorm"
)

type AdvanceSalary struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID      string         `json:"company_id" gorm:"type:uuid;not null"`
	EmployeeID     string         `json:"employee_id" gorm:"type:varchar(50);not null"`
	
	Amount         float64        `json:"amount" gorm:"type:decimal(12,2);not null"`
	AdvanceDate    string         `json:"advance_date" gorm:"type:date;not null"`
	DeductionMonth int            `json:"deduction_month" gorm:"not null"`
	DeductionYear  int            `json:"deduction_year" gorm:"not null"`
	Reason         string         `json:"reason" gorm:"type:text"`
	SalaryType     string         `json:"salary_type" gorm:"type:varchar(50)"`
	IsOvertime     bool           `json:"is_overtime" gorm:"default:false"`
	Status         string         `json:"status" gorm:"type:varchar(20);default:pending"` // pending, approved, rejected, deducted
	
	CreatedBy      string         `json:"created_by" gorm:"type:uuid"`
	UpdatedBy      *string        `json:"updated_by" gorm:"type:uuid"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Employee       Employee       `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company        Company        `json:"company" gorm:"foreignKey:CompanyID"`
}
