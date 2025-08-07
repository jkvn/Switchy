package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func DownloadSdk(sdkType, version string) error {
	versions, err := GetSdks(sdkType)
	if err != nil {
		return err
	}

	for _, v := range versions {
		if v.Version == version {
			resp, err := http.Get(v.Link)
			if err != nil {
				return fmt.Errorf("failed to download SDK: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to download SDK: %s", resp.Status)
			}
			file, err := os.Create(fmt.Sprintf("%s-%s.zip", sdkType, version))

			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer file.Close()

			if _, err := io.Copy(file, resp.Body); err != nil {
				return fmt.Errorf("failed to save SDK: %w", err)
			}

			fmt.Printf("SDK %s version %s downloaded successfully.\n", sdkType, version)
			return nil
		}
	}
	return fmt.Errorf("version %q not found for SDK type %q", version, sdkType)
}
