package server

import (
	"sync"
	"testing"
	"time"

	"github.com/sivdead/OmniBotGo/pkg/logger"
	rmqrpc "github.com/sivdead/OmniBotGo/pkg/rabbitmq/rmq_rpc"
	"github.com/stretchr/testify/require"
)

type nopLogger struct{}

func (nopLogger) Debug(_ interface{}, _ ...interface{}) {}
func (nopLogger) Info(_ string, _ ...interface{})       {}
func (nopLogger) Warn(_ string, _ ...interface{})       {}
func (nopLogger) Error(_ interface{}, _ ...interface{}) {}
func (nopLogger) Fatal(_ interface{}, _ ...interface{}) {}

// Ensure Interface is satisfied without importing concrete logger.
var _ logger.Interface = nopLogger{}

func TestShutdownIdempotentNoPanic(t *testing.T) {
	t.Parallel()

	s := &Server{
		conn:    rmqrpc.New("test", rmqrpc.Config{URL: "amqp://invalid", WaitTime: time.Millisecond, Attempts: 1}),
		error:   make(chan error, 1),
		stop:    make(chan struct{}),
		timeout: time.Millisecond,
		logger:  nopLogger{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			require.NotPanics(t, func() {
				err := s.Shutdown()
				require.NoError(t, err)
			})
		}()
	}

	wg.Wait()
}

func TestReconnectCanceledByShutdown(t *testing.T) {
	t.Parallel()

	s := &Server{
		conn:    rmqrpc.New("test", rmqrpc.Config{URL: "amqp://127.0.0.1:1", WaitTime: 50 * time.Millisecond, Attempts: 20}),
		error:   make(chan error, 1),
		stop:    make(chan struct{}),
		timeout: time.Millisecond,
		logger:  nopLogger{},
	}

	done := make(chan struct{})

	go func() {
		s.reconnect()
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	require.NoError(t, s.Shutdown())

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("reconnect did not return after Shutdown canceled in-flight AttemptConnect")
	}
}
