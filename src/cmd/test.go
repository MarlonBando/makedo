package cmd

import (
	"fmt"
	"os"

	"makedo/internal/engine"
	"makedo/internal/handlers"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [file]",
	Short: "Test code blocks in your markdown file",
	Long:  `Run code blocks followed by html comment in the format of '<!-- directive <content> -->'  verify their output matches the expected pattern.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := engine.NewRunContext()
		defer ctx.Cleanup()
		err := handlers.VerifyMarkdown(args[0], ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
