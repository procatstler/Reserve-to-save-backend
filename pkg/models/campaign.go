package models

import (
	"database/sql/driver"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type CampaignStatus string

const (
	StatusDraft       CampaignStatus = "draft"
	StatusRecruiting  CampaignStatus = "recruiting"
	StatusReached     CampaignStatus = "reached"
	StatusFulfillment CampaignStatus = "fulfillment"
	StatusSettled     CampaignStatus = "settled"
	StatusFailed      CampaignStatus = "failed"
	StatusCancelled   CampaignStatus = "cancelled"
)

type Campaign struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	ChainAddress   string          `json:"chain_address" db:"chain_address"`
	Title          string          `json:"title" db:"title"`
	Description    *string         `json:"description,omitempty" db:"description"`
	ImageURL       *string         `json:"image_url,omitempty" db:"image_url"`
	MerchantID     *uuid.UUID      `json:"merchant_id,omitempty" db:"merchant_id"`
	MerchantWallet string          `json:"merchant_wallet" db:"merchant_wallet"`
	BasePrice      *big.Int        `json:"base_price" db:"base_price"`
	MinQty         int             `json:"min_qty" db:"min_qty"`
	CurrentQty     int             `json:"current_qty" db:"current_qty"`
	TargetAmount   *big.Int        `json:"target_amount" db:"target_amount"`
	CurrentAmount  *big.Int        `json:"current_amount" db:"current_amount"`
	DiscountRate   int             `json:"discount_rate" db:"discount_rate"`
	SaveFloorBps   int             `json:"save_floor_bps" db:"save_floor_bps"`
	RMaxBps        int             `json:"r_max_bps" db:"r_max_bps"`
	MerchantFeeBps int             `json:"merchant_fee_bps" db:"merchant_fee_bps"`
	OpsFeeBps      int             `json:"ops_fee_bps" db:"ops_fee_bps"`
	StartTime      time.Time       `json:"start_time" db:"start_time"`
	EndTime        time.Time       `json:"end_time" db:"end_time"`
	SettlementDate *time.Time      `json:"settlement_date,omitempty" db:"settlement_date"`
	Status         CampaignStatus  `json:"status" db:"status"`
	TxHash         *string         `json:"tx_hash,omitempty" db:"tx_hash"`
	BlockNumber    *int64          `json:"block_number,omitempty" db:"block_number"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
}

type Participation struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	CampaignID        uuid.UUID  `json:"campaign_id" db:"campaign_id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	WalletAddress     string     `json:"wallet_address" db:"wallet_address"`
	DepositAmount     *big.Int   `json:"deposit_amount" db:"deposit_amount"`
	JoinedAt          time.Time  `json:"joined_at" db:"joined_at"`
	CancelPending     *big.Int   `json:"cancel_pending" db:"cancel_pending"`
	ExpectedRebate    *big.Int   `json:"expected_rebate" db:"expected_rebate"`
	ActualRebate      *big.Int   `json:"actual_rebate,omitempty" db:"actual_rebate"`
	Status            string     `json:"status" db:"status"`
	TxHash            *string    `json:"tx_hash,omitempty" db:"tx_hash"`
	CancelTxHash      *string    `json:"cancel_tx_hash,omitempty" db:"cancel_tx_hash"`
	SettlementTxHash  *string    `json:"settlement_tx_hash,omitempty" db:"settlement_tx_hash"`
	RefundTxHash      *string    `json:"refund_tx_hash,omitempty" db:"refund_tx_hash"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
}

// BigInt is a wrapper for big.Int to handle database operations
type BigInt struct {
	*big.Int
}

func (b *BigInt) Scan(value interface{}) error {
	if value == nil {
		b.Int = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		b.Int = new(big.Int)
		b.Int.SetString(string(v), 10)
	case string:
		b.Int = new(big.Int)
		b.Int.SetString(v, 10)
	}
	return nil
}

func (b BigInt) Value() (driver.Value, error) {
	if b.Int == nil {
		return nil, nil
	}
	return b.String(), nil
}