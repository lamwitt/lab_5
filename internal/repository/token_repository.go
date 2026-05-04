package repository

import (
	"books-api/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(token *models.Token) error {
	return r.db.Create(token).Error
}

func (r *TokenRepository) FindByHash(hash string) (*models.Token, error) {
	var token models.Token
	if err := r.db.Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *TokenRepository) RevokeByID(id uuid.UUID) error {
	return r.db.Model(&models.Token{}).Where("id = ?", id).Update("revoked", true).Error
}

func (r *TokenRepository) RevokeByHash(hash string) error {
	return r.db.Model(&models.Token{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

func (r *TokenRepository) RevokeAllByUser(userID uuid.UUID) error {
	return r.db.Model(&models.Token{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}
