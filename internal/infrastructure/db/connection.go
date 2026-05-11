package db

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
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
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
