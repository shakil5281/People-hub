package models

import (
	"time"

	"gorm.io/gorm"
)

type Requirement struct {
	ID            string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Position      string         `json:"position" gorm:"type:varchar(255);not null"`
	DepartmentID  string         `json:"department_id" gorm:"type:uuid"`
	SectionID     string         `json:"section_id" gorm:"type:uuid"`
	DesignationID string         `json:"designation_id" gorm:"type:uuid"`
	GroupType     string         `json:"group_type" gorm:"type:varchar(20);default:Worker"`
	Vacancies     int            `json:"vacancies" gorm:"not null"`
	Applicants    int            `json:"applicants" gorm:"default:0"`
	Status        string         `json:"status" gorm:"type:varchar(20);default:Open"`
	Priority      string         `json:"priority" gorm:"type:varchar(20);default:Medium"`
	Description   string         `json:"description" gorm:"type:text"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy     *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy     *string        `json:"updated_by" gorm:"type:uuid"`

	Department  Department  `json:"department" gorm:"foreignKey:DepartmentID"`
	Section     Section     `json:"section" gorm:"foreignKey:SectionID"`
	Designation Designation `json:"designation" gorm:"foreignKey:DesignationID"`
}
