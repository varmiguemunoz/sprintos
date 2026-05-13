package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/tui"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Launch the CommandPM terminal UI",
	RunE: func(cmd *cobra.Command, args []string) error {
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
