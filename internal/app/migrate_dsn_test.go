package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrationSourceDir(t *testing.T) {
	t.Parallel()

	dir, err := MigrationSourceDir("mysql")
	require.NoError(t, err)
	require.Equal(t, "migrations/mysql", dir)

	dir, err = MigrationSourceDir("postgres")
	require.NoError(t, err)
	require.Equal(t, "migrations/postgres", dir)

	dir, err = MigrationSourceDir("sqlite")
	require.NoError(t, err)
	require.Equal(t, "migrations/sqlite", dir)

	_, err = MigrationSourceDir("oracle")
	require.Error(t, err)
}

func TestBuildMigrationDatabaseURL_MySQL(t *testing.T) {
	t.Parallel()

	u, err := BuildMigrationDatabaseURL("mysql", "user:pass@tcp(localhost:3306)/db?parseTime=true")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(u, "mysql://"))

	u2, err := BuildMigrationDatabaseURL("mysql", "mysql://user:pass@tcp(localhost:3306)/db")
	require.NoError(t, err)
	require.Equal(t, "mysql://user:pass@tcp(localhost:3306)/db", u2)
}

func TestBuildMigrationDatabaseURL_PostgresFullURL(t *testing.T) {
	t.Parallel()

	raw := "postgres://u:p@localhost:5432/omnibot?sslmode=disable"
	u, err := BuildMigrationDatabaseURL("postgres", raw)
	require.NoError(t, err)
	require.Equal(t, raw, u)

	raw2 := "postgresql://u:p@localhost:5432/omnibot"
	u2, err := BuildMigrationDatabaseURL("postgresql", raw2)
	require.NoError(t, err)
	require.Equal(t, raw2, u2)
}

func TestBuildMigrationDatabaseURL_PostgresKeywordDSN(t *testing.T) {
	t.Parallel()

	u, err := BuildMigrationDatabaseURL("postgres", "host=localhost port=5432 user=u password=p dbname=omnibot sslmode=disable")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(u, "postgres://"))
	require.Contains(t, u, "omnibot")
	require.Contains(t, u, "sslmode=disable")
	require.False(t, strings.HasPrefix(u, "postgres://host="), "must not blindly prefix keyword DSN")
}

func TestBuildMigrationDatabaseURL_SQLite(t *testing.T) {
	t.Parallel()

	u, err := BuildMigrationDatabaseURL("sqlite", "/tmp/test.db")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(u, "sqlite://"))
	require.Contains(t, u, "_pragma=foreign_keys(1)")
}

func TestEnsureSQLiteForeignKeysDSN(t *testing.T) {
	t.Parallel()

	require.Contains(t, EnsureSQLiteForeignKeysDSN("file:test.db"), "_pragma=foreign_keys(1)")
	require.Equal(t, "file:test.db?_pragma=foreign_keys(1)&x=1", EnsureSQLiteForeignKeysDSN("file:test.db?_pragma=foreign_keys(1)&x=1"))
}
