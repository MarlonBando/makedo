package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makedo/internal/version"
)

//TODO: add version flag where we get only the version in return

var rootCmd = &cobra.Command{
	Use:   "makedo",
	Short: "A markdown-based task runner",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Get())
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
