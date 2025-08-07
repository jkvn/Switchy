package main

import (
	"log"

	"github.com/jkvn/Switchy/cmd"
	"github.com/jkvn/Switchy/internal/config"
)

func main() {
	path, err := config.EnsureConfigFile()
	if err != nil {
		log.Fatalf("error while creating config: %v", err)
	}
	log.Println("config-path", path)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Println("path", cfg.DefaultSdkPath)

	cmd.Execute()
}
