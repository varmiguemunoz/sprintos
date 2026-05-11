package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/varmiguemunoz/command_pm_app/internal/infrastructure/db"
	"gorm.io/gorm"
)

var DB *gorm.DB

var rootCmd = &cobra.Command{
	Use:   "commandpm",
	Short: "CommandPM — project manager for your terminal",
	Long:  `A fast, keyboard-driven project and task manager that lives entirely in your terminal.`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Read the config.yaml file
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("could not read config file: %w", err)
		}

		// Get the database path from config.yaml
		dbPath := viper.GetString("database.path")
		if dbPath == "" {
			return fmt.Errorf("database.path is not set in config.yaml")
		}

		// Open the database and run AutoMigrate
		conn, err := db.Connect(dbPath)
		if err != nil {
			return fmt.Errorf("could not connect to database: %w", err)
		}

		DB = conn
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	_ = godotenv.Load()
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
}
