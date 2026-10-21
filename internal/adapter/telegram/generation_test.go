package telegram

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/stretchr/testify/require"
)

func TestNewStreamSession_Independent(t *testing.T) {
	t.Parallel()

	factory := NewAdapter(zerolog.Nop())
	a := factory.NewStreamSession()
	b := factory.NewStreamSession()
	require.NotSame(t, a, b)

	var _ port.StreamSessionFactory = factory
}

func TestPollLoop_OldGenerationDoesNotClearNewConnection(t *testing.T) {
	t.Parallel()

	a := NewAdapter(zerolog.Nop())
	a.mu.Lock()
	a.generation = 1
	a.isConnected = true
	a.mu.Unlock()

	// Simulate old pollLoop defer for gen 1 after Start bumped to gen 2.
	a.mu.Lock()
	a.generation = 2
	a.isConnected = true
	a.mu.Unlock()

	a.mu.Lock()
	if a.generation == 1 {
		a.isConnected = false
	}
	a.mu.Unlock()

	require.True(t, a.IsConnected())
}

func TestStop_TimeoutReturnsError(t *testing.T) {
	t.Parallel()

	a := NewAdapter(zerolog.Nop())
	a.mu.Lock()
	a.isConnected = true
	a.cancel = func() {}
	a.wg.Add(1) // never Done — forces timeout wait
	a.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := a.Stop(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, errStopTimedOut)

	// Unblock WaitGroup so the background waiter can exit.
	a.wg.Done()
	time.Sleep(20 * time.Millisecond)
}

func TestStop_SuccessWhenPollExits(t *testing.T) {
	t.Parallel()

	a := NewAdapter(zerolog.Nop())

	var once sync.Once

	release := make(chan struct{})

	a.mu.Lock()
	a.isConnected = true
	a.cancel = func() { once.Do(func() { close(release) }) }
	a.wg.Add(1)
	a.mu.Unlock()

	go func() {
		<-release
		a.wg.Done()
	}()

	require.NoError(t, a.Stop(context.Background()))
}

func TestGenerationScopedDefer_ClearsOnlyMatchingGen(t *testing.T) {
	t.Parallel()

	a := NewAdapter(zerolog.Nop())
	handler := func(context.Context, *dto.UnifiedMessage) error { return nil }

	// Already-canceled context so pollLoop exits without HTTP.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	a.mu.Lock()
	a.generation = 7
	a.isConnected = true
	a.mu.Unlock()

	a.wg.Add(1)
	go a.pollLoop(ctx, handler, "token", "ch", 7)
	a.wg.Wait()

	require.False(t, a.IsConnected())
}
