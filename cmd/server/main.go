package main

import (
	"fmt"
	"os"

	"broadcast-platform/internal/bootstrap"
)

func main() {
	configPath := "configs/config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	if err := bootstrap.Run(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

