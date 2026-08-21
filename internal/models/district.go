package models

import (
	"time"

	"gorm.io/gorm"
)

type District struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name       string         `json:"name" gorm:"type:varchar(255);not null"`
	NameBn     string         `json:"name_bn" gorm:"type:varchar(255);default:''"`
	DivisionID string         `json:"division_id" gorm:"type:uuid;not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	Division   Division       `json:"division,omitempty" gorm:"foreignKey:DivisionID"`
	Upazilas   []Upazila      `json:"upazilas,omitempty" gorm:"foreignKey:DistrictID"`
}
