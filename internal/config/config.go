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
	defaultSDKPath = ".switchy"
)

type Config struct {
	DefaultSdkPath string
}

func GetConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "switchy", configFileName), nil
}

func EnsureConfigFile() (string, error) {
	path, err := GetConfigPath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		home, _ := os.UserHomeDir()
		def := filepath.Join(home, defaultSDKPath)
		content := fmt.Sprintf("default_sdk_path=%s\n", def)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
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
	cfg := &Config{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			return nil, errors.New("invalid config line")
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k == "default_sdk_path" {
			cfg.DefaultSdkPath = v
		}
	}
	if cfg.DefaultSdkPath == "" {
		home, _ := os.UserHomeDir()
		cfg.DefaultSdkPath = filepath.Join(home, defaultSDKPath)
	}
	return cfg, nil
}