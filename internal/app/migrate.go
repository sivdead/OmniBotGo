//go:build migrate

package app

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// migrate database drivers
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sivdead/OmniBotGo/internal/config"
)

const (
	_defaultAttempts = 20
	_defaultTimeout  = time.Second
)

func init() {
	dbType := config.GetString("db.type")
	if dbType == "" {
		dbType = "mysql"
	}

	dbDSN := config.GetString("db.dsn")
	if dbDSN == "" {
		log.Fatalf("migrate: database DSN not configured. Please set db.dsn in config file or DB_DSN environment variable")
	}

	sourceDir, err := MigrationSourceDir(dbType)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}

	migrationURL, err := BuildMigrationDatabaseURL(dbType, dbDSN)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}

	log.Printf("Migrate: Using database type: %s source=%s", dbType, sourceDir)

	var (
		attempts = _defaultAttempts
		m        *migrate.Migrate
	)

	sourceURL := fmt.Sprintf("file://%s", sourceDir)

	for attempts > 0 {
		m, err = migrate.New(sourceURL, migrationURL)
		if err == nil {
			break
		}

		log.Printf("Migrate: %s is trying to connect, attempts left: %d", dbType, attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Fatalf("Migrate: %s connect error: %s", dbType, err)
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Migrate: up error: %s", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")

		return
	}

	log.Printf("Migrate: up success")
}
