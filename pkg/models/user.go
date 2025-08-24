package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	WalletAddress   string         `json:"wallet_address" db:"wallet_address"`
	LineUserID      *string        `json:"line_user_id,omitempty" db:"line_user_id"`
	LineDisplayName *string        `json:"line_display_name,omitempty" db:"line_display_name"`
	LinePictureURL  *string        `json:"line_picture_url,omitempty" db:"line_picture_url"`
	Email           *string        `json:"email,omitempty" db:"email"`
	KYCTier         int            `json:"kyc_tier" db:"kyc_tier"`
	Status          string         `json:"status" db:"status"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
	LastLoginAt     *time.Time     `json:"last_login_at,omitempty" db:"last_login_at"`
	Metadata        pq.StringArray `json:"metadata" db:"metadata"`
}

type Session struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash         string     `json:"token_hash" db:"token_hash"`
	RefreshTokenHash  *string    `json:"refresh_token_hash,omitempty" db:"refresh_token_hash"`
	IPAddress         *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent         *string    `json:"user_agent,omitempty" db:"user_agent"`
	DeviceFingerprint *string    `json:"device_fingerprint,omitempty" db:"device_fingerprint"`
	ExpiresAt         time.Time  `json:"expires_at" db:"expires_at"`
	RefreshExpiresAt  *time.Time `json:"refresh_expires_at,omitempty" db:"refresh_expires_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt        time.Time  `json:"last_used_at" db:"last_used_at"`
}