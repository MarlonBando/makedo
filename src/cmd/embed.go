package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"makedo/internal/engine"
	"makedo/internal/handlers"
)

var embedCmd = &cobra.Command{
	Use:   "embed <file>",
	Short: "Run fenced code blocks and embed output into the markdown file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mdPath := strings.TrimSpace(args[0])
		ctx, err := engine.NewRunContext(mdPath)
		if err != nil {
			return err
		}
		defer ctx.Cleanup()
		return handlers.EmbedMarkdownFile(mdPath, ctx)
	},
}

func init() {
	rootCmd.AddCommand(embedCmd)
}
