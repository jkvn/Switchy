package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
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
