package dingtalk_stream

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/stretchr/testify/require"
)

func TestNewStreamSession_ReturnsIndependentAdapters(t *testing.T) {
	t.Parallel()

	factory := NewAdapter(zerolog.Nop())
	a, ok := factory.NewStreamSession().(*DingtalkStreamAdapter)
	require.True(t, ok)
	b, ok := factory.NewStreamSession().(*DingtalkStreamAdapter)
	require.True(t, ok)
	require.NotSame(t, a, b)
	require.NotSame(t, factory, a)

	var _ port.StreamSessionFactory = factory
}

func TestStreamController_StopClearsRunningWithoutClient(t *testing.T) {
	t.Parallel()

	ctrl := NewStreamController(
		failHandler{},
		zerolog.Nop(),
		&config.DingtalkStreamConfig{ClientID: "id", ClientSecret: "sec"},
		"ch-1",
	)

	// Simulate C03 post-Start state: marked running with cancel/done, no live SDK.
	runCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	close(done)
	ctrl.mu.Lock()
	ctrl.isRunning = true
	ctrl.cancel = cancel
	ctrl.done = done
	ctrl.mu.Unlock()

	require.True(t, ctrl.IsRunning())
	require.NoError(t, ctrl.Stop())
	require.False(t, ctrl.IsRunning())
	require.Equal(t, context.Canceled, runCtx.Err())
}

func TestStreamController_StopIdempotentWhenNotRunning(t *testing.T) {
	t.Parallel()

	ctrl := NewStreamController(failHandler{}, zerolog.Nop(), &config.DingtalkStreamConfig{}, "ch-2")
	require.NoError(t, ctrl.Stop())
	require.False(t, ctrl.IsRunning())
}

func TestStreamController_StopWaitsForDone(t *testing.T) {
	t.Parallel()

	ctrl := NewStreamController(failHandler{}, zerolog.Nop(), &config.DingtalkStreamConfig{}, "ch-3")
	done := make(chan struct{})
	_, cancel := context.WithCancel(context.Background())

	ctrl.mu.Lock()
	ctrl.isRunning = true
	ctrl.cancel = cancel
	ctrl.done = done
	ctrl.mu.Unlock()

	go func() {
		time.Sleep(20 * time.Millisecond)
		close(done)
	}()

	start := time.Now()

	require.NoError(t, ctrl.Stop())
	require.GreaterOrEqual(t, time.Since(start), 15*time.Millisecond)
	require.False(t, ctrl.IsRunning())
}
