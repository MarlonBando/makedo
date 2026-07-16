package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"makedo/internal/engine"
	"makedo/internal/handlers"
)

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Run fenced code blocks from a markdown file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mdPath := strings.TrimSpace(args[0])
		ctx, err := engine.NewRunContext(mdPath)
		if err != nil {
			return err
		}
		defer ctx.Cleanup()
		return handlers.RunMarkdownFile(mdPath, ctx)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
