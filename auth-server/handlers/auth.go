package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"r2s/auth-server/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// GetNonce generates a nonce for wallet authentication
func (h *AuthHandler) GetNonce(c *gin.Context) {
	address := c.Query("address")
	chainID := c.DefaultQuery("chainId", "1001")

	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Address is required",
		})
		return
	}

	nonce, message, requestID, expiresAt, err := h.authService.GenerateNonce(address, chainID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nonce":     nonce,
		"message":   message,
		"requestId": requestID,
		"expiresAt": expiresAt,
	})
}

// VerifySignature verifies wallet signature and issues JWT
func (h *AuthHandler) VerifySignature(c *gin.Context) {
	var req struct {
		Address   string `json:"address" binding:"required"`
		Signature string `json:"signature" binding:"required"`
		Message   string `json:"message" binding:"required"`
		RequestID string `json:"requestId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	tokens, user, err := h.authService.VerifySignature(
		req.Address,
		req.Signature,
		req.Message,
		req.RequestID,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"user": gin.H{
			"id":            user.ID,
			"address":       user.WalletAddress,
			"kycTier":       user.KYCTier,
			"lineConnected": user.LineUserID != nil,
		},
	})
}

// LineAuth handles LINE authentication
func (h *AuthHandler) LineAuth(c *gin.Context) {
	var req struct {
		IDToken     string `json:"idToken" binding:"required"`
		AccessToken string `json:"accessToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	token, user, err := h.authService.LineAuth(
		req.IDToken,
		req.AccessToken,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user": gin.H{
			"id":              user.ID,
			"lineUserId":      user.LineUserID,
			"displayName":     user.LineDisplayName,
			"pictureUrl":      user.LinePictureURL,
			"walletConnected": user.WalletAddress != "",
			"kycTier":         user.KYCTier,
		},
	})
}

// RefreshToken refreshes an access token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	accessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"accessToken": accessToken,
	})
}

// Logout invalidates the current session
func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Token required",
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if err := h.authService.Logout(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to logout",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// ValidateToken validates a JWT token (internal use)
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Token required",
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"claims":  claims,
	})
}