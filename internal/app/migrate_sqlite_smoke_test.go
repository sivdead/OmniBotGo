package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

func TestSQLiteMigrationsUpSmoke(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "smoke.db")

	sourceDir, err := MigrationSourceDir("sqlite")
	require.NoError(t, err)

	absSource, err := filepath.Abs(sourceDir)
	if err != nil || !dirExists(absSource) {
		// go test may run with package cwd; try module root via walking up.
		absSource = findMigrationsDir(t, sourceDir)
	}

	dbURL, err := BuildMigrationDatabaseURL("sqlite", dbPath)
	require.NoError(t, err)

	m, err := migrate.New("file://"+absSource, dbURL)
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = m.Close() })

	err = m.Up()
	require.NoError(t, err)

	version, dirty, err := m.Version()
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, uint(1), version)
}

func dirExists(path string) bool {
	st, err := os.Stat(path)

	return err == nil && st.IsDir()
}

func findMigrationsDir(t *testing.T, rel string) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	dir := wd
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, rel)
		if dirExists(candidate) {
			abs, err := filepath.Abs(candidate)
			require.NoError(t, err)

			return abs
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	t.Fatalf("could not locate %s from %s", rel, wd)

	return ""
}
