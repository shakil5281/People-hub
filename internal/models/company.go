package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Company struct {
	ID            string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyNameBn string         `json:"company_name_bn" gorm:"type:varchar(255);not null"`
	CompanyNameEn string         `json:"company_name_en" gorm:"type:varchar(255);not null"`
	Slug          string         `json:"slug" gorm:"type:varchar(255);uniqueIndex;not null"`
	AddressBn     string         `json:"address_bn" gorm:"type:text"`
	AddressEn     string         `json:"address_en" gorm:"type:text"`
	Phone         string         `json:"phone" gorm:"type:varchar(20)"`
	Email         string         `json:"email" gorm:"type:varchar(255)"`
	Signature     string         `json:"signature" gorm:"type:text"`
	OwnerID       *string        `json:"owner_id" gorm:"type:uuid"`
	Status        string         `json:"status" gorm:"type:varchar(20);default:active"`
	Settings      datatypes.JSON `json:"settings" gorm:"type:jsonb"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy     *string        `json:"created_by" gorm:"type:uuid"`
	UpdatedBy     *string        `json:"updated_by" gorm:"type:uuid"`
	DeletedBy     *string        `json:"deleted_by" gorm:"type:uuid"`
}
