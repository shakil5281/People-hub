package models

import (
	"time"

	"gorm.io/gorm"
)

type NightBill struct {
	ID         string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID  string `json:"company_id" gorm:"type:uuid;not null"`
	EmployeeID string `json:"employee_id" gorm:"type:varchar(50);not null"`

	DepartmentID  *string `json:"department_id" gorm:"type:uuid"`
	SectionID     *string `json:"section_id" gorm:"type:uuid"`
	DesignationID *string `json:"designation_id" gorm:"type:uuid"`
	LineID        *string `json:"line_id" gorm:"type:uuid"`
	GroupID       *string `json:"group_id" gorm:"type:uuid"`
	FloorID       *string `json:"floor_id" gorm:"type:uuid"`

	Date        string  `json:"date" gorm:"type:date;not null"`
	NightHours  float64 `json:"night_hours" gorm:"type:decimal(6,2);default:0"`
	Rate        float64 `json:"rate" gorm:"type:decimal(12,2);default:0"`
	Amount      float64 `json:"amount" gorm:"type:decimal(12,2);default:0"`
	Month       int     `json:"month" gorm:"type:int;not null"`
	Year        int     `json:"year" gorm:"type:int;not null"`
	Status      string  `json:"status" gorm:"type:varchar(20);default:pending"`
	Remarks     string  `json:"remarks" gorm:"type:text"`
	IsProcessed bool    `json:"is_processed" gorm:"default:false"`
	ProcessID   *string `json:"process_id" gorm:"type:uuid"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy string         `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *string        `json:"updated_by" gorm:"type:uuid"`

	Employee    Employee     `json:"employee" gorm:"foreignKey:EmployeeID;references:EmployeeID"`
	Company     Company      `json:"company" gorm:"foreignKey:CompanyID"`
	Department  *Department  `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	Section     *Section     `json:"section,omitempty" gorm:"foreignKey:SectionID"`
	Designation *Designation `json:"designation,omitempty" gorm:"foreignKey:DesignationID"`
	Line        *Line        `json:"line,omitempty" gorm:"foreignKey:LineID"`
	Group       *Group       `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	Floor       *Floor       `json:"floor,omitempty" gorm:"foreignKey:FloorID"`
}

type NightBillProcess struct {
	ID        string `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID string `json:"company_id" gorm:"type:uuid;not null"`

	Month          int     `json:"month" gorm:"type:int;not null"`
	Year           int     `json:"year" gorm:"type:int;not null"`
	TotalEmployees int     `json:"total_employees" gorm:"default:0"`
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(14,2);default:0"`
	Status         string  `json:"status" gorm:"type:varchar(20);default:completed"`
	Remarks        string  `json:"remarks" gorm:"type:text"`

	ProcessedAt time.Time      `json:"processed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	ProcessedBy string         `json:"processed_by" gorm:"type:uuid"`

	Company Company `json:"company" gorm:"foreignKey:CompanyID"`
}
