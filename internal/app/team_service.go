package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/command_pm_app/internal/domain"
	"gorm.io/gorm"
)

type TeamService struct {
	db *gorm.DB
}

// Constructor.
func NewTeamService(db *gorm.DB) *TeamService {
	return &TeamService{db: db}
}

// AddMember adds a user to an organization with a given role.
// Returns an error if the user is already a member.
func (s *TeamService) AddMember(userID, organizationID uint, role string) (*domain.TeamMember, error) {
	// Check the user is not already a member of this organization.
	var existing domain.TeamMember
	result := s.db.Where("user_id = ? AND organization_id = ?", userID, organizationID).First(&existing)
	if result.Error == nil {
		return nil, fmt.Errorf("user %d is already a member of this organization", userID)
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking membership: %w", result.Error)
	}

	member := domain.TeamMember{
		UserID:         userID,
		OrganizationID: organizationID,
		Role:           role,
	}

	if err := s.db.Create(&member).Error; err != nil {
		return nil, fmt.Errorf("could not add member: %w", err)
	}

	return &member, nil
}

// RemoveMember removes a user from an organization.
func (s *TeamService) RemoveMember(userID, organizationID uint) error {
	result := s.db.Where("user_id = ? AND organization_id = ?", userID, organizationID).
		Delete(&domain.TeamMember{})

	if result.Error != nil {
		return fmt.Errorf("could not remove member: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user %d is not a member of this organization", userID)
	}

	return nil
}

// ListMembers returns all members of an organization, including their user data.
// Preload("User") tells GORM to also fetch the related User record for each member.
func (s *TeamService) ListMembers(organizationID uint) ([]domain.TeamMember, error) {
	var members []domain.TeamMember

	if err := s.db.Preload("User").
		Where("organization_id = ?", organizationID).
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("could not list members: %w", err)
	}

	return members, nil
}

// GetMemberRole returns the role of a user within an organization.
func (s *TeamService) GetMemberRole(userID, organizationID uint) (string, error) {
	var member domain.TeamMember

	err := s.db.Where("user_id = ? AND organization_id = ?", userID, organizationID).
		First(&member).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("user %d is not a member of this organization", userID)
	}
	if err != nil {
		return "", fmt.Errorf("error fetching role: %w", err)
	}

	return member.Role, nil
}
