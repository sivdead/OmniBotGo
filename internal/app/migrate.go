//go:build migrate

package app

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/sivdead/OmniBotGo/internal/config"

	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts = 20
	_defaultTimeout  = time.Second
)

func init() {
	// Get database configuration from Viper
	dbType := config.GetString("db.type")
	if dbType == "" {
		dbType = "mysql" // default to mysql for backward compatibility
	}

	dbDSN := config.GetString("db.dsn")
	if dbDSN == "" {
		log.Fatalf("migrate: database DSN not configured. Please set db.dsn in config file or DB_DSN environment variable")
	}

	// Construct migration URL based on database type
	var migrationURL string
	switch strings.ToLower(dbType) {
	case "mysql":
		migrationURL = fmt.Sprintf("mysql://%s", dbDSN)
	case "postgres", "postgresql":
		migrationURL = fmt.Sprintf("postgres://%s", dbDSN)
	default:
		log.Fatalf("migrate: unsupported database type: %s. Supported types: mysql, postgres", dbType)
	}

	log.Printf("Migrate: Using database type: %s", dbType)

	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", migrationURL)
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
