package models

import (
	"time"

	"gorm.io/gorm"
)

type Leave struct {
	ID              string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID       string         `json:"company_id" gorm:"type:uuid;not null"`
	EmployeeID      string         `json:"employee_id" gorm:"type:varchar(50);not null"`
	LeaveTypeID     string         `json:"leave_type_id" gorm:"type:uuid;not null"`
	FromDate        string         `json:"from_date" gorm:"type:date;not null"`
	ToDate          string         `json:"to_date" gorm:"type:date;not null"`
	TotalDays       int            `json:"total_days" gorm:"not null"`
	Reason          string         `json:"reason" gorm:"type:text"`
	Status          string         `json:"status" gorm:"type:varchar(20);default:pending"`
	ApprovedBy      *string        `json:"approved_by" gorm:"type:uuid"`
	ApprovedAt      *time.Time     `json:"approved_at"`
	RejectionReason string         `json:"rejection_reason" gorm:"type:text"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy       *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy       *string        `json:"updated_by" gorm:"type:uuid"`

	Company   Company   `json:"company" gorm:"foreignKey:CompanyID"`
	Employee  Employee  `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	LeaveType LeaveType `json:"leave_type" gorm:"foreignKey:LeaveTypeID"`
}

type LeaveAllocation struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID  string         `json:"employee_id" gorm:"type:varchar(50);not null;uniqueIndex:idx_alloc_emp_leave_year"`
	LeaveTypeID string         `json:"leave_type_id" gorm:"type:uuid;not null;uniqueIndex:idx_alloc_emp_leave_year"`
	Year        int            `json:"year" gorm:"not null;uniqueIndex:idx_alloc_emp_leave_year"`
	TotalDays   int            `json:"total_days" gorm:"not null"`
	UsedDays    int            `json:"used_days" gorm:"default:0"`
	PendingDays int            `json:"pending_days" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Employee  Employee  `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	LeaveType LeaveType `json:"leave_type" gorm:"foreignKey:LeaveTypeID"`
}
