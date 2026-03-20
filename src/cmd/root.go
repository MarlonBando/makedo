package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makedo/internal/version"
)

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
