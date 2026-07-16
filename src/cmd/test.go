package cmd

import (
	"github.com/spf13/cobra"
	"makedo/internal/engine"
	"makedo/internal/handlers"
	"strings"
)

var testCmd = &cobra.Command{
	Use:   "test [file]",
	Short: "Test code blocks in your markdown file",
	Long:  `Run code blocks followed by html comment in the format of '<!-- directive <content> -->'  verify their output matches the expected pattern.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mdPath := strings.TrimSpace(args[0])
		ctx, err := engine.NewRunContext(mdPath)
		if err != nil {
			return err
		}

		defer ctx.Cleanup()
		return handlers.VerifyMarkdown(mdPath, ctx)
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
