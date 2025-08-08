package cmd

import (
	"log"
	"strings"

	"github.com/jkvn/Switchy/internal/local"
	"github.com/jkvn/Switchy/internal/sdk"
	"github.com/jucardi/go-streams/streams"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(useCmd)
}

var useCmd = &cobra.Command{
	Use:     "use [sdkType] [version]",
	Short:   "Use a specific SDK version",
	Example: "switchy use java 21.0.6",
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Println("Usage: switchy use [sdkType] [version]")
			return
		}

		sdkType := strings.ToLower(args[0])

		versions, err := sdk.GetSdks(sdkType)
		if err != nil {
			log.Printf("Error fetching SDKs for %s: %v\n", sdkType, err)
			return
		}

		version := args[1]
		found := streams.From(versions).AnyMatch(func(i any) bool {
			return i.(sdk.Version).Version == version
		})

		if !found {
			log.Printf("Version %s is not available for %s.\n", version, sdkType)
			return
		}

		local.SetSdkVersion(sdkType, version)
	},
}
