package models

import (
	"time"

	"gorm.io/gorm"
)

type EmployeeMigration struct {
	ID               string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID       string         `json:"employee_id" gorm:"type:varchar(50);not null;index"`
	EmployeeName     string         `json:"employee_name" gorm:"type:varchar(255);not null"`
	CompanyID        string         `json:"company_id" gorm:"type:uuid;index"`
	MigrationType    string         `json:"migration_type" gorm:"type:varchar(50);not null"`
	OldDepartmentID  *string        `json:"old_department_id" gorm:"type:uuid"`
	NewDepartmentID  *string        `json:"new_department_id" gorm:"type:uuid"`
	OldSectionID     *string        `json:"old_section_id" gorm:"type:uuid"`
	NewSectionID     *string        `json:"new_section_id" gorm:"type:uuid"`
	OldDesignationID *string        `json:"old_designation_id" gorm:"type:uuid"`
	NewDesignationID *string        `json:"new_designation_id" gorm:"type:uuid"`
	OldLineID        *string        `json:"old_line_id" gorm:"type:uuid"`
	NewLineID        *string        `json:"new_line_id" gorm:"type:uuid"`
	EffectiveDate    string         `json:"effective_date" gorm:"type:varchar(10);not null"`
	Status           string         `json:"status" gorm:"type:varchar(20);default:Completed"`
	Reason           string         `json:"reason" gorm:"type:text"`
	Remarks          string         `json:"remarks" gorm:"type:text"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy        *string        `json:"created_by" gorm:"type:uuid"`

	OldDepartment  *Department  `json:"old_department,omitempty" gorm:"foreignKey:OldDepartmentID"`
	NewDepartment  *Department  `json:"new_department,omitempty" gorm:"foreignKey:NewDepartmentID"`
	OldSection     *Section     `json:"old_section,omitempty" gorm:"foreignKey:OldSectionID"`
	NewSection     *Section     `json:"new_section,omitempty" gorm:"foreignKey:NewSectionID"`
	OldDesignation *Designation `json:"old_designation,omitempty" gorm:"foreignKey:OldDesignationID"`
	NewDesignation *Designation `json:"new_designation,omitempty" gorm:"foreignKey:NewDesignationID"`
	OldLine        *Line        `json:"old_line,omitempty" gorm:"foreignKey:OldLineID"`
	NewLine        *Line        `json:"new_line,omitempty" gorm:"foreignKey:NewLineID"`
}
