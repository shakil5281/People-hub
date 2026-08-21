package models

import (
	"time"

	"gorm.io/gorm"
)

type Upazila struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name       string         `json:"name" gorm:"type:varchar(255);not null"`
	NameBn     string         `json:"name_bn" gorm:"type:varchar(255);default:''"`
	DistrictID string         `json:"district_id" gorm:"type:uuid;not null"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	District   District       `json:"district,omitempty" gorm:"foreignKey:DistrictID"`
	Unions     []Union        `json:"unions,omitempty" gorm:"foreignKey:UpazilaID"`
}
