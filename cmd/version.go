package cmd

import (
	"fmt"

	"github.com/jkvn/Switchy/internal/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Switchy version:", core.Version, "Commit:", core.Commit)
	},
}
