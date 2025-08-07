package cmd

import (
	"log"
	"strings"

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

		sdkType := strings.ToLower(args[0])

		versions, err := sdk.GetSdks(sdkType)
		if err != nil {
			log.Printf("Error fetching SDKs for %s: %v\n", sdkType, err)
			return
		}

		if len(args) < 2 {
			log.Printf("No version provided! Available versions for %s:\n", sdkType)
			for _, v := range versions {
				log.Printf("- Version %s (%s)\n", v.Version, v.Link)
			}
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
		
		fileName, err := sdk.DownloadSdk(sdkType, version)
		if err != nil {
			log.Printf("Error downloading SDK %s version %s: %v\n", sdkType, version, err)
			return
		}

		err = sdk.ExtractSdk(fileName)
		if err != nil {
			log.Printf("Error extract SDK %s version %s: %v\n", sdkType, version, err)
			return
		}
	},
}
