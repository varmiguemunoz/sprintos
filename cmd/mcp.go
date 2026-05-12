package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	sprintosmcp "github.com/varmiguemunoz/command_pm_app/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the SprintOS MCP server for AI integrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := sprintosmcp.StartServer(DB); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
