package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "switchy",
	Short: "Switchy is a CLI tool for managing your sdk",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("All available commands:")
		for _, c := range cmd.Commands() {
			fmt.Printf("- %s: %s\n", c.Use, c.Short)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
