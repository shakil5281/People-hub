package models

import (
	"time"

	"gorm.io/gorm"
)

type PostOffice struct {
	ID         string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name       string         `json:"name" gorm:"type:varchar(255);not null"`
	NameBn     string         `json:"name_bn" gorm:"type:varchar(255);default:''"`
	PostalCode string         `json:"postal_code" gorm:"type:varchar(20);not null"`
	DistrictID string         `json:"district_id" gorm:"type:uuid;not null"`
	UpazilaID  string         `json:"upazila_id" gorm:"type:uuid"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}
