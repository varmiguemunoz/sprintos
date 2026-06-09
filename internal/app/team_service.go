package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type TeamService struct {
	db *gorm.DB
}

func NewTeamService(db *gorm.DB) *TeamService {
	return &TeamService{db: db}
}

func (s *TeamService) AddMember(userID, organizationID uint, role string) (*domain.TeamMember, error) {
	var existing domain.TeamMember
	result := s.db.Where("user_id = ? AND organization_id = ?", userID, organizationID).First(&existing)
	if result.Error == nil {
		return &existing, nil
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking membership: %w", result.Error)
	}

	if role == "" {
		role = domain.RoleMember
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

func (s *TeamService) RemoveMemberByID(memberID uint) error {
	result := s.db.Delete(&domain.TeamMember{}, memberID)

	if result.Error != nil {
		return fmt.Errorf("could not remove member: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("member with id %d not found", memberID)
	}

	return nil
}

func (s *TeamService) ListMembers(organizationID uint) ([]domain.TeamMember, error) {
	var members []domain.TeamMember

	if err := s.db.Preload("User").
		Where("organization_id = ?", organizationID).
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("could not list members: %w", err)
	}

	return members, nil
}

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

func (s *TeamService) UpdateMemberRole(userID, organizationID uint, newRole string) error {
	result := s.db.Model(&domain.TeamMember{}).
		Where("user_id = ? AND organization_id = ?", userID, organizationID).
		Update("role", newRole)

	if result.Error != nil {
		return fmt.Errorf("could not update member role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user %d is not a member of this organization", userID)
	}

	return nil
}

func (s *TeamService) GetOrganizationsByMemberUserID(userID uint) ([]domain.Organization, error) {
	var orgs []domain.Organization

	err := s.db.
		Joins("JOIN team_members ON team_members.organization_id = organizations.id AND team_members.deleted_at IS NULL").
		Where("team_members.user_id = ? AND organizations.deleted_at IS NULL AND organizations.owner_id != ?", userID, userID).
		Find(&orgs).Error

	if err != nil {
		return nil, fmt.Errorf("could not fetch member organizations: %w", err)
	}

	return orgs, nil
}
