package models

import (
	"time"

	"gorm.io/gorm"
)

// MissingAttendance stores manually entered attendance overrides.
// These records have the HIGHEST priority during daily processing:
// if a missing_attendance record exists for an employee+date, it is
// applied BEFORE any ZKTeco log data, and log processing is skipped
// for that employee on that date.
type MissingAttendance struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID string         `json:"employee_id" gorm:"type:varchar(50);not null;index:idx_missing_att_emp_date"`
	CompanyID  string         `json:"company_id" gorm:"type:uuid;not null"`
	Date       string         `json:"date" gorm:"type:date;not null;index:idx_missing_att_emp_date"`
	CheckIn    *time.Time     `json:"check_in" gorm:"type:timestamp"`
	CheckOut   *time.Time     `json:"check_out" gorm:"type:timestamp"`
	Status     string         `json:"status" gorm:"type:varchar(20);not null;default:present"`
	Notes      string         `json:"notes" gorm:"type:text"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy  *string        `json:"created_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
