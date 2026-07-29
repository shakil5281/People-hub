package models

import (
	"time"

	"gorm.io/gorm"
)

type TemporaryShift struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID string         `json:"employee_id" gorm:"type:varchar(50);not null;uniqueIndex:idx_temp_shift_emp_date"`
	ShiftID    string         `json:"shift_id" gorm:"type:uuid;not null"`
	CompanyID  string         `json:"company_id" gorm:"type:uuid;not null"`
	Date       string         `json:"date" gorm:"type:date;not null;uniqueIndex:idx_temp_shift_emp_date"`
	Reason     string         `json:"reason" gorm:"type:text"`
	Status     string         `json:"status" gorm:"type:varchar(20);default:active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy  *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy  *string        `json:"updated_by" gorm:"type:uuid"`

	Employee Employee `json:"employee,omitempty" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Shift    Shift    `json:"shift,omitempty" gorm:"foreignKey:ShiftID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
