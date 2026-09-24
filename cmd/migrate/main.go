package main

import (
	"fmt"
	"os"

	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/database"
	"broadcast-platform/internal/platform/logging"
)

func main() {
	configPath := "configs/config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: config: %v\n", err)
		os.Exit(1)
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: logging: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if cfg.Database.DSN == "" {
		fmt.Fprintf(os.Stderr, "fatal: database DSN is empty\n")
		os.Exit(1)
	}

	db, err := database.New(cfg.Database, logger.WithModule("migrate"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	fmt.Println("Running migrations...")
	if err := db.RunMigrations("migrations"); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: migrations: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migrations completed successfully.")
}
