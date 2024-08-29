// handlers/product_handler.go
package handlers

import (
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/yeboahd24/subscription-stripe/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/price"
	"github.com/stripe/stripe-go/v79/product"
)

func GetProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product
		if err := db.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		c.JSON(http.StatusOK, products)
	}
}

func createStripeProduct(name string, description string, monthlyPrice float64, yearlyPrice float64) (*models.Product, error) {
	// Create the product in Stripe
	params := &stripe.ProductParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
	}
	stripeProduct, err := product.New(params)
	if err != nil {
		return nil, err
	}

	// Create monthly price
	monthlyPriceParams := &stripe.PriceParams{
		Product:    stripe.String(stripeProduct.ID),
		UnitAmount: stripe.Int64(int64(monthlyPrice * 100)), // Stripe uses cents
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalMonth)),
		},
	}
	monthlyStripePrice, err := price.New(monthlyPriceParams)
	if err != nil {
		return nil, err
	}

	// Create yearly price
	yearlyPriceParams := &stripe.PriceParams{
		Product:    stripe.String(stripeProduct.ID),
		UnitAmount: stripe.Int64(int64(yearlyPrice * 100)), // Stripe uses cents
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String(string(stripe.PriceRecurringIntervalYear)),
		},
	}
	yearlyStripePrice, err := price.New(yearlyPriceParams)
	if err != nil {
		return nil, err
	}

	// Create the product in your database
	product := &models.Product{
		ID:                   uuid.New(),
		Name:                 name,
		Description:          description,
		MonthlyPrice:         monthlyPrice,
		YearlyPrice:          yearlyPrice,
		StripeMonthlyPriceID: monthlyStripePrice.ID,
		StripeYearlyPriceID:  yearlyStripePrice.ID,
	}

	return product, nil
}

func CreateProductHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

		userID, exists := c.Get("user_id") // Change to retrieve user ID
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Fetch the user from the database using userID
		var user models.CustomUser
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		var input struct {
			Name         string  `json:"name" binding:"required"`
			Description  string  `json:"description" binding:"required"`
			MonthlyPrice float64 `json:"monthly_price" binding:"required"`
			YearlyPrice  float64 `json:"yearly_price" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		product, err := createStripeProduct(input.Name, input.Description, input.MonthlyPrice, input.YearlyPrice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		// Save product to database (assuming a SaveProduct function exists)
		if err := db.Create(product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save product", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}
