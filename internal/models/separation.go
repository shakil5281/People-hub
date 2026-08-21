package models

import (
	"time"

	"gorm.io/gorm"
)

type Separation struct {
	ID            string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeRefID *string        `json:"employee_ref_id" gorm:"type:uuid"`
	Employee      string         `json:"employee" gorm:"type:varchar(255);not null"`
	EmployeeID    string         `json:"employee_id" gorm:"type:varchar(50)"`
	CompanyID     string         `json:"company_id" gorm:"type:uuid"`
	DepartmentID  string         `json:"department_id" gorm:"type:uuid;not null"`
	Type          string         `json:"type" gorm:"type:varchar(50);not null"`
	Date          string         `json:"date" gorm:"type:varchar(10)"`
	Status        string         `json:"status" gorm:"type:varchar(20);default:Pending"`
	Reason        string         `json:"reason" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy     *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy     *string        `json:"updated_by" gorm:"type:uuid"`

	Department Department `json:"department" gorm:"foreignKey:DepartmentID"`
}
