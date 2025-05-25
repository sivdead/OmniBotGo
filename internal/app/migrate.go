//go:build migrate

package app

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
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
	// Get database type and DSN from environment
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "mysql" // default to mysql for backward compatibility
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		// Fallback to MYSQL_DSN for backward compatibility
		dbDSN = os.Getenv("MYSQL_DSN")
		if dbDSN == "" {
			log.Fatalf("migrate: environment variable not declared: DB_DSN or MYSQL_DSN")
		}
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
