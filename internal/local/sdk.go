package local

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/jkvn/Switchy/internal/config"
	"github.com/jkvn/Switchy/internal/sdk"
)

func SetSdkVersion(sdkType, version string) error {
	if isDownloaded(sdkType, version) {
		log.Printf("SDK %s version %s is already downloaded.\n", sdkType, version)
	} else {

		fileName, err := sdk.DownloadSdk(sdkType, version)
		if err != nil {
			log.Printf("Error downloading SDK %s version %s: %v\n", sdkType, version, err)
			return nil
		}

		err = sdk.ExtractSdk(fileName, sdkType, version)
		if err != nil {
			log.Printf("Error extract SDK %s version %s: %v\n", sdkType, version, err)
			return nil
		}
	}

	copySdkToDefaultPath(sdkType, version)
	return nil
}

func copySdkToDefaultPath(sdkType, version string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	sdkPath := filepath.Join(cfg.DefaultSdkPath, "sdks", sdkType, version)
	if _, err := os.Stat(sdkPath); os.IsNotExist(err) {
		log.Printf("SDK %s version %s not found in default path.\n", sdkType, version)
		return nil
	}

	destPath := filepath.Join(cfg.DefaultSdkPath, "sdks", sdkType, "default")
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return err
	}

	return copyDir(sdkPath, destPath)

}

func isDownloaded(sdkType, version string) bool {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}

	sdkPath := filepath.Join(cfg.DefaultSdkPath, "sdks", sdkType, version)
	_, err = os.Stat(sdkPath)
	return !os.IsNotExist(err)
}

func copyDir(src string, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}
