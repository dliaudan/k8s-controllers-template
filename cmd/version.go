package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", appVersion)
		fmt.Printf("Build Date: %s\n", buildDate)
		fmt.Printf("Commit SHA: %s\n", commitSHA)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
