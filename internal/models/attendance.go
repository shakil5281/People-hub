package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Attendance struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID  string         `json:"employee_id" gorm:"type:varchar(50);not null;index:idx_attendance_employee_date"`
	CompanyID   string         `json:"company_id" gorm:"type:uuid;not null"`
	ShiftID     *string        `json:"shift_id" gorm:"type:uuid"`
	Date        string         `json:"date" gorm:"type:date;not null;index:idx_attendance_employee_date"`
	CheckIn     *time.Time     `json:"check_in" gorm:"type:timestamp"`
	CheckOut    *time.Time     `json:"check_out" gorm:"type:timestamp"`
	TotalHours  *string        `json:"total_hours" gorm:"type:varchar(5)"`
	OverTime    *string        `json:"over_time" gorm:"type:varchar(5)"`
	Status      string         `json:"status" gorm:"type:varchar(20);default:present"`
	LateMinutes int            `json:"late_minutes" gorm:"type:int;default:0"`
	PunchNumber *string        `json:"punch_number" gorm:"type:varchar(50)"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy   *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy   *string        `json:"updated_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
	Shift    *Shift   `json:"shift,omitempty" gorm:"foreignKey:ShiftID"`
}

// CalculateHours sets TotalHours from check_in and check_out.
// It does NOT touch OverTime — that is exclusively set by the attendance processor.
func (a *Attendance) CalculateHours() {
	if a.CheckIn == nil || a.CheckOut == nil {
		a.TotalHours = nil
		return
	}
	duration := a.CheckOut.Sub(*a.CheckIn)
	if duration <= 0 {
		a.TotalHours = nil
		return
	}
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	th := fmt.Sprintf("%02d:%02d", hours, minutes)
	a.TotalHours = &th
}

// BeforeSave recalculates TotalHours before every DB write.
// OverTime is intentionally left untouched here; the attendance processor
// sets it via the 45-minute OT rule and employee.overtime_status.
func (a *Attendance) BeforeSave(tx *gorm.DB) error {
	a.CalculateHours()
	return nil
}
