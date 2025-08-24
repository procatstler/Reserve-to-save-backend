package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Demo data for testing
var demoData = struct {
	merchants []map[string]interface{}
	campaigns []map[string]interface{}
	users     []map[string]interface{}
}{
	merchants: []map[string]interface{}{
		{
			"id":            uuid.New().String(),
			"wallet":        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
			"name":          "Starbucks Korea",
			"email":         "merchant@starbucks.kr",
			"kyc_tier":      2,
		},
		{
			"id":            uuid.New().String(),
			"wallet":        "0x5B38Da6a701c568545dCfcB03FCB875f56bedDC4",
			"name":          "CU Convenience Store",
			"email":         "merchant@cu.kr",
			"kyc_tier":      2,
		},
	},
	campaigns: []map[string]interface{}{
		{
			"id":              uuid.New().String(),
			"chain_address":   "0x1234567890123456789012345678901234567890",
			"title":           "Starbucks Americano - 30% OFF",
			"description":     "Reserve your daily coffee with 30% discount!",
			"image_url":       "https://example.com/starbucks-coffee.jpg",
			"merchant_wallet": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
			"base_price":      "5000000", // 5 USDT (6 decimals)
			"min_qty":         100,
			"target_amount":   "500000000", // 500 USDT
			"current_amount":  "350000000", // 350 USDT
			"current_qty":     70,
			"discount_rate":   3000, // 30%
			"save_floor_bps":  500,  // 5%
			"r_max_bps":       1000, // 10%
			"status":          "recruiting",
			"start_time":      time.Now().Add(-24 * time.Hour),
			"end_time":        time.Now().Add(7 * 24 * time.Hour),
		},
		{
			"id":              uuid.New().String(),
			"chain_address":   "0x2345678901234567890123456789012345678901",
			"title":           "CU Lunch Box Special",
			"description":     "Get 25% off on lunch boxes!",
			"image_url":       "https://example.com/cu-lunchbox.jpg",
			"merchant_wallet": "0x5B38Da6a701c568545dCfcB03FCB875f56bedDC4",
			"base_price":      "8000000", // 8 USDT
			"min_qty":         50,
			"target_amount":   "400000000", // 400 USDT
			"current_amount":  "280000000", // 280 USDT
			"current_qty":     35,
			"discount_rate":   2500, // 25%
			"save_floor_bps":  400,  // 4%
			"r_max_bps":       800,  // 8%
			"status":          "recruiting",
			"start_time":      time.Now().Add(-12 * time.Hour),
			"end_time":        time.Now().Add(5 * 24 * time.Hour),
		},
		{
			"id":              uuid.New().String(),
			"chain_address":   "0x3456789012345678901234567890123456789012",
			"title":           "GS25 Snack Bundle",
			"description":     "Save 20% on snack bundles!",
			"image_url":       "https://example.com/gs25-snacks.jpg",
			"merchant_wallet": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb1",
			"base_price":      "3000000", // 3 USDT
			"min_qty":         200,
			"target_amount":   "600000000", // 600 USDT
			"current_amount":  "600000000", // 600 USDT (reached!)
			"current_qty":     200,
			"discount_rate":   2000, // 20%
			"save_floor_bps":  300,  // 3%
			"r_max_bps":       600,  // 6%
			"status":          "reached",
			"start_time":      time.Now().Add(-48 * time.Hour),
			"end_time":        time.Now().Add(2 * 24 * time.Hour),
		},
	},
	users: []map[string]interface{}{
		{
			"id":               uuid.New().String(),
			"wallet_address":   "0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2",
			"line_user_id":     "U123456789",
			"line_display_name": "Alice Kim",
			"email":            "alice@example.com",
			"kyc_tier":         1,
		},
		{
			"id":               uuid.New().String(),
			"wallet_address":   "0x4B20993Bc481177ec7E8f571ceCaE8A9e22C02db",
			"line_user_id":     "U987654321",
			"line_display_name": "Bob Lee",
			"email":            "bob@example.com",
			"kyc_tier":         1,
		},
		{
			"id":               uuid.New().String(),
			"wallet_address":   "0x78731D3Ca6b7E34aC0F824c42a7cC18A495cabaB",
			"line_user_id":     "U555666777",
			"line_display_name": "Carol Park",
			"email":            "carol@example.com",
			"kyc_tier":         0,
		},
	},
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

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database")

	// Clear existing demo data
	log.Println("Clearing existing demo data...")
	clearDemoData(db)

	// Insert users
	log.Println("Inserting demo users...")
	insertUsers(db)

	// Insert campaigns
	log.Println("Inserting demo campaigns...")
	insertCampaigns(db)

	// Insert participations
	log.Println("Inserting demo participations...")
	insertParticipations(db)

	log.Println("Demo data seeded successfully!")
	log.Println("\nDemo Users:")
	log.Println("- Alice: 0xAb8483F64d9C6d1EcF9b849Ae677dD3315835cb2")
	log.Println("- Bob:   0x4B20993Bc481177ec7E8f571ceCaE8A9e22C02db")
	log.Println("- Carol: 0x78731D3Ca6b7E34aC0F824c42a7cC18A495cabaB")
	log.Println("\nDemo Campaigns:")
	log.Println("- Starbucks Americano (recruiting)")
	log.Println("- CU Lunch Box (recruiting)")
	log.Println("- GS25 Snacks (reached)")
}

