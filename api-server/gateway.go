package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceConfig holds the configuration for a microservice
type ServiceConfig struct {
	Name    string
	BaseURL string
	Timeout time.Duration
}

// Gateway handles routing requests to microservices
type Gateway struct {
	services map[string]*ServiceConfig
	client   *http.Client
}

// NewGateway creates a new API gateway
func NewGateway() *Gateway {
	return &Gateway{
		services: map[string]*ServiceConfig{
			"auth": {
				Name:    "auth-server",
				BaseURL: "http://localhost:3002",
				Timeout: 10 * time.Second,
			},
			"core": {
				Name:    "core-server",
				BaseURL: "http://localhost:3003",
				Timeout: 30 * time.Second,
			},
			"query": {
				Name:    "query-server",
				BaseURL: "http://localhost:3004",
				Timeout: 10 * time.Second,
			},
			"batch": {
				Name:    "batch-server",
				BaseURL: "http://localhost:3005",
				Timeout: 60 * time.Second,
			},
			"tx-helper": {
				Name:    "tx-helper",
				BaseURL: "http://localhost:3006",
				Timeout: 20 * time.Second,
			},
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest forwards a request to the appropriate microservice
func (g *Gateway) ProxyRequest(c *gin.Context, service string, path string) {
	config, exists := g.services[service]
	if !exists {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Service '%s' not found", service),
		})
		return
	}

	// Build target URL
	targetURL := config.BaseURL + path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}

	// Create new request
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create request",
		})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set timeout for this specific request
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Failed to reach %s service", service),
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read response",
		})
		return
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Return response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// AuthMiddleware validates JWT tokens by calling auth-server
func (g *Gateway) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for certain paths
		if strings.HasPrefix(c.Request.URL.Path, "/api/auth/") || 
		   c.Request.URL.Path == "/health" ||
		   c.Request.URL.Path == "/api-docs" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort()
			return
		}

		// Validate token with auth-server
		req, _ := http.NewRequest("GET", g.services["auth"].BaseURL+"/auth/validate", nil)
		req.Header.Set("Authorization", authHeader)

		resp, err := g.client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token",
			})
			c.Abort()
			return
		}
		defer resp.Body.Close()

		// Parse claims from response
		var result struct {
			Success bool                   `json:"success"`
			Claims  map[string]interface{} `json:"claims"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || !result.Success {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Token validation failed",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user", result.Claims)
		c.Next()
	}
}

// SetupRoutes configures all API routes
func (g *Gateway) SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "api-gateway",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (no auth middleware)
		auth := api.Group("/auth")
		{
			auth.GET("/nonce", func(c *gin.Context) {
				g.ProxyRequest(c, "auth", "/auth/nonce")
			})
			auth.POST("/verify", func(c *gin.Context) {
				g.ProxyRequest(c, "auth", "/auth/verify")
			})
			auth.POST("/line", func(c *gin.Context) {
				g.ProxyRequest(c, "auth", "/auth/line")
			})
			auth.POST("/refresh", func(c *gin.Context) {
				g.ProxyRequest(c, "auth", "/auth/refresh")
			})
			auth.POST("/logout", func(c *gin.Context) {
				g.ProxyRequest(c, "auth", "/auth/logout")
			})
		}

		// Protected routes (require auth)
		protected := api.Group("/")
		protected.Use(g.AuthMiddleware())
		{
			// Campaign routes
			campaigns := protected.Group("/campaigns")
			{
				campaigns.GET("", func(c *gin.Context) {
					g.ProxyRequest(c, "query", "/campaigns")
				})
				campaigns.GET("/:id", func(c *gin.Context) {
					g.ProxyRequest(c, "query", "/campaigns/"+c.Param("id"))
				})
				campaigns.POST("", func(c *gin.Context) {
					g.ProxyRequest(c, "core", "/campaigns")
				})
				campaigns.PUT("/:id", func(c *gin.Context) {
					g.ProxyRequest(c, "core", "/campaigns/"+c.Param("id"))
				})
			}

			// Payment routes
			payments := protected.Group("/payment")
			{
				payments.POST("/create", func(c *gin.Context) {
					g.ProxyRequest(c, "core", "/payments/process")
				})
				payments.GET("/:id/status", func(c *gin.Context) {
					g.ProxyRequest(c, "core", "/payments/"+c.Param("id")+"/status")
				})
			}

			// Participation routes
			participations := protected.Group("/participations")
			{
				participations.GET("/my", func(c *gin.Context) {
					// Get user ID from context
					user, _ := c.Get("user")
					userClaims := user.(map[string]interface{})
					userID := userClaims["user_id"].(string)
					g.ProxyRequest(c, "query", "/participations/user/"+userID)
				})
				participations.POST("/cancel", func(c *gin.Context) {
					g.ProxyRequest(c, "tx-helper", "/tx/cancel-participation")
				})
			}

			// Transaction helper routes
			tx := protected.Group("/tx")
			{
				tx.POST("/join", func(c *gin.Context) {
					g.ProxyRequest(c, "tx-helper", "/tx/join-campaign")
				})
				tx.POST("/cancel", func(c *gin.Context) {
					g.ProxyRequest(c, "tx-helper", "/tx/cancel-participation")
				})
				tx.GET("/estimate-gas", func(c *gin.Context) {
					g.ProxyRequest(c, "tx-helper", "/tx/estimate-gas")
				})
			}

			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", func(c *gin.Context) {
					g.ProxyRequest(c, "query", "/users/profile")
				})
				users.PUT("/profile", func(c *gin.Context) {
					g.ProxyRequest(c, "core", "/users/profile")
				})
			}
		}
	}

	// Webhook routes (no auth, but verify signature)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/payment", func(c *gin.Context) {
			g.ProxyRequest(c, "core", "/payments/webhook")
		})
		webhooks.POST("/blockchain", func(c *gin.Context) {
			g.ProxyRequest(c, "event-receiver", "/events/webhook")
		})
	}
}