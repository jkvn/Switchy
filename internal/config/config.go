package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configFileName = "switchy.conf"
	defaultSDKPath = ".switchy/sdks"
)

type Config struct {
	DefaultSdkPath string
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "switchy", configFileName), nil
}

func EnsureConfigFile() (string, error) {
	path, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultPath := filepath.Join(os.Getenv("HOME"), defaultSDKPath)
		content := fmt.Sprintf("# Switchy config\ndefault_sdk_path=%s\n", defaultPath)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return "", err
		}
	}

	return path, nil
}

func LoadConfig() (*Config, error) {
	path, err := EnsureConfigFile()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	cfg := &Config{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid config line: " + line)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "default_sdk_path":
			cfg.DefaultSdkPath = val
		}
	}

	if cfg.DefaultSdkPath == "" {
		cfg.DefaultSdkPath = filepath.Join(os.Getenv("HOME"), defaultSDKPath)
	}

	return cfg, nil
}
