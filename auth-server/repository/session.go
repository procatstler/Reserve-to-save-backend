package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"r2s/pkg/database"
	"r2s/pkg/models"
)

type SessionRepository struct {
	db *database.DB
}

func NewSessionRepository(db *database.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *models.Session) error {
	query := `
		INSERT INTO sessions (
			id, user_id, token_hash, refresh_token_hash,
			ip_address, user_agent, device_fingerprint,
			expires_at, refresh_expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`
	
	_, err := r.db.Exec(
		query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.RefreshTokenHash,
		session.IPAddress,
		session.UserAgent,
		session.DeviceFingerprint,
		session.ExpiresAt,
		session.RefreshExpiresAt,
	)
	return err
}

func (r *SessionRepository) FindByToken(tokenHash string) (*models.Session, error) {
	var session models.Session
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash,
		       ip_address, user_agent, device_fingerprint,
		       expires_at, refresh_expires_at, created_at, last_used_at
		FROM sessions 
		WHERE token_hash = $1`
	
	err := r.db.Get(&session, query, tokenHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &session, err
}

func (r *SessionRepository) FindByRefreshToken(refreshTokenHash string) (*models.Session, error) {
	var session models.Session
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash,
		       ip_address, user_agent, device_fingerprint,
		       expires_at, refresh_expires_at, created_at, last_used_at
		FROM sessions 
		WHERE refresh_token_hash = $1`
	
	err := r.db.Get(&session, query, refreshTokenHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &session, err
}

func (r *SessionRepository) Update(session *models.Session) error {
	query := `
		UPDATE sessions 
		SET token_hash = $2, expires_at = $3, last_used_at = $4
		WHERE id = $1`
	
	_, err := r.db.Exec(
		query,
		session.ID,
		session.TokenHash,
		session.ExpiresAt,
		session.LastUsedAt,
	)
	return err
}

func (r *SessionRepository) UpdateLastUsed(id uuid.UUID) error {
	query := `UPDATE sessions SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SessionRepository) DeleteByToken(tokenHash string) error {
	query := `DELETE FROM sessions WHERE token_hash = $1`
	_, err := r.db.Exec(query, tokenHash)
	return err
}

func (r *SessionRepository) DeleteExpired() error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.Exec(query)
	return err
}

func (r *SessionRepository) DeleteByUserID(userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *SessionRepository) DeleteOldSessions(userID uuid.UUID, keepCount int) error {
	query := `
		DELETE FROM sessions 
		WHERE user_id = $1 
		AND id NOT IN (
			SELECT id FROM sessions 
			WHERE user_id = $1 
			ORDER BY created_at DESC 
			LIMIT $2
		)`
	
	_, err := r.db.Exec(query, userID, keepCount)
	return err
}