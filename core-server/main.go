package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"r2s/core-server/handlers"
	"r2s/core-server/services"
	"r2s/pkg/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Database configuration
	dbConfig := database.Config{
		Host:         os.Getenv("DB_HOST"),
		Port:         5432,
		User:         os.Getenv("DB_USER"),
		Password:     os.Getenv("DB_PASSWORD"),
		Database:     os.Getenv("DB_NAME"),
		MaxOpenConns: 25,
		MaxIdleConns: 10,
		MaxLifetime:  5 * time.Minute,
	}

	// Initialize database
	db, err := database.NewDB(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Redis configuration
	redisConfig := database.RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     6379,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		PoolSize: 10,
	}

	// Initialize Redis
	redis, err := database.NewRedisClient(redisConfig)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redis.Close()

	// Initialize services
	campaignService := services.NewCampaignService(db, redis)
	participationService := services.NewParticipationService(db, redis)
	paymentService := services.NewPaymentService(db, redis)

	// Initialize handlers
	campaignHandler := handlers.NewCampaignHandler(campaignService)
	participationHandler := handlers.NewParticipationHandler(participationService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	// Setup router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "core-server",
		})
	})

	// Campaign routes
	campaignGroup := router.Group("/campaigns")
	{
		campaignGroup.GET("", campaignHandler.ListCampaigns)
		campaignGroup.GET("/:id", campaignHandler.GetCampaign)
		campaignGroup.POST("", campaignHandler.CreateCampaign)
		campaignGroup.PUT("/:id", campaignHandler.UpdateCampaign)
		campaignGroup.POST("/:id/settle", campaignHandler.SettleCampaign)
	}

	// Participation routes
	participationGroup := router.Group("/participations")
	{
		participationGroup.GET("/user/:userId", participationHandler.GetUserParticipations)
		participationGroup.GET("/campaign/:campaignId", participationHandler.GetCampaignParticipations)
		participationGroup.POST("", participationHandler.CreateParticipation)
		participationGroup.PUT("/:id/cancel", participationHandler.CancelParticipation)
	}

	// Payment routes
	paymentGroup := router.Group("/payments")
	{
		paymentGroup.POST("/process", paymentHandler.ProcessPayment)
		paymentGroup.GET("/:id/status", paymentHandler.GetPaymentStatus)
		paymentGroup.POST("/webhook", paymentHandler.HandleWebhook)
	}

	// Start server
	port := os.Getenv("CORE_SERVER_PORT")
	if port == "" {
		port = "3003"
	}

	log.Printf("Core server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}