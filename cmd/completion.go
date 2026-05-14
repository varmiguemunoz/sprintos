package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `Add shell completion by running one of:

  # Bash
  sprintos completion bash > /usr/local/etc/bash_completion.d/sprintos

  # Zsh
  sprintos completion zsh > "${fpath[1]}/_sprintos"

  # Fish
  sprintos completion fish > ~/.config/fish/completions/sprintos.fish`,
	ValidArgs: []string{"bash", "zsh"},
	Args:      cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		default:
			return cmd.Help()
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
