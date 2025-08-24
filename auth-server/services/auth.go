package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"r2s/auth-server/repository"
	"r2s/pkg/database"
	"r2s/pkg/models"
	"r2s/pkg/utils"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	redis       *database.RedisClient
	jwtManager  *utils.JWTManager
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	redis *database.RedisClient,
	jwtManager *utils.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		redis:       redis,
		jwtManager:  jwtManager,
	}
}

// GenerateNonce generates a nonce for wallet authentication
func (s *AuthService) GenerateNonce(address, chainID string) (string, string, string, string, error) {
	// Validate address
	if !utils.IsValidAddress(address) {
		return "", "", "", "", errors.New("invalid wallet address")
	}

	// Generate nonce
	nonce := utils.GenerateNonce()
	requestID := uuid.New().String()
	issuedAt := time.Now().Format(time.RFC3339)
	expiresAt := time.Now().Add(6 * time.Minute).Format(time.RFC3339)

	// Create message
	domain := "https://r2s.io"
	message := utils.CreateSignMessage(domain, address, chainID, nonce, issuedAt, expiresAt, requestID)

	// Store nonce in Redis
	nonceHash := utils.HashString(nonce)
	nonceData := map[string]string{
		"address":   strings.ToLower(address),
		"chainId":   chainID,
		"requestId": requestID,
		"expiresAt": expiresAt,
	}
	
	nonceJSON, _ := json.Marshal(nonceData)
	if err := s.redis.SetWithExpiry("nonce:"+nonceHash, string(nonceJSON), 6*time.Minute); err != nil {
		return "", "", "", "", fmt.Errorf("failed to store nonce: %w", err)
	}

	return nonce, message, requestID, expiresAt, nil
}

// VerifySignature verifies wallet signature and issues JWT
func (s *AuthService) VerifySignature(address, signature, message, requestID, ipAddress, userAgent string) (*Tokens, *models.User, error) {
	// Extract nonce from message
	nonceRegex := regexp.MustCompile(`Nonce: ([a-f0-9]{32})`)
	matches := nonceRegex.FindStringSubmatch(message)
	if len(matches) != 2 {
		return nil, nil, errors.New("invalid message format")
	}
	nonce := matches[1]

	// Get nonce data from Redis
	nonceHash := utils.HashString(nonce)
	nonceDataStr, err := s.redis.GetString("nonce:" + nonceHash)
	if err != nil {
		return nil, nil, errors.New("invalid or expired nonce")
	}

	var nonceData map[string]string
	if err := json.Unmarshal([]byte(nonceDataStr), &nonceData); err != nil {
		return nil, nil, errors.New("invalid nonce data")
	}

	// Validate nonce data
	if strings.ToLower(nonceData["address"]) != strings.ToLower(address) {
		return nil, nil, errors.New("address mismatch")
	}

	expiresAt, _ := time.Parse(time.RFC3339, nonceData["expiresAt"])
	if time.Now().After(expiresAt) {
		return nil, nil, errors.New("nonce expired")
	}

	// Verify signature
	valid, err := utils.VerifySignature(message, signature, address)
	if err != nil || !valid {
		return nil, nil, errors.New("invalid signature")
	}

	// Delete nonce (one-time use)
	s.redis.Del("nonce:" + nonceHash)

	// Get or create user
	user, err := s.userRepo.FindByWalletAddress(strings.ToLower(address))
	if err != nil {
		// Create new user
		user = &models.User{
			ID:            uuid.New(),
			WalletAddress: strings.ToLower(address),
			KYCTier:       0,
			Status:        "active",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update last login
		s.userRepo.UpdateLastLogin(user.ID)
	}

	// Generate tokens
	sessionID := uuid.New()
	claims := &utils.JWTClaims{
		UserID:    user.ID,
		Address:   user.WalletAddress,
		KYCTier:   user.KYCTier,
		SessionID: sessionID,
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(claims)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.WalletAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := &models.Session{
		ID:               sessionID,
		UserID:           user.ID,
		TokenHash:        utils.HashString(accessToken),
		RefreshTokenHash: stringPtr(utils.HashString(refreshToken)),
		IPAddress:        &ipAddress,
		UserAgent:        &userAgent,
		ExpiresAt:        time.Now().Add(15 * time.Minute),
		RefreshExpiresAt: timePtr(time.Now().Add(7 * 24 * time.Hour)),
		CreatedAt:        time.Now(),
		LastUsedAt:       time.Now(),
	}
	
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, user, nil
}

// LineAuth handles LINE authentication
func (s *AuthService) LineAuth(idToken, accessToken, ipAddress, userAgent string) (string, *models.User, error) {
	// TODO: Implement LINE token verification
	// This would involve calling LINE API to verify the tokens
	// For now, returning an error
	return "", nil, errors.New("LINE authentication not implemented")
}

// RefreshToken generates a new access token from refresh token
func (s *AuthService) RefreshToken(refreshToken string) (string, error) {
	// Verify refresh token
	claims, err := s.jwtManager.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// Get session
	refreshTokenHash := utils.HashString(refreshToken)
	session, err := s.sessionRepo.FindByRefreshToken(refreshTokenHash)
	if err != nil || session.UserID != claims.UserID {
		return "", errors.New("invalid session")
	}

	// Get user
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Generate new access token
	newClaims := &utils.JWTClaims{
		UserID:    user.ID,
		Address:   user.WalletAddress,
		KYCTier:   user.KYCTier,
		SessionID: session.ID,
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(newClaims)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Update session
	session.TokenHash = utils.HashString(accessToken)
	session.ExpiresAt = time.Now().Add(15 * time.Minute)
	session.LastUsedAt = time.Now()
	
	if err := s.sessionRepo.Update(session); err != nil {
		return "", fmt.Errorf("failed to update session: %w", err)
	}

	return accessToken, nil
}

// Logout invalidates the current session
func (s *AuthService) Logout(token string) error {
	tokenHash := utils.HashString(token)
	
	// Delete session
	if err := s.sessionRepo.DeleteByToken(tokenHash); err != nil {
		return err
	}

	// Add token to blacklist
	claims, _ := s.jwtManager.VerifyAccessToken(token)
	if claims != nil {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining > 0 {
			s.redis.SetWithExpiry("blacklist:"+tokenHash, "1", remaining)
		}
	}

	return nil
}

// ValidateToken validates and returns token claims
func (s *AuthService) ValidateToken(token string) (*utils.JWTClaims, error) {
	// Check blacklist
	tokenHash := utils.HashString(token)
	blacklisted, _ := s.redis.Exists("blacklist:" + tokenHash)
	if blacklisted {
		return nil, errors.New("token has been revoked")
	}

	// Verify token
	claims, err := s.jwtManager.VerifyAccessToken(token)
	if err != nil {
		return nil, err
	}

	// Check session
	session, err := s.sessionRepo.FindByToken(tokenHash)
	if err != nil || session.UserID != claims.UserID {
		return nil, errors.New("invalid session")
	}

	// Check expiry
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	// Update last used
	go s.sessionRepo.UpdateLastUsed(session.ID)

	return claims, nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}