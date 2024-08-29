package handlers

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/yeboahd24/subscription-stripe/models"
	"github.com/yeboahd24/subscription-stripe/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/sub"
	"gorm.io/gorm"
)

func Subscribe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

		userID, _ := c.Get("user_id")

		var subscribeRequest struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Plan      string    `json:"plan" binding:"required,oneof=monthly yearly"` // Removed "trial"
		}

		if err := c.ShouldBindJSON(&subscribeRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var product models.Product
		if err := db.First(&product, subscribeRequest.ProductID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		var user models.CustomUser
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Create Stripe customer
		params := &stripe.CustomerParams{
			Email: stripe.String(user.Email),
		}
		stripeCustomer, err := customer.New(params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe customer"})
			return
		}

		stripePriceID, err := getStripePriceID(db, product, subscribeRequest.Plan)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Stripe Price ID"})
			return
		}

		// Create Stripe subscription
		subParams := &stripe.SubscriptionParams{
			Customer: stripe.String(stripeCustomer.ID),
			Items: []*stripe.SubscriptionItemsParams{
				{
					Price: stripe.String(stripePriceID),
				},
			},
			TrialPeriodDays: stripe.Int64(30),
		}
		stripeSub, err := sub.New(subParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe subscription"})
			return
		}

		// Check if the user already has an active subscription
		var existingSubscription models.Subscription
		if err := db.Where("user_id = ? AND status != ?", user.ID, "cancelled").First(&existingSubscription).Error; err == nil {
			// User already has an active subscription
			c.JSON(http.StatusConflict, gin.H{"error": "User already has an active subscription"})
			return
		}

		var existingUserSubscription models.Subscription
		if err := db.Where("user_id = ? AND (status = ? OR status = ?)", user.ID, "active", "completed").First(&existingUserSubscription).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already has an active or completed subscription"})
			return
		}

		// Create local subscription
		subscription := models.Subscription{
			UserID:       user.ID,
			ProductID:    product.ID,
			StartDate:    time.Now(),
			TrialEndDate: time.Time{}, // Set to zero value since trial is removed
			EndDate: time.Now().AddDate(0, func() int {
				if subscribeRequest.Plan == "monthly" {
					return 1
				} else if subscribeRequest.Plan == "yearly" {
					return 12
				}
				return 0 // No end date for trial
			}(), 0),
			Status:    "active",
			Plan:      subscribeRequest.Plan,
			StripeID:  stripeSub.ID,
			IsInTrial: false, // Set IsInTrial to false since trials are removed
		}

		if err := db.Create(&subscription).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":      "Subscription created successfully",
			"subscription": subscription,
			"stripe_id":    stripeSub.ID,
		})
	}
}

// func GetSubscription(db *gorm.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		userID, _ := c.Get("userID")

// 		var subscription models.Subscription
// 		if err := db.Where("user_id = ?", userID).Last(&subscription).Error; err != nil {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, subscription)
// 	}
// }

func GetSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}

		utils.Log("Fetching subscription for userID:", userID)

		// Update trial status before fetching the subscription
		if err := UpdateTrialStatus(db); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trial status"})
			return
		}

		var subscription models.Subscription
		if err := db.Where("user_id = ? AND status = ?", userID, "active").Last(&subscription).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Active Subscription not found for userID: " + userID.(uuid.UUID).String()})
			return
		}

		response := struct {
			ID           string    `json:"id"`
			UserID       string    `json:"user_id"`
			ProductID    string    `json:"product_id"`
			StartDate    time.Time `json:"start_date"`
			EndDate      time.Time `json:"end_date"`
			TrialEndDate time.Time `json:"trial_end_date"`
			Status       string    `json:"status"`
			Plan         string    `json:"plan"`
			StripeID     string    `json:"stripe_id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			IsInTrial    bool      `json:"is_in_trial"`
		}{
			ID:           subscription.ID.String(),
			UserID:       subscription.UserID.String(),
			ProductID:    subscription.ProductID.String(),
			StartDate:    subscription.StartDate,
			EndDate:      subscription.EndDate,
			TrialEndDate: subscription.TrialEndDate,
			Status:       subscription.Status,
			Plan:         subscription.Plan,
			StripeID:     subscription.StripeID,
			CreatedAt:    subscription.CreatedAt,
			UpdatedAt:    subscription.UpdatedAt,
			IsInTrial:    subscription.IsInTrial,
		}

		c.JSON(http.StatusOK, response)

	}
}

func CancelSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

		var request struct {
			SubscriptionID string `json:"subscription_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Use subscription_id from the request
		var subscription models.Subscription
		if err := db.Where("id = ? AND status != ?", request.SubscriptionID, "cancelled").Last(&subscription).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Active subscription not found"})
			return
		}

		// Cancel Stripe subscription
		_, err := sub.Cancel(subscription.StripeID, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel Stripe subscription"})
			return
		}

		// Update local subscription
		subscription.Status = "cancelled"
		subscription.EndDate = time.Now()

		if err := db.Save(&subscription).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Subscription cancelled successfully"})
	}
}

