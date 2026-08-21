package models

import (
	"time"

	"gorm.io/gorm"
)

type NightBill struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID      string         `json:"company_id" gorm:"type:uuid;not null;index"`
	AttendanceID   string         `json:"attendance_id" gorm:"type:uuid;index"`
	EmployeeID     string         `json:"employee_id" gorm:"type:varchar(50);not null;index"`
	AttendanceDate string         `json:"attendance_date" gorm:"type:varchar(10);not null;index"`
	ShiftID        *string        `json:"shift_id" gorm:"type:uuid"`
	InTime         *time.Time     `json:"in_time" gorm:"type:timestamp"`
	OutTime        *time.Time     `json:"out_time" gorm:"type:timestamp"`
	BillType       string         `json:"bill_type" gorm:"type:varchar(20);not null;default:fixed"` // fixed | hourly | manual
	ShiftEndTime   *string        `json:"shift_end_time" gorm:"type:varchar(10)"`
	EligibleHours  float64        `json:"eligible_hours" gorm:"type:decimal(6,2);default:0"`
	Rate           float64        `json:"rate" gorm:"type:decimal(12,2);default:0"`
	Amount         float64        `json:"amount" gorm:"type:decimal(12,2);default:0"`
	Remarks        string         `json:"remarks" gorm:"type:varchar(255)"`
	ProcessedAt    *time.Time     `json:"processed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy      *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy      *string        `json:"updated_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
	Shift    *Shift   `json:"shift" gorm:"foreignKey:ShiftID"`
	Attendance Attendance `json:"attendance" gorm:"foreignKey:AttendanceID;references:ID"`
}
