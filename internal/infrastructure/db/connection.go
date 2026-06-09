package db

import (
	"fmt"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	database, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := configurePool(database); err != nil {
		return nil, err
	}

	if err := Migrate(database); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

func Migrate(database *gorm.DB) error {
	return database.AutoMigrate(
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
		&domain.Subtask{},
		&domain.SubtaskComment{},
		&domain.TimeEntry{},
		&domain.ActiveTimer{},
	)
}

func configurePool(database *gorm.DB) error {
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)
	return nil
}

