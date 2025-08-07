package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List available SDKs",
	Example: "switchy list",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available SDKs:")
		// Waiting for implementation
	},
}