func getStripePriceID(db *gorm.DB, product models.Product, plan string) (string, error) {
	var stripePriceID string

	switch plan {
	case "monthly":
		if err := db.Model(&product).Select("stripe_monthly_price_id").Scan(&stripePriceID).Error; err != nil {
			utils.Log("Error fetching monthly price ID:", err) // Log the error
			return "", err
		}
	case "yearly":
		if err := db.Model(&product).Select("stripe_yearly_price_id").Scan(&stripePriceID).Error; err != nil {
			utils.Log("Error fetching yearly price ID:", err) // Log the error
			return "", err
		}
	// case "trial":
	// 	if err := db.Model(&product).Select("stripe_trial_price_id").Scan(&stripePriceID).Error; err != nil { // Change to stripe_trial_price_id
	// 		utils.Log("Error fetching trial price ID:", err) // Log the error
	// 		return "", err
	// 	}
	default:
		return "", errors.New("invalid plan type")
	}

	if stripePriceID == "" {
		utils.Log("Stripe price ID not found for plan:", plan) // Log if ID is empty
		return "", errors.New("stripe price ID not found for the given product and plan")
	}

	return stripePriceID, nil
}

func UpdateTrialStatus(db *gorm.DB) error {
	var subscriptions []models.Subscription

	// Get all subscriptions that are still in trial
	if err := db.Where("is_in_trial = ? AND trial_end_date < ?", true, time.Now()).Find(&subscriptions).Error; err != nil {
		return err
	}

	// Update each subscription's IsInTrial status
	for _, subscription := range subscriptions {
		subscription.IsInTrial = false
		if err := db.Save(&subscription).Error; err != nil {
			return err
		}
	}

	return nil
}

// - **Performance**:
// 	Depending on the number of subscriptions,
// 	you may want to optimize the query or run this function
// 	in a background job to avoid performance issues.

// - **Scheduled Jobs**: For a more robust solution,
// 	consider implementing a scheduled job
// 	(using a library like `cron` in Go) to periodically check
//  	and update trial statuses without needing to rely on user actions.

func TrialSubscribe(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var trialRequest struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&trialRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var product models.Product
		if err := db.First(&product, trialRequest.ProductID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		var user models.CustomUser
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		// Check if the user already has an active subscription
		var existingSubscription models.Subscription
		if err := db.Where("user_id = ? AND status != ?", user.ID, "cancelled").First(&existingSubscription).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already has an active subscription"})
			return
		}

		// Check if the user already has an active or completed subscription
		var existingUserSubscription models.Subscription
		if err := db.Where("user_id = ? AND (status = ? OR status = ?)", user.ID, "active", "completed").First(&existingUserSubscription).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already has an active or completed subscription"})
			return
		}

		// Create local trial subscription
		subscription := models.Subscription{
			UserID:       user.ID,
			ProductID:    product.ID,
			StartDate:    time.Now(),
			TrialEndDate: time.Now().AddDate(0, 0, 30), // Set trial period of 30 days
			EndDate:      time.Time{},                  // No end date for trial
			Status:       "active",
			Plan:         "trial", // Set plan to trial
			IsInTrial:    true,    // Set IsInTrial to true
		}

		if err := db.Create(&subscription).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trial subscription"})
			return
		}

		// Exclude User details from the response
		c.JSON(http.StatusCreated, gin.H{
			"message":      "Trial subscription created successfully",
			"subscription": subscription,
		})
	}
}
