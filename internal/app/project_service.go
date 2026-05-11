package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/command_pm_app/internal/domain"
	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) Create(
	name string,
	description string,
	orgID uint,
	createdByID uint,
) (*domain.Project, error) {

	var organization domain.Organization

	if err := s.db.First(&organization, orgID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization with id %d not found", orgID)
		}
		return nil, fmt.Errorf("error checking organization: %w", err)
	}

	var creator domain.User

	if err := s.db.First(&creator, createdByID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("creator with id %d not found", createdByID)
		}
		return nil, fmt.Errorf("error checking creator: %w", err)
	}

	project := domain.Project{
		Name:           name,
		Description:    &description,
		OrganizationID: orgID,
		CreatedByID:    createdByID,
	}

	if err := s.db.Create(&project).Error; err != nil {
		return nil, fmt.Errorf("could not create project: %w", err)
	}

	return &project, nil
}

func (s *ProjectService) Update(
	id uint,
	name string,
	description string,
) (*domain.Project, error) {

	var project domain.Project

	if err := s.db.First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("project with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching project: %w", err)
	}

	updates := map[string]interface{}{
		"name":        name,
		"description": description,
	}

	if err := s.db.Model(&project).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("could not update project: %w", err)
	}

	return &project, nil
}

func (s *ProjectService) Delete(id uint) error {
	var project domain.Project

	if err := s.db.First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("project with id %d not found", id)
		}
		return fmt.Errorf("error fetching project: %w", err)
	}

	if err := s.db.Delete(&project).Error; err != nil {
		return fmt.Errorf("could not delete project: %w", err)
	}

	return nil
}

func (s *ProjectService) GetByID(id uint) (*domain.Project, error) {
	var project domain.Project

	if err := s.db.First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("project with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching project: %w", err)
	}

	return &project, nil
}

func (s *ProjectService) ListByOrganization(orgID uint) ([]domain.Project, error) {
	var projects []domain.Project

	if err := s.db.Where("organization_id = ?", orgID).Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("could not list projects: %w", err)
	}

	return projects, nil
}
