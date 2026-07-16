package cmd

import (
	"github.com/spf13/cobra"
	"makedo/internal/engine"
	"makedo/internal/handlers"
)

var embedCmd = &cobra.Command{
	Use:   "embed <file>",
	Short: "Run fenced code blocks and embed output into the markdown file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := engine.NewRunContext(args[0])
		if err != nil {
			return err
		}
		defer ctx.Cleanup()
		return handlers.EmbedMarkdownFile(args[0], ctx)
	},
}

func init() {
	rootCmd.AddCommand(embedCmd)
}
