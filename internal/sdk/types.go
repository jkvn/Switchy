package sdk

type SDK struct {
	Name           string    `json:"name"`
	DefaultVersion string    `json:"defaultVersion"`
	Versions       []Version `json:"versions"`
}

type Version struct {
	Version   string                     `json:"version"`
	Platforms map[string]VersionPlatform `json:"platforms"`
}

type VersionPlatform struct {
	Link   string `json:"link"`
	Sha256 string `json:"sha256"`
}

type SDKList struct {
	Sdks []SDK `json:"sdks"`
}
