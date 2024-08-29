// models/product.go
package models

import (
	"github.com/google/uuid"
)

type Product struct {
	ID                   uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name                 string    `gorm:"not null"`
	Description          string
	MonthlyPrice         float64
	YearlyPrice          float64
	StripeMonthlyPriceID string `gorm:"type:varchar(255)"`
	StripeYearlyPriceID  string `gorm:"type:varchar(255)"`
}
