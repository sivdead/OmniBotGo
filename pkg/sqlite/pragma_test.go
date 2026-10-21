package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureForeignKeysPragma(t *testing.T) {
	t.Parallel()

	require.Contains(t, ensureForeignKeysPragma("file:test.db"), "_pragma=foreign_keys(1)")
	require.Contains(t, ensureForeignKeysPragma(":memory:"), "_pragma=foreign_keys(1)")
	require.Equal(t,
		"file:test.db?_pragma=foreign_keys(1)&x=1",
		ensureForeignKeysPragma("file:test.db?_pragma=foreign_keys(1)&x=1"),
	)
}
