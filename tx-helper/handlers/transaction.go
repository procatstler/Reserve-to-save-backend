package handlers

import (
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"r2s/tx-helper/services"
)

type TransactionHandler struct {
	txService *services.TransactionService
}

func NewTransactionHandler(txService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		txService: txService,
	}
}

// BuildJoinCampaignTx handles POST /tx/join-campaign
func (h *TransactionHandler) BuildJoinCampaignTx(c *gin.Context) {
	var req struct {
		UserAddress     string `json:"userAddress" binding:"required"`
		CampaignAddress string `json:"campaignAddress" binding:"required"`
		Amount          string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	amount := new(big.Int)
	amount.SetString(req.Amount, 10)

	txMessage, err := h.txService.BuildJoinCampaignTx(
		req.UserAddress,
		req.CampaignAddress,
		amount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"transaction": txMessage,
			"message":     "Sign and send this transaction to join the campaign",
		},
	})
}

// BuildCancelParticipationTx handles POST /tx/cancel-participation
func (h *TransactionHandler) BuildCancelParticipationTx(c *gin.Context) {
	var req struct {
		UserAddress     string `json:"userAddress" binding:"required"`
		CampaignAddress string `json:"campaignAddress" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	// For full cancellation, we need to get user's deposit amount from campaign
	// This is simplified for demo
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Use /tx/request-cancel with specific amount",
		},
	})
}

// BuildRequestCancelTx handles POST /tx/request-cancel
func (h *TransactionHandler) BuildRequestCancelTx(c *gin.Context) {
	var req struct {
		UserAddress     string `json:"userAddress" binding:"required"`
		CampaignAddress string `json:"campaignAddress" binding:"required"`
		Amount          string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	amount := new(big.Int)
	amount.SetString(req.Amount, 10)

	txMessage, err := h.txService.BuildRequestCancelTx(
		req.UserAddress,
		req.CampaignAddress,
		amount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"transaction": txMessage,
			"message":     "Sign and send this transaction to request cancellation",
		},
	})
}

// BuildApproveUSDTTx handles POST /tx/approve-usdt
func (h *TransactionHandler) BuildApproveUSDTTx(c *gin.Context) {
	var req struct {
		UserAddress    string `json:"userAddress" binding:"required"`
		SpenderAddress string `json:"spenderAddress" binding:"required"`
		Amount         string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}

	amount := new(big.Int)
	amount.SetString(req.Amount, 10)

	txMessage, err := h.txService.BuildApproveUSDTTx(
		req.UserAddress,
		req.SpenderAddress,
		amount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"transaction": txMessage,
			"message":     "Sign and send this transaction to approve USDT spending",
		},
	})
}

// BuildConfirmFulfillmentTx handles POST /tx/confirm-fulfillment
func (h *TransactionHandler) BuildConfirmFulfillmentTx(c *gin.Context) {
	// Simplified for demo
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Merchant only function - not implemented for demo",
		},
	})
}

// BuildSettleCampaignTx handles POST /tx/settle-campaign
func (h *TransactionHandler) BuildSettleCampaignTx(c *gin.Context) {
	// Simplified for demo
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Admin only function - not implemented for demo",
		},
	})
}

// EstimateGas handles GET /tx/estimate-gas
func (h *TransactionHandler) EstimateGas(c *gin.Context) {
	gasPrice, err := h.txService.EstimateGasPrice()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"gasPrice":     gasPrice.String(),
			"gasPriceGwei": new(big.Int).Div(gasPrice, big.NewInt(1e9)).String(),
		},
	})
}

// GetCampaignInfo handles GET /tx/campaign-info
func (h *TransactionHandler) GetCampaignInfo(c *gin.Context) {
	campaignAddress := c.Query("address")
	if campaignAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Campaign address is required",
		})
		return
	}

	info, err := h.txService.GetCampaignInfo(campaignAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}