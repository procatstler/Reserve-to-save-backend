package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Create gateway
	gateway := NewGateway()

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Rate limiting middleware
	// router.Use(RateLimitMiddleware())

	// Setup routes
	gateway.SetupRoutes(router)

	// Serve Swagger documentation
	router.Static("/api-docs", "./docs/swagger-ui")
	router.StaticFile("/swagger.json", "./docs/swagger.json")

	// Start server
	port := os.Getenv("API_SERVER_PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("API Gateway starting on port %s", port)
	log.Printf("Swagger UI available at http://localhost:%s/api-docs", port)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}