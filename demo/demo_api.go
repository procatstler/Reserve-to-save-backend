package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type DemoAPI struct {
	db *sql.DB
}

func NewDemoAPI(db *sql.DB) *DemoAPI {
	return &DemoAPI{db: db}
}

// GetDemoUsers returns list of demo users for testing
func (d *DemoAPI) GetDemoUsers(c *gin.Context) {
	query := `
		SELECT id, wallet_address, line_user_id, line_display_name, email, kyc_tier
		FROM users
		WHERE email LIKE '%@example.com'
		ORDER BY line_display_name`
	
	rows, err := d.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, wallet, lineID, name, email string
		var kycTier int
		var lineIDPtr, namePtr *string
		
		err := rows.Scan(&id, &wallet, &lineIDPtr, &namePtr, &email, &kycTier)
		if err != nil {
			continue
		}
		
		if lineIDPtr != nil {
			lineID = *lineIDPtr
		}
		if namePtr != nil {
			name = *namePtr
		}
		
		users = append(users, map[string]interface{}{
			"id":           id,
			"wallet":       wallet,
			"lineUserId":   lineID,
			"displayName":  name,
			"email":        email,
			"kycTier":      kycTier,
			"testPassword": "demo123", // For demo purposes only
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"message": "Use these demo accounts for testing",
	})
}

// GetDemoCampaigns returns list of demo campaigns
func (d *DemoAPI) GetDemoCampaigns(c *gin.Context) {
	query := `
		SELECT 
			c.id, c.chain_address, c.title, c.description, c.image_url,
			c.merchant_wallet, c.base_price, c.min_qty, c.target_amount,
			c.current_amount, c.current_qty, c.discount_rate,
			c.save_floor_bps, c.r_max_bps, c.status,
			c.start_time, c.end_time,
			COUNT(DISTINCT p.user_id) as participant_count
		FROM campaigns c
		LEFT JOIN participations p ON c.id = p.campaign_id
		WHERE c.title LIKE '%Starbucks%' OR c.title LIKE '%CU%' OR c.title LIKE '%GS25%'
		GROUP BY c.id
		ORDER BY c.created_at DESC`
	
	rows, err := d.db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var campaigns []map[string]interface{}
	for rows.Next() {
		var id, chainAddr, title, desc, imgURL, merchantWallet string
		var basePrice, targetAmount, currentAmount string
		var minQty, currentQty, discountRate, saveFloor, rMax int
		var status string
		var startTime, endTime sql.NullTime
		var participantCount int
		var descPtr, imgURLPtr *string
		
		err := rows.Scan(
			&id, &chainAddr, &title, &descPtr, &imgURLPtr,
			&merchantWallet, &basePrice, &minQty, &targetAmount,
			&currentAmount, &currentQty, &discountRate,
			&saveFloor, &rMax, &status,
			&startTime, &endTime, &participantCount,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		
		if descPtr != nil {
			desc = *descPtr
		}
		if imgURLPtr != nil {
			imgURL = *imgURLPtr
		}
		
		// Calculate progress
		var progress float64
		if targetAmount != "0" {
			currentFloat := float64(0)
			targetFloat := float64(1)
			fmt.Sscanf(currentAmount, "%f", &currentFloat)
			fmt.Sscanf(targetAmount, "%f", &targetFloat)
			progress = (currentFloat / targetFloat) * 100
		}
		
		campaigns = append(campaigns, map[string]interface{}{
			"id":               id,
			"chainAddress":     chainAddr,
			"title":            title,
			"description":      desc,
			"imageUrl":         imgURL,
			"merchantWallet":   merchantWallet,
			"basePrice":        basePrice,
			"minQty":           minQty,
			"currentQty":       currentQty,
			"targetAmount":     targetAmount,
			"currentAmount":    currentAmount,
			"discountRate":     discountRate,
			"saveFloorBps":     saveFloor,
			"rMaxBps":          rMax,
			"status":           status,
			"progress":         progress,
			"participantCount": participantCount,
			"startTime":        startTime.Time,
			"endTime":          endTime.Time,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    campaigns,
	})
}

// GetDemoAuth provides a demo authentication token without signature
func (d *DemoAPI) GetDemoAuth(c *gin.Context) {
	wallet := c.Query("wallet")
	if wallet == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Wallet address required",
		})
		return
	}

	// For demo, return a simple token
	// In production, this would go through proper authentication
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"accessToken":  "demo_token_" + wallet,
			"refreshToken": "demo_refresh_" + wallet,
			"message":      "This is a demo token for testing only",
		},
	})
}

// GetDemoStats returns demo statistics
func (d *DemoAPI) GetDemoStats(c *gin.Context) {
	var stats struct {
		TotalUsers         int
		TotalCampaigns     int
		ActiveCampaigns    int
		TotalParticipations int
		TotalDeposited     string
	}

	// Get user count
	d.db.QueryRow("SELECT COUNT(*) FROM users WHERE email LIKE '%@example.com'").Scan(&stats.TotalUsers)
	
	// Get campaign counts
	d.db.QueryRow("SELECT COUNT(*) FROM campaigns").Scan(&stats.TotalCampaigns)
	d.db.QueryRow("SELECT COUNT(*) FROM campaigns WHERE status = 'recruiting'").Scan(&stats.ActiveCampaigns)
	
	// Get participation stats
	d.db.QueryRow("SELECT COUNT(*) FROM participations").Scan(&stats.TotalParticipations)
	d.db.QueryRow("SELECT COALESCE(SUM(CAST(deposit_amount AS NUMERIC)), 0) FROM participations").Scan(&stats.TotalDeposited)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:password@localhost:5432/r2s_dev?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize demo API
	demoAPI := NewDemoAPI(db)

	// Setup router
	router := gin.Default()

	// CORS for demo
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Demo routes
	demo := router.Group("/demo")
	{
		demo.GET("/users", demoAPI.GetDemoUsers)
		demo.GET("/campaigns", demoAPI.GetDemoCampaigns)
		demo.GET("/auth", demoAPI.GetDemoAuth)
		demo.GET("/stats", demoAPI.GetDemoStats)
		demo.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "demo-api",
			})
		})
	}

	// Info endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "R2S Demo API",
			"version": "1.0.0",
			"endpoints": []string{
				"GET /demo/users - Get demo user accounts",
				"GET /demo/campaigns - Get demo campaigns",
				"GET /demo/auth?wallet=0x... - Get demo auth token",
				"GET /demo/stats - Get demo statistics",
			},
			"message": "Use these endpoints for testing and demo purposes",
		})
	})

	// Start server
	port := os.Getenv("DEMO_PORT")
	if port == "" {
		port = "3008"
	}

	log.Printf("Demo API starting on port %s", port)
	log.Printf("Access demo endpoints at http://localhost:%s/demo", port)
	
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}