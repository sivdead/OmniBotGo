package app

import (
	"fmt"
	"net/url"
	"strings"
)

// MigrationSourceDir returns the dialect-specific migrations directory (E02).
func MigrationSourceDir(dbType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "", "mysql":
		return "migrations/mysql", nil
	case "postgres", "postgresql":
		return "migrations/postgres", nil
	case "sqlite", "sqlite3":
		return "migrations/sqlite", nil
	default:
		return "", fmt.Errorf("unsupported database type %q (supported: mysql, postgres, sqlite)", dbType)
	}
}

// BuildMigrationDatabaseURL builds a golang-migrate database URL (E03/E04).
//
// MySQL: mysql://user:pass@tcp(host:port)/db?...
// Postgres: accepts full postgres(ql):// URLs as-is; converts keyword DSNs
//
//	(host=... user=... dbname=...) into a URL; otherwise prefixes postgres://.
//
// SQLite: sqlite://path or sqlite://file:path?_pragma=...
func BuildMigrationDatabaseURL(dbType, dsn string) (string, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return "", fmt.Errorf("database DSN is empty")
	}

	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "", "mysql":
		if strings.HasPrefix(dsn, "mysql://") {
			return dsn, nil
		}

		return "mysql://" + dsn, nil

	case "postgres", "postgresql":
		return buildPostgresMigrationURL(dsn)

	case "sqlite", "sqlite3":
		return buildSQLiteMigrationURL(dsn)

	default:
		return "", fmt.Errorf("unsupported database type %q", dbType)
	}
}

func buildPostgresMigrationURL(dsn string) (string, error) {
	lower := strings.ToLower(dsn)
	if strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") {
		return dsn, nil
	}

	// libpq keyword/value DSN (e.g. host=localhost user=x password=y dbname=z sslmode=disable)
	if strings.Contains(dsn, "=") && !strings.Contains(dsn, "://") {
		u, err := postgresKeywordDSNToURL(dsn)
		if err != nil {
			return "", err
		}

		return u, nil
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
		return "", fmt.Errorf("postgres keyword DSN missing dbname")
	}

	user := fields["user"]
	password := fields["password"]

	u := &url.URL{
		Scheme: "postgres",
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
	if dsn == "" || dsn == ":memory:" {
		if dsn == ":memory:" {
			return "file::memory:?cache=shared&_pragma=foreign_keys(1)"
		}

		return dsn
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
