package models

import (
	"time"

	"gorm.io/gorm"
)

// OtEarlyExitDeduction records a day where an employee left before the end of
// their shift. The shortfall (expected work hours minus actual worked hours) is
// deducted from that employee's monthly overtime, and this table is the
// immutable ledger of every such deduction.
type OtEarlyExitDeduction struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID     string         `json:"employee_id" gorm:"type:varchar(50);not null;index:idx_ot_early_exit_emp_date"`
	CompanyID      string         `json:"company_id" gorm:"type:uuid;not null"`
	Date           string         `json:"date" gorm:"type:date;not null;index:idx_ot_early_exit_emp_date"`
	ShiftID        *string        `json:"shift_id" gorm:"type:uuid"`
	ShiftStartTime string         `json:"shift_start_time" gorm:"type:varchar(5)"`
	ShiftEndTime   string         `json:"shift_end_time" gorm:"type:varchar(5)"`
	ExpectedHours  float64        `json:"expected_hours" gorm:"type:numeric(6,2);default:0"`
	WorkedHours    float64        `json:"worked_hours" gorm:"type:numeric(6,2);default:0"`
	ShortfallHours float64        `json:"shortfall_hours" gorm:"type:numeric(6,2);default:0"`
	Status         string         `json:"status" gorm:"type:varchar(20);not null;default:present"`
	Month          int            `json:"month" gorm:"type:int;not null"`
	Year           int            `json:"year" gorm:"type:int;not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy      *string        `json:"created_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
