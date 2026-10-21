package client

import (
	"sync"
	"testing"
	"time"

	rmqrpc "github.com/sivdead/OmniBotGo/pkg/rabbitmq/rmq_rpc"
	"github.com/stretchr/testify/require"
)

func TestClientShutdownIdempotentNoPanic(t *testing.T) {
	t.Parallel()

	c := &Client{
		conn:    rmqrpc.New("test", rmqrpc.Config{URL: "amqp://invalid", WaitTime: time.Millisecond, Attempts: 1}),
		error:   make(chan error, 1),
		stop:    make(chan struct{}),
		calls:   make(map[string]*pendingCall),
		timeout: time.Millisecond,
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			require.NotPanics(t, func() {
				err := c.Shutdown()
				require.NoError(t, err)
			})
		}()
	}

	wg.Wait()
}

func TestClientReconnectCanceledByShutdown(t *testing.T) {
	t.Parallel()

	c := &Client{
		conn:    rmqrpc.New("test", rmqrpc.Config{URL: "amqp://127.0.0.1:1", WaitTime: 50 * time.Millisecond, Attempts: 20}),
		error:   make(chan error, 1),
		stop:    make(chan struct{}),
		calls:   make(map[string]*pendingCall),
		timeout: time.Millisecond,
	}

	done := make(chan struct{})

	go func() {
		c.reconnect()
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	require.NoError(t, c.Shutdown())

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("reconnect did not return after Shutdown canceled in-flight AttemptConnect")
	}
}
