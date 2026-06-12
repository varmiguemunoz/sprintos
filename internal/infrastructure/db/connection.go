package db

import (
	"fmt"
	"time"

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
	m, err := newMigrator(database)
	if err != nil {
		return nil, fmt.Errorf("failed to init migrator: %w", err)
	}
	if err := m.Run(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	return database, nil
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
