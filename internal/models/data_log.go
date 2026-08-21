package models

import (
	"time"

	"gorm.io/gorm"
)

type DataLog struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       int            `json:"user_id" gorm:"type:int;not null;index"`
	BadgeNumber  string         `json:"badge_number" gorm:"type:varchar(50)"`
	EmployeeName string         `json:"employee_name" gorm:"type:varchar(255)"`
	PunchTime    time.Time      `json:"punch_time" gorm:"type:timestamp;not null;index"`
	PunchType    string         `json:"punch_type" gorm:"type:varchar(1)"`
	DeviceID     int            `json:"device_id" gorm:"type:int;default:0"`
	DeviceSN     string         `json:"device_sn" gorm:"type:varchar(50)"`
	Date         string         `json:"date" gorm:"type:date;not null;index"`
	Processed    bool           `json:"processed" gorm:"type:boolean;default:false"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
