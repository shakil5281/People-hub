package models

import "time"

type RefreshToken struct {
	ID         string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     string     `json:"user_id" gorm:"type:uuid;not null"`
	TokenHash  string     `json:"-" gorm:"type:varchar(255);uniqueIndex;not null"`
	DeviceInfo string     `json:"device_info"`
	IPAddress  string     `json:"ip_address" gorm:"type:inet"`
	ExpiresAt  time.Time  `json:"expires_at" gorm:"not null"`
	RevokedAt  *time.Time `json:"revoked_at"`
	CreatedAt  time.Time  `json:"created_at"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}
