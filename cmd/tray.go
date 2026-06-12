package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/tray"
)

var trayCmd = &cobra.Command{
	Use:   "tray",
	Short: "Run the SprintOS menu bar companion app",
	Long:  `Starts the SprintOS menu bar icon with task timer. macOS only.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tray.Run(DB)
	},
}

var trayInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the SprintOS tray app as a login item (macOS only)",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := tray.Install(); err != nil {
			return err
		}
		fmt.Println("✓ SprintOS tray installed. It will launch automatically on login.")
		return nil
	},
}

var trayUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the SprintOS tray app login item (macOS only)",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := tray.Uninstall(); err != nil {
			return err
		}
		fmt.Println("✓ SprintOS tray uninstalled.")
		return nil
	},
}

var trayStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether the SprintOS tray app is installed",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if tray.IsInstalled() {
			fmt.Println("SprintOS tray: installed ✓")
		} else {
			fmt.Println("SprintOS tray: not installed")
			fmt.Println("Run 'sprintos tray install' to install it.")
		}
		return nil
	},
}

func init() {
	trayCmd.AddCommand(trayInstallCmd)
	trayCmd.AddCommand(trayUninstallCmd)
	trayCmd.AddCommand(trayStatusCmd)
	rootCmd.AddCommand(trayCmd)
}
