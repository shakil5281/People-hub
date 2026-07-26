package models

import (
	"time"

	"gorm.io/gorm"
)

type IdCard struct {
	ID            string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeRefID *string        `json:"employee_ref_id" gorm:"type:uuid"`
	Employee      string         `json:"employee" gorm:"type:varchar(255);not null"`
	EmployeeID    string         `json:"employee_id" gorm:"type:varchar(50)"`
	DesignationID string         `json:"designation_id" gorm:"type:uuid"`
	DepartmentID  string         `json:"department_id" gorm:"type:uuid;not null"`
	CardNo        string         `json:"card_no" gorm:"type:varchar(100);not null"`
	Issued        string         `json:"issued" gorm:"type:varchar(10)"`
	Expiry        string         `json:"expiry" gorm:"type:varchar(10)"`
	Status        string         `json:"status" gorm:"type:varchar(20);default:Active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy     *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy     *string        `json:"updated_by" gorm:"type:uuid"`

	Department  Department  `json:"department" gorm:"foreignKey:DepartmentID"`
	Designation Designation `json:"designation" gorm:"foreignKey:DesignationID"`
}
