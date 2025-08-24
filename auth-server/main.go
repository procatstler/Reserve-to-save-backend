package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"r2s/auth-server/handlers"
	"r2s/auth-server/repository"
	"r2s/auth-server/services"
	"r2s/pkg/database"
	"r2s/pkg/utils"
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

	// Initialize JWT Manager
	jwtManager := utils.NewJWTManager(
		os.Getenv("JWT_SECRET"),
		os.Getenv("JWT_REFRESH_SECRET"),
		15*time.Minute,
		7*24*time.Hour,
	)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, redis, jwtManager)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Setup router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "auth-server",
		})
	})

	// Auth routes
	authGroup := router.Group("/auth")
	{
		authGroup.GET("/nonce", authHandler.GetNonce)
		authGroup.POST("/verify", authHandler.VerifySignature)
		authGroup.POST("/line", authHandler.LineAuth)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.GET("/validate", authHandler.ValidateToken)
	}

	// Start server
	port := os.Getenv("AUTH_SERVER_PORT")
	if port == "" {
		port = "3002"
	}

	log.Printf("Auth server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}