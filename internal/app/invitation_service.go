package app

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type InvitationService struct {
	db *gorm.DB
}

func NewInvitationService(db *gorm.DB) *InvitationService {
	return &InvitationService{db: db}
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *InvitationService) Create(email string, orgID uint) (*domain.Invitation, error) {
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	invitation := domain.Invitation{
		Email:          email,
		OrganizationID: orgID,
		Token:          token,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.db.Create(&invitation).Error; err != nil {
		return nil, fmt.Errorf("could not create invitation: %w", err)
	}

	return &invitation, nil
}

func (s *InvitationService) GetByToken(token string) (*domain.Invitation, error) {
	var invitation domain.Invitation

	err := s.db.Preload("Organization").
		Where("token = ?", token).
		First(&invitation).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("invitation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching invitation: %w", err)
	}

	return &invitation, nil
}

func (s *InvitationService) Accept(token string) (*domain.Invitation, error) {
	inv, err := s.GetByToken(token)
	if err != nil {
		return nil, err
	}

	if inv.AcceptedAt != nil {
		return nil, fmt.Errorf("this invitation has already been accepted")
	}

	if time.Now().After(inv.ExpiresAt) {
		return nil, fmt.Errorf("this invitation has expired")
	}

	now := time.Now()
	if err := s.db.Model(inv).Update("accepted_at", now).Error; err != nil {
		return nil, fmt.Errorf("could not accept invitation: %w", err)
	}

	inv.AcceptedAt = &now
	return inv, nil
}
