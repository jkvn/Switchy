package local

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jkvn/Switchy/internal/config"
	copier "github.com/otiai10/copy"
)

var ErrSDKNotFound = errors.New("sdk not found")

func SetSdkVersion(sdkType, version string) error {
	if !isDownloaded(sdkType, version) {
		return ErrSDKNotFound
	}
	return copySdkToDefaultPath(sdkType, version)
}

func SetSdkEnvironment(sdkType, version string) error {
	return nil
}

func copySdkToDefaultPath(sdkType, version string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	src := filepath.Join(cfg.DefaultSdkPath, "sdks", sdkType, version)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("stat src: %w", err)
	}
	dst := filepath.Join(cfg.DefaultSdkPath, "sdks", sdkType, "default")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("mkdir dst: %w", err)
	}
	return copier.Copy(src, dst, copier.Options{
		PreserveTimes: true,
	})
}

func isDownloaded(sdkType, version string) bool {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}
	path := filepath.Join(cfg.DefaultSdkPath, "sdks", sdkType, version)
	_, statErr := os.Stat(path)
	return statErr == nil
}
