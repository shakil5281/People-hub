package models

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID *string        `json:"company_id" gorm:"type:uuid"`
	Name      string         `json:"name" gorm:"type:varchar(255);not null"`
	NameBn    string         `json:"name_bn" gorm:"type:varchar(255);default:''"`
	HeadID    *string        `json:"head_id" gorm:"type:uuid"`
	Status    string         `json:"status" gorm:"type:varchar(20);default:active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy *string        `json:"updated_by" gorm:"type:uuid"`

	Company   Company    `json:"company" gorm:"foreignKey:CompanyID"`
	Employees []Employee `json:"employees" gorm:"foreignKey:DepartmentID"`
	Sections  []Section  `json:"sections,omitempty" gorm:"foreignKey:DepartmentID"`
}
