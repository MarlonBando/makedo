package cmd

import (
	"github.com/spf13/cobra"
	"makedo/internal/handlers"
)

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Run fenced code blocks from a markdown file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlers.RunMarkdownFile(args[0])
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