func clearDemoData(db *sql.DB) {
	// Clear in order to respect foreign key constraints
	queries := []string{
		"DELETE FROM participations WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com')",
		"DELETE FROM payments WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com')",
		"DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com')",
		"DELETE FROM campaigns WHERE title LIKE '%Demo%' OR title LIKE '%Starbucks%' OR title LIKE '%CU%' OR title LIKE '%GS25%'",
		"DELETE FROM users WHERE email LIKE '%@example.com' OR email LIKE '%@starbucks.kr' OR email LIKE '%@cu.kr'",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Warning: Failed to clear data: %v", err)
		}
	}
}

func insertUsers(db *sql.DB) {
	// Insert merchants first
	for _, merchant := range demoData.merchants {
		query := `
			INSERT INTO users (id, wallet_address, line_display_name, email, kyc_tier, status)
			VALUES ($1, $2, $3, $4, $5, 'active')
			ON CONFLICT (wallet_address) DO NOTHING`
		
		_, err := db.Exec(query,
			merchant["id"],
			merchant["wallet"],
			merchant["name"],
			merchant["email"],
			merchant["kyc_tier"],
		)
		if err != nil {
			log.Printf("Failed to insert merchant: %v", err)
		}
	}

	// Insert regular users
	for _, user := range demoData.users {
		query := `
			INSERT INTO users (id, wallet_address, line_user_id, line_display_name, email, kyc_tier, status)
			VALUES ($1, $2, $3, $4, $5, $6, 'active')
			ON CONFLICT (wallet_address) DO NOTHING`
		
		_, err := db.Exec(query,
			user["id"],
			user["wallet_address"],
			user["line_user_id"],
			user["line_display_name"],
			user["email"],
			user["kyc_tier"],
		)
		if err != nil {
			log.Printf("Failed to insert user: %v", err)
		}
	}
}

func insertCampaigns(db *sql.DB) {
	for _, campaign := range demoData.campaigns {
		// Get merchant ID
		var merchantID string
		err := db.QueryRow("SELECT id FROM users WHERE wallet_address = $1", campaign["merchant_wallet"]).Scan(&merchantID)
		if err != nil {
			log.Printf("Failed to get merchant ID: %v", err)
			continue
		}

		query := `
			INSERT INTO campaigns (
				id, chain_address, title, description, image_url,
				merchant_id, merchant_wallet, base_price, min_qty,
				target_amount, current_amount, current_qty,
				discount_rate, save_floor_bps, r_max_bps,
				merchant_fee_bps, ops_fee_bps,
				status, start_time, end_time
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			) ON CONFLICT (chain_address) DO NOTHING`
		
		_, err = db.Exec(query,
			campaign["id"],
			campaign["chain_address"],
			campaign["title"],
			campaign["description"],
			campaign["image_url"],
			merchantID,
			campaign["merchant_wallet"],
			campaign["base_price"],
			campaign["min_qty"],
			campaign["target_amount"],
			campaign["current_amount"],
			campaign["current_qty"],
			campaign["discount_rate"],
			campaign["save_floor_bps"],
			campaign["r_max_bps"],
			250, // merchant_fee_bps
			100, // ops_fee_bps
			campaign["status"],
			campaign["start_time"],
			campaign["end_time"],
		)
		if err != nil {
			log.Printf("Failed to insert campaign: %v", err)
		}
	}
}

func insertParticipations(db *sql.DB) {
	// Create some demo participations
	participations := []struct {
		userEmail    string
		campaignTitle string
		amount       string
	}{
		{"alice@example.com", "Starbucks Americano - 30% OFF", "50000000"},  // Alice: 50 USDT
		{"bob@example.com", "Starbucks Americano - 30% OFF", "100000000"},   // Bob: 100 USDT
		{"carol@example.com", "CU Lunch Box Special", "80000000"},           // Carol: 80 USDT
		{"alice@example.com", "CU Lunch Box Special", "40000000"},           // Alice: 40 USDT
		{"bob@example.com", "GS25 Snack Bundle", "60000000"},                // Bob: 60 USDT (in reached campaign)
	}

	for _, p := range participations {
		// Get user ID
		var userID, walletAddress string
		err := db.QueryRow("SELECT id, wallet_address FROM users WHERE email = $1", p.userEmail).Scan(&userID, &walletAddress)
		if err != nil {
			log.Printf("Failed to get user ID for %s: %v", p.userEmail, err)
			continue
		}

		// Get campaign ID
		var campaignID string
		err = db.QueryRow("SELECT id FROM campaigns WHERE title = $1", p.campaignTitle).Scan(&campaignID)
		if err != nil {
			log.Printf("Failed to get campaign ID for %s: %v", p.campaignTitle, err)
			continue
		}

		query := `
			INSERT INTO participations (
				id, campaign_id, user_id, wallet_address,
				deposit_amount, expected_rebate, status
			) VALUES (
				$1, $2, $3, $4, $5, $6, 'active'
			) ON CONFLICT (campaign_id, user_id) DO NOTHING`
		
		// Calculate expected rebate (simplified: 7% of deposit)
		depositAmount := new(big.Int)
		depositAmount.SetString(p.amount, 10)
		expectedRebate := new(big.Int).Div(new(big.Int).Mul(depositAmount, big.NewInt(700)), big.NewInt(10000))
		
		_, err = db.Exec(query,
			uuid.New().String(),
			campaignID,
			userID,
			walletAddress,
			p.amount,
			expectedRebate.String(),
		)
		if err != nil {
			log.Printf("Failed to insert participation: %v", err)
		}
	}
}