package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/tray"
	"github.com/varmiguemunoz/sprintos/internal/tui"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Launch the SprintOS terminal UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		go func() {
			_ = tray.EnsureInstalled()
		}()

		if err := tui.Start(DB); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
