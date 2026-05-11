package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/command_pm_app/internal/domain"
	"gorm.io/gorm"
)

type StateService struct {
	db *gorm.DB
}

// Constructor.
func NewStateService(db *gorm.DB) *StateService {
	return &StateService{db: db}
}

// stateTemplate defines the shape of a single state inside a template.
type stateTemplate struct {
	Name     string
	Color    string
	Position uint
	IsDone   bool
}

// templates is the catalog of available templates.
// To add a new template in the future, just add a new entry here.
var templates = map[string][]stateTemplate{
	"standard": {
		{Name: "Backlog", Color: "#6B7280", Position: 1, IsDone: false},
		{Name: "In Progress", Color: "#3B82F6", Position: 2, IsDone: false},
		{Name: "In Review", Color: "#F59E0B", Position: 3, IsDone: false},
		{Name: "Done", Color: "#10B981", Position: 4, IsDone: true},
	},
}

// ApplyTemplate creates all the states from a named template for a given project.
func (s *StateService) ApplyTemplate(projectID uint, templateName string) error {
	tmpl, ok := templates[templateName]
	if !ok {
		return fmt.Errorf("template %q does not exist", templateName)
	}

	for _, t := range tmpl {
		state := domain.State{
			Name:      t.Name,
			Color:     t.Color,
			Position:  t.Position,
			IsDone:    t.IsDone,
			ProjectID: projectID,
		}
		if err := s.db.Create(&state).Error; err != nil {
			return fmt.Errorf("could not create state %q: %w", t.Name, err)
		}
	}

	return nil
}

// Create creates a single custom state for a project.
func (s *StateService) Create(
	name string,
	color string,
	position uint,
	isDone bool,
	projectID uint,
) (*domain.State, error) {
	state := domain.State{
		Name:      name,
		Color:     color,
		Position:  position,
		IsDone:    isDone,
		ProjectID: projectID,
	}

	if err := s.db.Create(&state).Error; err != nil {
		return nil, fmt.Errorf("could not create state: %w", err)
	}

	return &state, nil
}

// Update updates the name, color, position, and isDone flag of a state.
func (s *StateService) Update(
	id uint,
	name string,
	color string,
	position uint,
	isDone bool,
) (*domain.State, error) {
	var state domain.State

	if err := s.db.First(&state, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("state with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching state: %w", err)
	}

	updates := map[string]interface{}{
		"name":     name,
		"color":    color,
		"position": position,
		"is_done":  isDone,
	}

	if err := s.db.Model(&state).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("could not update state: %w", err)
	}

	return &state, nil
}

// Delete soft-deletes a state by its primary key.
func (s *StateService) Delete(id uint) error {
	var state domain.State

	if err := s.db.First(&state, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("state with id %d not found", id)
		}
		return fmt.Errorf("error fetching state: %w", err)
	}

	if err := s.db.Delete(&state).Error; err != nil {
		return fmt.Errorf("could not delete state: %w", err)
	}

	return nil
}

// GetByID fetches a state by its primary key.
func (s *StateService) GetByID(id uint) (*domain.State, error) {
	var state domain.State

	if err := s.db.First(&state, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("state with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching state: %w", err)
	}

	return &state, nil
}

// ListByProject returns all states for a project, ordered by position.
func (s *StateService) ListByProject(projectID uint) ([]domain.State, error) {
	var states []domain.State

	if err := s.db.Where("project_id = ?", projectID).
		Order("position asc").
		Find(&states).Error; err != nil {
		return nil, fmt.Errorf("could not list states: %w", err)
	}

	return states, nil
}
