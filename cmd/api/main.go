package main

import (
	"log"

	"em/config"
)

// @title Subscriptions API
// @version 1.0
// @description Subscriptions CRUDL Service.
// @BasePath /api/v1
func main() {
	cfg, err := config.LoadConfig("config/config-default.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := Migrate(cfg.Postgres.DSN()); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	NewApp(cfg).Run()
}
