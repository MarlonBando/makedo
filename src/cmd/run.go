package cmd

import (
	"github.com/spf13/cobra"
	"makedo/internal/engine"
	"makedo/internal/handlers"
)

var runCmd = &cobra.Command{
	Use:   "run <file>",
	Short: "Run fenced code blocks from a markdown file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := engine.NewRunContext()
		if err != nil {
			return err
		}
		defer ctx.Cleanup()
		return handlers.RunMarkdownFile(args[0], ctx)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
