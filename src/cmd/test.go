package cmd

import (
	"makedo/internal/engine"
	"makedo/internal/handlers"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [file]",
	Short: "Test code blocks in your markdown file",
	Long:  `Run code blocks followed by html comment in the format of '<!-- directive <content> -->'  verify their output matches the expected pattern.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := engine.NewRunContext()
		if err != nil {
			return err
		}
		ctx.CheckURLs, _ = cmd.Flags().GetBool("check-urls")

		defer ctx.Cleanup()
		return handlers.VerifyMarkdown(args[0], ctx)
	},
}

func init() {
	testCmd.Flags().Bool("check-urls", false, "GET every http(s) link in the document and fail if any do not return 2xx")
	rootCmd.AddCommand(testCmd)
}
