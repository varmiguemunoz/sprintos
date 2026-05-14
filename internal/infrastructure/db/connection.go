package db

import (
	"fmt"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(
		&domain.User{},
		&domain.Organization{},
		&domain.TeamMember{},
		&domain.Project{},
		&domain.State{},
		&domain.Task{},
		&domain.Comment{},
		&domain.Invitation{},
		&domain.GitHubIntegration{},
		&domain.APIKey{},
		&domain.OutboundWebhook{},
		&domain.Sprint{},
		&domain.BurndownSnapshot{},
		&domain.StateTransition{},
		&domain.NotificationConfig{},
		&domain.NotificationPreference{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
