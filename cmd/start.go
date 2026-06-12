package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/config"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/db"
	"github.com/varmiguemunoz/sprintos/internal/tui"
	"gorm.io/gorm"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Launch the SprintOS terminal UI",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		dbURL := config.GetDatabaseURL()
		if dbURL == "" {
			fmt.Fprintln(os.Stderr, "database URL is not configured")
			return fmt.Errorf("database URL is not configured")
		}

		dbChan := make(chan *gorm.DB, 1)
		dbErrChan := make(chan error, 1)

		go func() {
			conn, err := db.Connect(dbURL)
			if err != nil {
				dbErrChan <- err
				return
			}
			dbChan <- conn
		}()

		if err := tui.Start(dbChan, dbErrChan); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
