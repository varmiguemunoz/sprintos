package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type OrganizationService struct {
	db *gorm.DB
}

// Constructor.
func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

func (s *OrganizationService) Create(
	name string,
	description string,
	whatsappNumber string,
	ownerID uint,
) (*domain.Organization, error) {

	var owner domain.User
	if err := s.db.First(&owner, ownerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("owner with id %d does not exist", ownerID)
		}
		return nil, fmt.Errorf("error checking owner: %w", err)
	}

	org := domain.Organization{
		Name:           name,
		Description:    &description,
		WhatsappNumber: whatsappNumber,
		OwnerID:        ownerID,
	}

	if err := s.db.Create(&org).Error; err != nil {
		return nil, fmt.Errorf("could not create organization: %w", err)
	}

	return &org, nil
}

func (s *OrganizationService) Update(
	id uint,
	name string,
	description string,
	whatsappNumber string,
) (*domain.Organization, error) {

	var org domain.Organization

	if err := s.db.First(&org, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching organization: %w", err)
	}

	updates := map[string]interface{}{
		"name":            name,
		"description":     description,
		"whatsapp_number": whatsappNumber,
	}

	if err := s.db.Model(&org).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("could not update organization: %w", err)
	}

	return &org, nil
}

func (s *OrganizationService) GetByID(id uint) (*domain.Organization, error) {
	var org domain.Organization

	if err := s.db.First(&org, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching organization: %w", err)
	}

	return &org, nil
}

func (s *OrganizationService) GetByOwnerID(ownerID uint) (*domain.Organization, error) {
	var org domain.Organization

	if err := s.db.Where("owner_id = ?", ownerID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no organization found for owner id %d", ownerID)
		}
		return nil, fmt.Errorf("error fetching organization: %w", err)
	}

	return &org, nil
}

func (s *OrganizationService) Delete(id uint) error {
	var org domain.Organization

	if err := s.db.First(&org, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("organization with id %d not found", id)
		}
		return fmt.Errorf("error fetching organization: %w", err)
	}

	if err := s.db.Delete(&org).Error; err != nil {
		return fmt.Errorf("could not delete organization: %w", err)
	}

	return nil
}
