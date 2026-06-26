package main

import (
	"fmt"

	"em"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func Migrate(dsn string) (migrateErr error) {
	src, err := iofs.New(em.MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	defer func() {
		if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
			migrateErr = fmt.Errorf("migrate: srcErr %w, dbErr %w", srcErr, dbErr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
