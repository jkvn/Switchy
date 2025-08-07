package sdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const SdkJsonUrl = "https://raw.githubusercontent.com/jkvn/Switchy/refs/heads/main/sdk/sdkVersions.json"

func GetSdkTypes() ([]string, error) {
	resp, err := http.Get(SdkJsonUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var list SDKList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(list.Sdks))
	for _, sdk := range list.Sdks {
		names = append(names, sdk.Name)
	}
	return names, nil
}

func GetSdks(typeName string) ([]Version, error) {
	resp, err := http.Get(SdkJsonUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var list SDKList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	for _, sdk := range list.Sdks {
		if sdk.Name == typeName {
			return sdk.Versions, nil
		}
	}
	return nil, fmt.Errorf("SDK type %q not found", typeName)
}

func  DownloadSdk(sdkType, version string) (string, error) {
	versions, err := GetSdks(sdkType)
	if err != nil {
		return "", err
	}

	for _, v := range versions {
		if v.Version == version {
			resp, err := http.Get(v.Link)
			if err != nil {
				return "", fmt.Errorf("failed to download SDK: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("failed to download SDK: %s", resp.Status)
			}

			var ext string
			if strings.HasSuffix(v.Link, ".tar.gz") {
				ext = ".tar.gz"
			} else if strings.HasSuffix(v.Link, ".zip") {
				ext = ".zip"
			} else {
				ext = ""
			}

			fileName := fmt.Sprintf("%s-%s%s", sdkType, version, ext)
			file, err := os.Create(fileName)
			if err != nil {
				return "", fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()

			if _, err := io.Copy(file, resp.Body); err != nil {
				return "", fmt.Errorf("failed to save SDK: %w", err)
			}

			fmt.Printf("SDK %s version %s downloaded successfully.\n", sdkType, version)
			return fileName, nil
		}
	}
	return "", fmt.Errorf("version %q not found for SDK type %q", version, sdkType)
}

func ExtractSdk(fileName string) error {
	var destDir string
	if strings.HasSuffix(fileName, ".tar.gz") {
		destDir = strings.TrimSuffix(fileName, ".tar.gz")
	} else if strings.HasSuffix(fileName, ".zip") {
		destDir = strings.TrimSuffix(fileName, ".zip")
	} else {
		return fmt.Errorf("unknown archive format: %s", fileName)
	}

	// Create the destination directory if it does not exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	if strings.HasSuffix(fileName, ".tar.gz") {
		// Open the file
		f, err := os.Open(fileName)
		if err != nil {
			return fmt.Errorf("error opening file: %w", err)
		}
		defer f.Close()

		gzr, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("error creating gzip reader: %w", err)
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading tar archive: %w", err)
			}

			targetPath := filepath.Join(destDir, header.Name)
			switch header.Typeflag {
			case tar.TypeDir:
				if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
					return fmt.Errorf("error creating directory %s: %w", targetPath, err)
				}
			case tar.TypeReg:
				if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
					return fmt.Errorf("error creating directory %s: %w", filepath.Dir(targetPath), err)
				}
				outFile, err := os.Create(targetPath)
				if err != nil {
					return fmt.Errorf("error creating file %s: %w", targetPath, err)
				}
				if _, err := io.Copy(outFile, tr); err != nil {
					outFile.Close()
					return fmt.Errorf("error writing file %s: %w", targetPath, err)
				}
				outFile.Close()
				if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
					return fmt.Errorf("error setting permissions for %s: %w", targetPath, err)
				}
			}
		}
	} else if strings.HasSuffix(fileName, ".zip") {
		r, err := zip.OpenReader(fileName)
		if err != nil {
			return fmt.Errorf("error opening ZIP archive: %w", err)
		}
		defer r.Close()

		for _, f := range r.File {
			fpath := filepath.Join(destDir, f.Name)
			if f.FileInfo().IsDir() {
				if err := os.MkdirAll(fpath, f.Mode()); err != nil {
					return fmt.Errorf("error creating directory %s: %w", fpath, err)
				}
				continue
			}

			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %w", filepath.Dir(fpath), err)
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("error creating file %s: %w", fpath, err)
			}

			rc, err := f.Open()
			if err != nil {
				outFile.Close()
				return fmt.Errorf("error opening file in ZIP: %w", err)
			}

			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return fmt.Errorf("error writing file %s: %w", fpath, err)
			}
		}
	} else {
		return fmt.Errorf("unknown archive format: %s", fileName)
	}

	fmt.Printf("SDK successfully extracted to %s\n", destDir)
	return nil
}

