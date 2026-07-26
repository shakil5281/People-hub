package models

import (
	"time"

	"gorm.io/gorm"
)

type DailySchedule struct {
	ID         string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  string `json:"company_id" gorm:"type:uuid;not null"`
	EmployeeID string `json:"employee_id" gorm:"type:varchar(50);not null"`

	Date         string `json:"date" gorm:"type:date;not null"`
	ScheduleType string `json:"schedule_type" gorm:"type:varchar(50);not null"`
	StartTime    string `json:"start_time" gorm:"type:varchar(5)"`
	EndTime      string `json:"end_time" gorm:"type:varchar(5)"`
	Remarks      string `json:"remarks" gorm:"type:text"`
	Status       string `json:"status" gorm:"type:varchar(20);default:active"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy string         `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *string        `json:"updated_by" gorm:"type:uuid"`

	Employee Employee `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company  Company  `json:"company" gorm:"foreignKey:CompanyID"`
}
