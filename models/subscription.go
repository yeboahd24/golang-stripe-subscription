// models/subscription.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// type Subscription struct {
// 	ProductID    uuid.UUID `gorm:"type:uuid;primary_key;"`
// 	UserID       uuid.UUID `gorm:"references:ID"`
// 	StartDate    time.Time
// 	EndDate      time.Time
// 	TrialEndDate time.Time
// 	Status       string // e.g., "active", "cancelled", "trial"
// 	Plan         string // "monthly" or "yearly"
// 	StripeID     string `json:"stripe_id"` // Add this line

// }

type Subscription struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;"`
	User         CustomUser `gorm:"foreignKey:UserID"`
	ProductID    uuid.UUID  `gorm:"type:uuid;"`
	StartDate    time.Time
	EndDate      time.Time
	TrialEndDate time.Time
	Status       string // e.g., "active", "cancelled", "trial"
	Plan         string // "monthly" or "yearly"
	StripeID     string `json:"stripe_id"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsInTrial    bool `json:"is_in_trial"` // New field to track trial status

}

func (sub *Subscription) BeforeCreate(tx *gorm.DB) error {
	sub.ID = uuid.New()
	return nil
}
