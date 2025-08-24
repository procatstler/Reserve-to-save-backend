package repository

import (
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"r2s/pkg/database"
	"r2s/pkg/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, wallet_address, line_user_id, line_display_name, 
		       line_picture_url, email, kyc_tier, status, 
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE id = $1`
	
	err := r.db.Get(&user, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByWalletAddress(address string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, wallet_address, line_user_id, line_display_name, 
		       line_picture_url, email, kyc_tier, status, 
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE LOWER(wallet_address) = LOWER($1)`
	
	err := r.db.Get(&user, query, address)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByLineUserID(lineUserID string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, wallet_address, line_user_id, line_display_name, 
		       line_picture_url, email, kyc_tier, status, 
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE line_user_id = $1`
	
	err := r.db.Get(&user, query, lineUserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (
			id, wallet_address, line_user_id, line_display_name, 
			line_picture_url, email, kyc_tier, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)`
	
	_, err := r.db.Exec(
		query,
		user.ID,
		strings.ToLower(user.WalletAddress),
		user.LineUserID,
		user.LineDisplayName,
		user.LinePictureURL,
		user.Email,
		user.KYCTier,
		user.Status,
	)
	return err
}

func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET line_user_id = $2, line_display_name = $3, line_picture_url = $4,
		    email = $5, kyc_tier = $6, status = $7, updated_at = NOW()
		WHERE id = $1`
	
	_, err := r.db.Exec(
		query,
		user.ID,
		user.LineUserID,
		user.LineDisplayName,
		user.LinePictureURL,
		user.Email,
		user.KYCTier,
		user.Status,
	)
	return err
}

func (r *UserRepository) UpdateLastLogin(id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *UserRepository) UpdateLineProfile(id uuid.UUID, displayName, pictureURL string) error {
	query := `
		UPDATE users 
		SET line_display_name = $2, line_picture_url = $3, updated_at = NOW()
		WHERE id = $1`
	
	_, err := r.db.Exec(query, id, displayName, pictureURL)
	return err
}