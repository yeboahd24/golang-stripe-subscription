package routes

import (
	"log"

	"github.com/yeboahd24/subscription-stripe/config"
	"github.com/yeboahd24/subscription-stripe/handlers"
	"github.com/yeboahd24/subscription-stripe/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {

	config, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	authHandler := handlers.NewAuthHandler(db, config.JWTSecret)

	// Public routes
	r.POST("/register", handlers.Register(authHandler))
	r.POST("/login", handlers.Login(authHandler))

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware(config.JWTSecret))
	{
		protected.GET("/products", handlers.GetProducts(db))
		protected.POST("/subscribe", handlers.Subscribe(db))
		protected.GET("/subscription", handlers.GetSubscription(db))
		protected.POST("/cancel-subscription", handlers.CancelSubscription(db))
		protected.POST("/create-product", handlers.CreateProductHandler(db))
		protected.POST("/promote-to-admin", handlers.PromoteToAdmin(db))
		protected.POST("/trial-subscribe", handlers.TrialSubscribe(db))
	}
}
