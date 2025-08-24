package models

import (
	"math/big"
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string
type PaymentMode string
type Currency string

const (
	PaymentPending    PaymentStatus = "pending"
	PaymentProcessing PaymentStatus = "processing"
	PaymentCompleted  PaymentStatus = "completed"
	PaymentFailed     PaymentStatus = "failed"
	PaymentRefunded   PaymentStatus = "refunded"

	ModeCrypto PaymentMode = "crypto"
	ModeStripe PaymentMode = "stripe"

	CurrencyUSDT Currency = "USDT"
	CurrencyKAIA Currency = "KAIA"
	CurrencyKRW  Currency = "KRW"
	CurrencyUSD  Currency = "USD"
)

type Payment struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	PaymentID        string                 `json:"payment_id" db:"payment_id"`
	CampaignID       *uuid.UUID             `json:"campaign_id,omitempty" db:"campaign_id"`
	UserID           *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	ParticipationID  *uuid.UUID             `json:"participation_id,omitempty" db:"participation_id"`
	Amount           *big.Int               `json:"amount" db:"amount"`
	Currency         Currency               `json:"currency" db:"currency"`
	Mode             PaymentMode            `json:"mode" db:"mode"`
	Status           PaymentStatus          `json:"status" db:"status"`
	TransactionHash  *string                `json:"transaction_hash,omitempty" db:"transaction_hash"`
	ProviderResponse map[string]interface{} `json:"provider_response,omitempty" db:"provider_response"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	FailedAt         *time.Time             `json:"failed_at,omitempty" db:"failed_at"`
	RefundedAt       *time.Time             `json:"refunded_at,omitempty" db:"refunded_at"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
}

type WebhookLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	EventID      string                 `json:"event_id" db:"event_id"`
	EventType    string                 `json:"event_type" db:"event_type"`
	Payload      map[string]interface{} `json:"payload" db:"payload"`
	Signature    *string                `json:"signature,omitempty" db:"signature"`
	Processed    bool                   `json:"processed" db:"processed"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	ErrorMessage *string                `json:"error_message,omitempty" db:"error_message"`
	ReceivedAt   time.Time              `json:"received_at" db:"received_at"`
	ProcessedAt  *time.Time             `json:"processed_at,omitempty" db:"processed_at"`
}