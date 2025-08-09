package cmd

import (
	"io"
	"log"
	"strings"

	"github.com/jkvn/Switchy/internal/local"
	"github.com/jkvn/Switchy/internal/sdk"
	"github.com/jucardi/go-streams/streams"
	"github.com/schollz/progressbar/v3"
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
			log.Println("Usage:", cmd.Use)
			return
		}
		sdkType := strings.ToLower(args[0])
		version := args[1]

		versions, err := sdk.GetSdks(sdkType)
		if err != nil {
			log.Printf("Error fetching SDKs for %s: %v\n", sdkType, err)
			return
		}
		found := streams.From(versions).AnyMatch(func(i any) bool {
			return i.(sdk.Version).Version == version
		})
		if !found {
			log.Printf("Version %s is not available for %s.\n", version, sdkType)
			return
		}

		sdk.ProgressWriterFactory = func(total int64) io.Writer {
			return progressbar.DefaultBytes(total, sdkType+" "+version)
		}
		defer func() { sdk.ProgressWriterFactory = nil }()

		path, err := sdk.DownloadSdk(sdkType, version)
		if err != nil {
			log.Printf("Download failed: %v\n", err)
			return
		}
		if err := sdk.ExtractSdk(path, sdkType, version); err != nil {
			log.Printf("Extract failed: %v\n", err)
			return
		}
		if err := local.SetSdkVersion(sdkType, version); err != nil {
			log.Printf("Activate failed: %v\n", err)
			return
		}
		log.Printf("Using %s %s\n", sdkType, version)
	},
}
