package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/config"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/db"
	"gorm.io/gorm"
)

var DB *gorm.DB
var migrateDB bool

var rootCmd = &cobra.Command{
	Use:   "sprintos",
	Short: "SprintOS — project manager for your terminal",
	Long:  `A fast, keyboard-driven project and task manager that lives entirely in your terminal.`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		databaseURL := config.GetDatabaseURL()
		if databaseURL == "" {
			return fmt.Errorf("database URL is not configured")
		}

		conn, err := db.Connect(databaseURL, migrateDB)
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
	rootCmd.PersistentFlags().BoolVar(&migrateDB, "migrate", false, "Run database schema migrations before starting")
}
