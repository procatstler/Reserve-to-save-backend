package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"github.com/Reserve-to-save-backend/pkg/proto/query"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// QueryServer는 gRPC QueryService를 구현합니다
type QueryServer struct {
	query.UnimplementedQueryServiceServer
	db *sql.DB
}

// NewQueryServer는 새로운 QueryServer 인스턴스를 생성합니다
func NewQueryServer(db *sql.DB) *QueryServer {
	return &QueryServer{db: db}
}

// GetCampaigns는 캠페인 목록을 조회합니다
func (s *QueryServer) GetCampaigns(ctx context.Context, req *query.GetCampaignsRequest) (*query.GetCampaignsResponse, error) {
	log.Printf("GetCampaigns called with limit=%d, offset=%d, state=%d", req.Limit, req.Offset, req.State)

	// 기본값 설정
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// SQL 쿼리 구성
	baseQuery := `
		SELECT 
			c.id, c.address, c.merchant_id, m.name as merchant_name,
			c.base_price, c.min_qty, c.lock_start, c.lock_end,
			c.rmax_bps, c.savefloor_bps, c.merchant_fee_bps, c.ops_fee_bps,
			c.state, c.metadata_uri, c.created_at
		FROM campaigns c
		JOIN merchants m ON c.merchant_id = m.id
	`
	
	countQuery := "SELECT COUNT(*) FROM campaigns c"
	
	var whereClause string
	var args []interface{}
	
	// 상태 필터 적용
	if req.State > 0 {
		whereClause = " WHERE c.state = $1"
		args = append(args, req.State)
		baseQuery += whereClause
		countQuery += whereClause
	}
	
	// 페이징 추가
	baseQuery += fmt.Sprintf(" ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	// 총 개수 조회
	var totalCount int64
	err := s.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting campaigns: %v", err)
		return nil, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// 캠페인 목록 조회
	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		log.Printf("Error querying campaigns: %v", err)
		return nil, fmt.Errorf("failed to query campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*query.Campaign
	
	for rows.Next() {
		var c query.Campaign
		var addressBytes []byte
		var lockStart, lockEnd, createdAt sql.NullTime
		
		err := rows.Scan(
			&c.Id, &addressBytes, &c.MerchantId, &c.MerchantName,
			&c.BasePrice, &c.MinQty, &lockStart, &lockEnd,
			&c.RmaxBps, &c.SavefloorBps, &c.MerchantFeeBps, &c.OpsFeeBps,
			&c.State, &c.MetadataUri, &createdAt,
		)
		if err != nil {
			log.Printf("Error scanning campaign row: %v", err)
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}

		// BYTEA를 hex string으로 변환
		c.Address = "0x" + hex.EncodeToString(addressBytes)
		
		// timestamp 변환
		if lockStart.Valid {
			c.LockStart = timestamppb.New(lockStart.Time)
		}
		if lockEnd.Valid {
			c.LockEnd = timestamppb.New(lockEnd.Time)
		}
		if createdAt.Valid {
			c.CreatedAt = timestamppb.New(createdAt.Time)
		}

		campaigns = append(campaigns, &c)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating campaign rows: %v", err)
		return nil, fmt.Errorf("failed to iterate campaigns: %w", err)
	}

	response := &query.GetCampaignsResponse{
		Campaigns:  campaigns,
		TotalCount: totalCount,
	}

	log.Printf("Returning %d campaigns, total count: %d", len(campaigns), totalCount)
	return response, nil
}

// GetCampaign은 특정 캠페인을 조회합니다
func (s *QueryServer) GetCampaign(ctx context.Context, req *query.GetCampaignRequest) (*query.GetCampaignResponse, error) {
	log.Printf("GetCampaign called with campaign_id=%d", req.CampaignId)

	sqlQuery := `
		SELECT 
			c.id, c.address, c.merchant_id, m.name as merchant_name,
			c.base_price, c.min_qty, c.lock_start, c.lock_end,
			c.rmax_bps, c.savefloor_bps, c.merchant_fee_bps, c.ops_fee_bps,
			c.state, c.metadata_uri, c.created_at
		FROM campaigns c
		JOIN merchants m ON c.merchant_id = m.id
		WHERE c.id = $1
	`

	var c query.Campaign
	var addressBytes []byte
	var lockStart, lockEnd, createdAt sql.NullTime

	err := s.db.QueryRowContext(ctx, sqlQuery, req.CampaignId).Scan(
		&c.Id, &addressBytes, &c.MerchantId, &c.MerchantName,
		&c.BasePrice, &c.MinQty, &lockStart, &lockEnd,
		&c.RmaxBps, &c.SavefloorBps, &c.MerchantFeeBps, &c.OpsFeeBps,
		&c.State, &c.MetadataUri, &createdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Campaign not found: %d", req.CampaignId)
			return &query.GetCampaignResponse{Found: false}, nil
		}
		log.Printf("Error querying campaign: %v", err)
		return nil, fmt.Errorf("failed to query campaign: %w", err)
	}

	// BYTEA를 hex string으로 변환
	c.Address = "0x" + hex.EncodeToString(addressBytes)
	
	// timestamp 변환
	if lockStart.Valid {
		c.LockStart = timestamppb.New(lockStart.Time)
	}
	if lockEnd.Valid {
		c.LockEnd = timestamppb.New(lockEnd.Time)
	}
	if createdAt.Valid {
		c.CreatedAt = timestamppb.New(createdAt.Time)
	}

	response := &query.GetCampaignResponse{
		Campaign: &c,
		Found:    true,
	}

	log.Printf("Found campaign: %s", c.Address)
	return response, nil
}

func main() {
	// PostgreSQL 연결
	dbURL := "postgres://myuser:mypassword@localhost:5432/mydb?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 연결 테스트
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL database")

	// gRPC 서버 생성
	server := grpc.NewServer()
	queryServer := NewQueryServer(db)
	
	// 서비스 등록
	query.RegisterQueryServiceServer(server, queryServer)

	// 리스너 생성
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("Query server starting on :50051")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
} 