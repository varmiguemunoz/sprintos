package cmd

import (
	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/tray"
)

var trayCmd = &cobra.Command{
	Use:   "tray",
	Short: "Run the SprintOS menu bar companion app",
	Long:  `Starts the SprintOS system tray icon with Pomodoro timer and task time tracking.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tray.Run(DB)
	},
}

func init() {
	rootCmd.AddCommand(trayCmd)
}
