package app

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const (
	dbTypeMySQL         = "mysql"
	dbTypePostgres      = "postgres"
	dbTypePostgreSQLAlt = "postgresql"
	dbTypeSQLite        = "sqlite"
	dbTypeSQLite3Alt    = "sqlite3"

	sqliteMemoryDSN = ":memory:"
)

var (
	errUnsupportedDBType = errors.New("unsupported database type")
	errEmptyDatabaseDSN  = errors.New("database DSN is empty")
	errPostgresMissingDB = errors.New("postgres keyword DSN missing dbname")
)

// MigrationSourceDir returns the dialect-specific migrations directory (E02).
func MigrationSourceDir(dbType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "", dbTypeMySQL:
		return "migrations/" + dbTypeMySQL, nil
	case dbTypePostgres, dbTypePostgreSQLAlt:
		return "migrations/" + dbTypePostgres, nil
	case dbTypeSQLite, dbTypeSQLite3Alt:
		return "migrations/" + dbTypeSQLite, nil
	default:
		return "", fmt.Errorf("%w: %q (supported: mysql, postgres, sqlite)", errUnsupportedDBType, dbType)
	}
}

// BuildMigrationDatabaseURL builds a golang-migrate database URL (E03/E04).
func BuildMigrationDatabaseURL(dbType, dsn string) (string, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return "", errEmptyDatabaseDSN
	}

	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "", dbTypeMySQL:
		if strings.HasPrefix(dsn, "mysql://") {
			return dsn, nil
		}

		return "mysql://" + dsn, nil

	case dbTypePostgres, dbTypePostgreSQLAlt:
		return buildPostgresMigrationURL(dsn)

	case dbTypeSQLite, dbTypeSQLite3Alt:
		return buildSQLiteMigrationURL(dsn)

	default:
		return "", fmt.Errorf("%w: %q", errUnsupportedDBType, dbType)
	}
}

func buildPostgresMigrationURL(dsn string) (string, error) {
	lower := strings.ToLower(dsn)
	if strings.HasPrefix(lower, dbTypePostgres+"://") || strings.HasPrefix(lower, dbTypePostgreSQLAlt+"://") {
		return dsn, nil
	}

	// libpq keyword/value DSN (e.g. host=localhost user=x password=y dbname=z sslmode=disable)
	if strings.Contains(dsn, "=") && !strings.Contains(dsn, "://") {
		return postgresKeywordDSNToURL(dsn)
	}

	// user:pass@host:port/db?sslmode=disable style (no scheme)
	return "postgres://" + dsn, nil
}

func postgresKeywordDSNToURL(dsn string) (string, error) {
	fields := parseLibpqKeywordDSN(dsn)

	host := fields["host"]
	if host == "" {
		host = "localhost"
	}

	port := fields["port"]
	if port == "" {
		port = "5432"
	}

	dbname := fields["dbname"]
	if dbname == "" {
		dbname = fields["database"]
	}

	if dbname == "" {
		return "", errPostgresMissingDB
	}

	user := fields["user"]
	password := fields["password"]

	u := &url.URL{
		Scheme: dbTypePostgres,
		Host:   host + ":" + port,
		Path:   "/" + dbname,
	}

	if user != "" {
		if password != "" {
			u.User = url.UserPassword(user, password)
		} else {
			u.User = url.User(user)
		}
	}

	q := url.Values{}

	for k, v := range fields {
		switch k {
		case "host", "port", "dbname", "database", "user", "password":
			continue
		default:
			q.Set(k, v)
		}
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

//nolint:gocyclo,cyclop // simple libpq keyword tokenizer
func parseLibpqKeywordDSN(dsn string) map[string]string {
	out := make(map[string]string)

	var key, val strings.Builder

	inKey := true
	inQuote := false

	var quote byte

	flush := func() {
		k := strings.TrimSpace(key.String())
		if k != "" {
			out[k] = val.String()
		}

		key.Reset()
		val.Reset()

		inKey = true
	}

	for i := 0; i < len(dsn); i++ {
		c := dsn[i]

		switch {
		case inQuote:
			if c == quote {
				inQuote = false
			} else {
				val.WriteByte(c)
			}
		case c == '\'' || c == '"':
			inQuote = true
			quote = c
		case inKey && c == '=':
			inKey = false
		case !inKey && (c == ' ' || c == '\t'):
			flush()
		case inKey:
			key.WriteByte(c)
		default:
			val.WriteByte(c)
		}
	}

	flush()

	return out
}

func buildSQLiteMigrationURL(dsn string) (string, error) {
	dsn = EnsureSQLiteForeignKeysDSN(dsn)
	if strings.HasPrefix(strings.ToLower(dsn), "sqlite://") {
		return dsn, nil
	}

	return "sqlite://" + dsn, nil
}

// EnsureSQLiteForeignKeysDSN appends modernc _pragma=foreign_keys(1) when missing (E08).
func EnsureSQLiteForeignKeysDSN(dsn string) string {
	if dsn == "" {
		return dsn
	}

	if dsn == sqliteMemoryDSN {
		return "file::memory:?cache=shared&_pragma=foreign_keys(1)"
	}

	lower := strings.ToLower(dsn)
	if strings.Contains(lower, "_pragma=foreign_keys") || strings.Contains(lower, "foreign_keys(1)") {
		return dsn
	}

	if strings.Contains(dsn, "?") {
		return dsn + "&_pragma=foreign_keys(1)"
	}

	return dsn + "?_pragma=foreign_keys(1)"
}
