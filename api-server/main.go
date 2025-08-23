package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Reserve-to-save-backend/pkg/proto/query"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// APIServer는 REST API 서버입니다
type APIServer struct {
	queryClient query.QueryServiceClient
}

// NewAPIServer는 새로운 APIServer 인스턴스를 생성합니다
func NewAPIServer(queryClient query.QueryServiceClient) *APIServer {
	return &APIServer{
		queryClient: queryClient,
	}
}

// GetCampaigns는 GET /query/campaigns 엔드포인트를 처리합니다
func (s *APIServer) GetCampaigns(c *gin.Context) {
	// 쿼리 파라미터 파싱
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	state, _ := strconv.Atoi(c.DefaultQuery("state", "0"))

	log.Printf("REST API called: limit=%d, offset=%d, state=%d", limit, offset, state)

	// gRPC 요청 생성
	req := &query.GetCampaignsRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
		State:  int32(state),
	}

	// gRPC 호출 (5초 타임아웃)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.queryClient.GetCampaigns(ctx, req)
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get campaigns",
		})
		return
	}

	log.Printf("gRPC response: %d campaigns, total=%d", len(resp.Campaigns), resp.TotalCount)

	// 응답 변환 (protobuf → JSON)
	campaigns := make([]map[string]interface{}, len(resp.Campaigns))
	for i, campaign := range resp.Campaigns {
		campaigns[i] = map[string]interface{}{
			"id":               campaign.Id,
			"address":          campaign.Address,
			"merchant_id":      campaign.MerchantId,
			"merchant_name":    campaign.MerchantName,
			"base_price":       campaign.BasePrice,
			"min_qty":          campaign.MinQty,
			"lock_start":       campaign.LockStart.AsTime().Format(time.RFC3339),
			"lock_end":         campaign.LockEnd.AsTime().Format(time.RFC3339),
			"rmax_bps":         campaign.RmaxBps,
			"savefloor_bps":    campaign.SavefloorBps,
			"merchant_fee_bps": campaign.MerchantFeeBps,
			"ops_fee_bps":      campaign.OpsFeeBps,
			"state":            campaign.State,
			"metadata_uri":     campaign.MetadataUri,
			"created_at":       campaign.CreatedAt.AsTime().Format(time.RFC3339),
		}
	}

	// JSON 응답
	c.JSON(http.StatusOK, gin.H{
		"campaigns":   campaigns,
		"total_count": resp.TotalCount,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetCampaign은 GET /query/campaigns/:id 엔드포인트를 처리합니다
func (s *APIServer) GetCampaign(c *gin.Context) {
	// 경로 파라미터 파싱
	campaignIDStr := c.Param("id")
	campaignID, err := strconv.ParseInt(campaignIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid campaign ID",
		})
		return
	}

	log.Printf("REST API called: campaign_id=%d", campaignID)

	// gRPC 요청 생성
	req := &query.GetCampaignRequest{
		CampaignId: campaignID,
	}

	// gRPC 호출
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.queryClient.GetCampaign(ctx, req)
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get campaign",
		})
		return
	}

	if !resp.Found {
		log.Printf("Campaign not found: %d", campaignID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Campaign not found",
		})
		return
	}

	campaign := resp.Campaign
	log.Printf("gRPC response: found campaign %s", campaign.Address)

	// 응답 변환 (protobuf → JSON)
	result := map[string]interface{}{
		"id":               campaign.Id,
		"address":          campaign.Address,
		"merchant_id":      campaign.MerchantId,
		"merchant_name":    campaign.MerchantName,
		"base_price":       campaign.BasePrice,
		"min_qty":          campaign.MinQty,
		"lock_start":       campaign.LockStart.AsTime().Format(time.RFC3339),
		"lock_end":         campaign.LockEnd.AsTime().Format(time.RFC3339),
		"rmax_bps":         campaign.RmaxBps,
		"savefloor_bps":    campaign.SavefloorBps,
		"merchant_fee_bps": campaign.MerchantFeeBps,
		"ops_fee_bps":      campaign.OpsFeeBps,
		"state":            campaign.State,
		"metadata_uri":     campaign.MetadataUri,
		"created_at":       campaign.CreatedAt.AsTime().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, result)
}

// HealthCheck는 GET /health 엔드포인트를 처리합니다
func (s *APIServer) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func main() {
	// gRPC 클라이언트 연결
	queryConn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to query-server: %v", err)
	}
	defer queryConn.Close()

	queryClient := query.NewQueryServiceClient(queryConn)
	log.Println("Connected to query-server via gRPC")

	// API 서버 생성
	apiServer := NewAPIServer(queryClient)

	// Gin 라우터 설정
	router := gin.Default()

	// CORS 미들웨어 (필요시)
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	})

	// 라우트 등록
	router.GET("/health", apiServer.HealthCheck)
	router.GET("/query/campaigns", apiServer.GetCampaigns)
	router.GET("/query/campaigns/:id", apiServer.GetCampaign)

	// 서버 시작
	log.Println("API server starting on :8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 