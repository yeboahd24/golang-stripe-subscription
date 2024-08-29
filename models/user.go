// models/user.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// type CustomUser struct {
// 	ID            uuid.UUID `gorm:"type:uuid;primary_key;"`
// 	Email         string    `gorm:"unique;not null"`
// 	PasswordHash  string    `gorm:"not null"`
// 	Subscriptions []Subscription
// 	IsAdmin       bool
// 	CreatedAt     time.Time
// 	UpdatedAt     time.Time
// }

type CustomUser struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string    `gorm:"unique;not null"`
	Password      string
	Subscriptions []Subscription `gorm:"foreignKey:UserID"`
	IsAdmin       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (user *CustomUser) BeforeCreate(tx *gorm.DB) error {
	user.ID = uuid.New()
	return nil
}
