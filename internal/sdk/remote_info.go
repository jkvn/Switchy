package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jkvn/Switchy/internal/config"
	"github.com/mholt/archiver"
	"github.com/schollz/progressbar/v3"
)

const SdkJsonUrl = "https://raw.githubusercontent.com/jkvn/Switchy/refs/heads/main/sdk/sdkVersions.json"

func getSdkList() (*SDKList, error) {
	resp, err := http.Get(SdkJsonUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SDK list from %s: %w", SdkJsonUrl, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch SDK list: received status code %d", resp.StatusCode)
	}

	var list SDKList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode SDK list JSON: %w", err)
	}
	return &list, nil
}

func GetSdkTypes() ([]string, error) {
	list, err := getSdkList()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(list.Sdks))
	for _, sdk := range list.Sdks {
		names = append(names, sdk.Name)
	}
	return names, nil
}

func GetSdks(typeName string) ([]Version, error) {
	list, err := getSdkList()
	if err != nil {
		return nil, err
	}

	for _, sdk := range list.Sdks {
		if sdk.Name == typeName {
			return sdk.Versions, nil
		}
	}
	return nil, fmt.Errorf("SDK type %q not found", typeName)
}

func DownloadSdk(sdkType, version string) (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	downloadCacheDir := filepath.Join(cfg.DefaultSdkPath, "cache")
	if err := os.MkdirAll(downloadCacheDir, 0755); err != nil {
		return "", fmt.Errorf("error creating download cache directory '%s': %w", downloadCacheDir, err)
	}

	versions, err := GetSdks(sdkType)
	if err != nil {
		return "", err
	}

	for _, v := range versions {
		if v.Version == version {
			resp, err := http.Get(v.Link)
			if err != nil {
				return "", fmt.Errorf("failed to download SDK from %s: %w", v.Link, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("failed to download SDK: received status code %d (%s)", resp.StatusCode, resp.Status)
			}

			bar := progressbar.DefaultBytes(resp.ContentLength, fmt.Sprintf("Downloading %s %s", sdkType, version))

			ext := getArchiveExtension(v.Link)
			fileName := filepath.Join(downloadCacheDir, fmt.Sprintf("%s-%s%s", sdkType, version, ext))
			file, err := os.Create(fileName)
			if err != nil {
				return "", fmt.Errorf("failed to create file '%s': %w", fileName, err)
			}
			defer file.Close()

			if _, err := io.Copy(io.MultiWriter(file, bar), resp.Body); err != nil {
				return "", fmt.Errorf("failed to save SDK to file: %w", err)
			}

			fmt.Printf("\nSDK %s version %s downloaded successfully to %s.\n", sdkType, version, fileName)
			return fileName, nil
		}
	}
	return "", fmt.Errorf("version %q not found for SDK type %q", version, sdkType)
}

func getArchiveExtension(link string) string {
	if strings.HasSuffix(link, ".tar.gz") {
		return ".tar.gz"
	}
	if strings.HasSuffix(link, ".tar.xz") {
		return ".tar.xz"
	}
	if strings.HasSuffix(link, ".tar.bz2") {
		return ".tar.bz2"
	}
	return filepath.Ext(link)
}

func ExtractSdk(fileName string, sdkType string, version string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	finalDestDir := filepath.Join(cfg.DefaultSdkPath, "sdks", sdkType, version)
	tempExtractDir := filepath.Join(os.TempDir(), fmt.Sprintf("switchy_extract_temp_%d", os.Getpid()))

	if err := os.MkdirAll(finalDestDir, 0755); err != nil {
		return fmt.Errorf("error creating final destination directory '%s': %w", finalDestDir, err)
	}
	if err := os.MkdirAll(tempExtractDir, 0755); err != nil {
		return fmt.Errorf("error creating temporary extraction directory '%s': %w", tempExtractDir, err)
	}
	defer os.RemoveAll(tempExtractDir)

	err = archiver.Unarchive(fileName, tempExtractDir)
	if err != nil {
		decompressedFileName := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))
		tempDecompressedPath := filepath.Join(os.TempDir(), decompressedFileName)

		decompressErr := archiver.DecompressFile(fileName, tempDecompressedPath)
		if decompressErr != nil {
			return fmt.Errorf("failed to extract or decompress '%s': %w (original unarchive error: %v)", fileName, decompressErr, err)
		}
		defer os.Remove(tempDecompressedPath)

		unarchiveDecompressedErr := archiver.Unarchive(tempDecompressedPath, tempExtractDir)
		if unarchiveDecompressedErr != nil {
			finalFilePath := filepath.Join(finalDestDir, filepath.Base(tempDecompressedPath))
			if err := os.Rename(tempDecompressedPath, finalFilePath); err != nil {
				return fmt.Errorf("failed to move decompressed file '%s' to '%s': %w", tempDecompressedPath, finalFilePath, err)
			}
			fmt.Printf("SDK successfully extracted (decompressed single file) to %s\n", finalFilePath)
			return nil
		}
	}

	entries, err := os.ReadDir(tempExtractDir)
	if err != nil {
		return fmt.Errorf("failed to read temporary extraction directory '%s': %w", tempExtractDir, err)
	}

	if len(entries) == 1 && entries[0].IsDir() {
		sourcePath := filepath.Join(tempExtractDir, entries[0].Name())
		items, err := os.ReadDir(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to read inner directory '%s': %w", sourcePath, err)
		}
		for _, item := range items {
			oldPath := filepath.Join(sourcePath, item.Name())
			newPath := filepath.Join(finalDestDir, item.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to move item '%s' to '%s': %w", oldPath, newPath, err)
			}
		}
	} else {
		for _, entry := range entries {
			oldPath := filepath.Join(tempExtractDir, entry.Name())
			newPath := filepath.Join(finalDestDir, entry.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to move entry '%s' to '%s': %w", oldPath, newPath, err)
			}
		}
	}

	fmt.Printf("SDK successfully extracted to %s\n", finalDestDir)
	return nil
}
