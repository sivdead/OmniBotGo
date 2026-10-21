package service

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_stream"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/stretchr/testify/require"
)

func TestResolveStreamAdapter_UsesFactory(t *testing.T) {
	t.Parallel()

	cm := &ConnectionManager{logger: zerolog.Nop()}
	factory := dingtalk_stream.NewAdapter(zerolog.Nop())

	a, ok := cm.resolveStreamAdapter(factory)
	require.True(t, ok)
	b, ok := cm.resolveStreamAdapter(factory)
	require.True(t, ok)
	require.NotSame(t, a, b)

	var _ port.StreamSessionFactory = factory
}
