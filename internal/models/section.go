package models

import (
	"time"

	"gorm.io/gorm"
)

type Section struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string         `json:"name" gorm:"type:varchar(255);not null"`
	NameBn       string         `json:"name_bn" gorm:"type:varchar(255);default:''"`
	DepartmentID string         `json:"department_id" gorm:"type:uuid;not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Department   Department     `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	Designations []Designation  `json:"designations,omitempty" gorm:"foreignKey:SectionID"`
	Lines        []Line         `json:"lines,omitempty" gorm:"foreignKey:SectionID"`
}
