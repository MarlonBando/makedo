package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"makedo/internal/version"
)

var showVersion bool

var rootCmd = &cobra.Command{
	Use:   "makedo",
	Short: "A markdown-based task runner",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Println(version.Get())
			return
		}
		fmt.Println(version.Get())
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
