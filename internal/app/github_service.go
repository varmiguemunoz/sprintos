package app

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type GitHubService struct {
	db *gorm.DB
}

func NewGitHubService(db *gorm.DB) *GitHubService {
	return &GitHubService{db: db}
}

func generateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *GitHubService) CreateIntegration(
	orgID uint,
	repoOwner string,
	repoName string,
	projectID uint,
	inReviewStateID uint,
	doneStateID uint,
) (*domain.GitHubIntegration, error) {
	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, fmt.Errorf("could not generate webhook secret: %w", err)
	}

	integration := domain.GitHubIntegration{
		OrganizationID:  orgID,
		RepoOwner:       repoOwner,
		RepoName:        repoName,
		ProjectID:       projectID,
		WebhookSecret:   secret,
		InReviewStateID: inReviewStateID,
		DoneStateID:     doneStateID,
	}

	if err := s.db.Create(&integration).Error; err != nil {
		return nil, fmt.Errorf("could not create integration: %w", err)
	}

	return &integration, nil
}

func (s *GitHubService) GetIntegrationByRepo(repoOwner, repoName string) (*domain.GitHubIntegration, error) {
	var integration domain.GitHubIntegration

	err := s.db.
		Preload("Organization").
		Preload("Project").
		Where("repo_owner = ? AND repo_name = ?", repoOwner, repoName).
		First(&integration).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no integration found for %s/%s", repoOwner, repoName)
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching integration: %w", err)
	}

	return &integration, nil
}

func (s *GitHubService) ListByOrg(orgID uint) ([]domain.GitHubIntegration, error) {
	var integrations []domain.GitHubIntegration

	if err := s.db.Where("organization_id = ?", orgID).Find(&integrations).Error; err != nil {
		return nil, fmt.Errorf("could not list integrations: %w", err)
	}

	return integrations, nil
}

func (s *GitHubService) DeleteIntegration(id uint) error {
	if err := s.db.Delete(&domain.GitHubIntegration{}, id).Error; err != nil {
		return fmt.Errorf("could not delete integration: %w", err)
	}
	return nil
}

func (s *GitHubService) GetNextTaskNumber(orgID uint) (int, error) {
	var max int
	err := s.db.Raw(`
		SELECT COALESCE(MAX(t.task_number), 0)
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		WHERE p.organization_id = ?
		AND t.deleted_at IS NULL
	`, orgID).Scan(&max).Error

	if err != nil {
		return 0, fmt.Errorf("could not get next task number: %w", err)
	}

	return max + 1, nil
}
