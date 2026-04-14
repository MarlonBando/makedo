package cmd

import (
	"github.com/spf13/cobra"
	"makedo/internal/handlers"
)

var embedCmd = &cobra.Command{
	Use:   "embed <file>",
	Short: "Run fenced code blocks and embed output into the markdown file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handlers.EmbedMarkdownFile(args[0])
	},
}

func init() {
	rootCmd.AddCommand(embedCmd)
}
