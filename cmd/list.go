package cmd

import (
	"fmt"
	"strings"

	"github.com/jkvn/Switchy/internal/sdk"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list [sdk]",
	Short:   "List available SDKs or versions",
	Example: "switchy list",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			sdkTypes, err := sdk.GetSdkTypes()
			if err != nil {
				fmt.Println("Error fetching SDK types:", err)
				return
			}

			fmt.Println("Available SDK Types:")
			for _, sdkType := range sdkTypes {
				fmt.Println("-", sdkType)
			}
			return
		}

		sdkType := strings.ToLower(args[0])

		versions, err := sdk.GetSdks(sdkType)
		if err != nil {
			fmt.Printf("Error fetching SDKs for %s: %v\n", sdkType, err)
			return
		}

		fmt.Printf("Available versions for %s:\n", sdkType)
		for _, v := range versions {
			fmt.Printf("- Version %s (%s)\n", v.Version, v.Link)
		}
	},
}
