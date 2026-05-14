package app

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type APIKeyService struct {
	db *gorm.DB
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk_" + hex.EncodeToString(bytes), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func (s *APIKeyService) Create(name string, userID, orgID uint) (string, *domain.APIKey, error) {
	raw, err := generateAPIKey()
	if err != nil {
		return "", nil, fmt.Errorf("could not generate key: %w", err)
	}

	key := domain.APIKey{
		Name:      name,
		KeyHash:   hashKey(raw),
		KeyPrefix: raw[:10] + "...",
		UserID:    userID,
		OrgID:     orgID,
	}

	if err := s.db.Create(&key).Error; err != nil {
		return "", nil, fmt.Errorf("could not save key: %w", err)
	}

	return raw, &key, nil
}

func (s *APIKeyService) ValidateKey(raw string) (*domain.APIKey, error) {
	hash := hashKey(raw)

	var key domain.APIKey
	err := s.db.Preload("User").Where("key_hash = ? AND revoked_at IS NULL", hash).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("invalid API key")
	}
	if err != nil {
		return nil, fmt.Errorf("error validating key: %w", err)
	}

	now := time.Now()
	s.db.Model(&key).Update("last_used_at", now)

	return &key, nil
}

func (s *APIKeyService) ListByOrg(orgID uint) ([]domain.APIKey, error) {
	var keys []domain.APIKey
	if err := s.db.Where("org_id = ? AND revoked_at IS NULL", orgID).Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("could not list keys: %w", err)
	}
	return keys, nil
}

func (s *APIKeyService) Revoke(id uint, orgID uint) error {
	now := time.Now()
	result := s.db.Model(&domain.APIKey{}).
		Where("id = ? AND org_id = ?", id, orgID).
		Update("revoked_at", now)
	if result.Error != nil {
		return fmt.Errorf("could not revoke key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("key not found")
	}
	return nil
}
