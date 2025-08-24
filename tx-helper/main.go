package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"r2s/tx-helper/handlers"
	"r2s/tx-helper/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize services
	txService := services.NewTransactionService(
		os.Getenv("BLOCKCHAIN_RPC_URL"),
		os.Getenv("CAMPAIGN_FACTORY_ADDRESS"),
		os.Getenv("USDT_ADDRESS"),
	)

	// Initialize handlers
	txHandler := handlers.NewTransactionHandler(txService)

	// Setup router
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "tx-helper",
		})
	})

	// Transaction routes
	txGroup := router.Group("/tx")
	{
		// Campaign transactions
		txGroup.POST("/join-campaign", txHandler.BuildJoinCampaignTx)
		txGroup.POST("/cancel-participation", txHandler.BuildCancelParticipationTx)
		txGroup.POST("/request-cancel", txHandler.BuildRequestCancelTx)
		
		// Merchant transactions
		txGroup.POST("/confirm-fulfillment", txHandler.BuildConfirmFulfillmentTx)
		txGroup.POST("/settle-campaign", txHandler.BuildSettleCampaignTx)
		
		// Utility
		txGroup.POST("/approve-usdt", txHandler.BuildApproveUSDTTx)
		txGroup.GET("/estimate-gas", txHandler.EstimateGas)
		txGroup.GET("/campaign-info", txHandler.GetCampaignInfo)
	}

	// Start server
	port := os.Getenv("TX_HELPER_PORT")
	if port == "" {
		port = "3006"
	}

	log.Printf("TX Helper starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}