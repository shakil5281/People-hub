package models

import (
	"time"

	"gorm.io/gorm"
)

type Punishment struct {
	ID         string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  string `json:"company_id" gorm:"type:uuid;not null"`
	EmployeeID string `json:"employee_id" gorm:"type:varchar(50);not null"`

	PunishmentType string  `json:"punishment_type" gorm:"type:varchar(20);not null;default:fixed"`
	Reason         string  `json:"reason" gorm:"type:text"`
	Amount         float64 `json:"amount" gorm:"type:decimal(12,2);default:0"`

	OvertimeLessHours *float64 `json:"overtime_less_hours" gorm:"type:decimal(10,2)"`
	OvertimeRate      *float64 `json:"overtime_rate" gorm:"type:decimal(10,2)"`
	AbsentDays        *int     `json:"absent_days" gorm:"type:int"`
	PerDayRate        *float64 `json:"per_day_rate" gorm:"type:decimal(10,2)"`

	Date    string `json:"date" gorm:"type:date;not null"`
	Status  string `json:"status" gorm:"type:varchar(20);default:active"`
	Remarks string `json:"remarks" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy string         `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *string        `json:"updated_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
