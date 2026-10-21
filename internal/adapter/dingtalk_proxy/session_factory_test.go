package dingtalk_proxy

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_enterprise"
	"github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_stream"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/stretchr/testify/require"
)

func TestDingtalkProxy_NewStreamSession_DistinctSessions(t *testing.T) {
	t.Parallel()

	proxy := NewAdapter(
		zerolog.Nop(),
		dingtalk_stream.NewAdapter(zerolog.Nop()),
		dingtalk_enterprise.NewAdapter(zerolog.Nop()),
	)

	var _ port.StreamSessionFactory = proxy

	a := proxy.NewStreamSession()
	b := proxy.NewStreamSession()
	require.NotSame(t, a, b)
	require.NotSame(t, a, proxy)

	pa, ok := a.(*DingtalkProxyAdapter)
	require.True(t, ok)
	pb, ok := b.(*DingtalkProxyAdapter)
	require.True(t, ok)
	require.NotSame(t, pa.streamAdapter, pb.streamAdapter)
}
