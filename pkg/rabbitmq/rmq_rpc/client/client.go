package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	rmqrpc "github.com/sivdead/OmniBotGo/pkg/rabbitmq/rmq_rpc"
)

// ErrConnectionClosed -.
var ErrConnectionClosed = errors.New("rmq_rpc client - Client - RemoteCall - Connection closed")

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 2 * time.Second
)

// Message -.
type Message struct {
	Queue         string
	Priority      uint8
	ContentType   string
	Body          []byte
	ReplyTo       string
	CorrelationID string
}

type pendingCall struct {
	done   chan struct{}
	status string
	body   []byte
}

// Client -.
type Client struct {
	conn           *rmqrpc.Connection
	serverExchange string
	error          chan error
	stop           chan struct{}

	rw    sync.RWMutex
	calls map[string]*pendingCall

	timeout time.Duration

	stopOnce        sync.Once
	errOnce         sync.Once
	shuttingDown    atomic.Bool
	reconnectMu     sync.Mutex
	reconnectCancel context.CancelFunc
}

// New -.
func New(url, serverExchange, clientExchange string, opts ...Option) (*Client, error) {
	cfg := rmqrpc.Config{
		URL:      url,
		WaitTime: _defaultWaitTime,
		Attempts: _defaultAttempts,
	}

	c := &Client{
		conn:           rmqrpc.New(clientExchange, cfg),
		serverExchange: serverExchange,
		error:          make(chan error, 1),
		stop:           make(chan struct{}),
		calls:          make(map[string]*pendingCall),
		timeout:        _defaultTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(c)
	}

	err := c.conn.AttemptConnect()
	if err != nil {
		return nil, fmt.Errorf("rmq_rpc client - NewClient - c.conn.AttemptConnect: %w", err)
	}

	go c.consumer()

	return c, nil
}

func (c *Client) publish(corrID, handler string, request interface{}) error {
	var (
		requestBody []byte
		err         error
	)

	if request != nil {
		requestBody, err = json.Marshal(request)
		if err != nil {
			return err
		}
	}

	err = c.conn.Channel.Publish(c.serverExchange, "", false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       c.conn.ConsumerExchange,
			Type:          handler,
			Body:          requestBody,
		})
	if err != nil {
		return fmt.Errorf("c.Channel.Publish: %w", err)
	}

	return nil
}

// RemoteCall -.
func (c *Client) RemoteCall(handler string, request, response interface{}) error { //nolint:cyclop // complex func
	select {
	case <-c.stop:
		return ErrConnectionClosed
	default:
	}

	corrID := uuid.New().String()

	err := c.publish(corrID, handler, request)
	if err != nil {
		return fmt.Errorf("rmq_rpc client - Client - RemoteCall - c.publish: %w", err)
	}

	call := &pendingCall{done: make(chan struct{})}

	c.addCall(corrID, call)
	defer c.deleteCall(corrID)

	select {
	case <-time.After(c.timeout):
		return rmqrpc.ErrTimeout
	case <-c.stop:
		return ErrConnectionClosed
	case <-call.done:
	}

	if call.status == rmqrpc.Success {
		err = json.Unmarshal(call.body, &response)
		if err != nil {
			return fmt.Errorf("rmq_rpc client - Client - RemoteCall - json.Unmarshal: %w", err)
		}

		return nil
	}

	if call.status == rmqrpc.ErrBadHandler.Error() {
		return rmqrpc.ErrBadHandler
	}

	if call.status == rmqrpc.ErrInternalServer.Error() {
		return rmqrpc.ErrInternalServer
	}

	return nil
}

func (c *Client) consumer() {
	for {
		select {
		case <-c.stop:
			return
		case d, opened := <-c.conn.Delivery:
			if !opened {
				c.reconnect()

				return
			}

			_ = d.Ack(false) //nolint:errcheck // don't need this

			c.getCall(&d)
		}
	}
}

func (c *Client) reconnect() {
	if c.shuttingDown.Load() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	c.reconnectMu.Lock()
	c.reconnectCancel = cancel
	c.reconnectMu.Unlock()

	defer func() {
		cancel()

		c.reconnectMu.Lock()
		c.reconnectCancel = nil
		c.reconnectMu.Unlock()
	}()

	err := c.conn.AttemptConnectContext(ctx)
	if err != nil {
		if c.shuttingDown.Load() || ctx.Err() != nil {
			return
		}

		c.reportFatal(err)

		return
	}

	if c.shuttingDown.Load() {
		return
	}

	go c.consumer()
}

func (c *Client) reportFatal(err error) {
	c.errOnce.Do(func() {
		select {
		case c.error <- err:
		default:
		}

		close(c.error)
	})
}

func (c *Client) signalStop() {
	c.stopOnce.Do(func() {
		close(c.stop)
	})
}

func (c *Client) getCall(d *amqp.Delivery) {
	c.rw.RLock()
	call, ok := c.calls[d.CorrelationId]
	c.rw.RUnlock()

	if !ok {
		return
	}

	call.status = d.Type
	call.body = d.Body
	close(call.done)
}

func (c *Client) addCall(corrID string, call *pendingCall) {
	c.rw.Lock()
	c.calls[corrID] = call
	c.rw.Unlock()
}

func (c *Client) deleteCall(corrID string) {
	c.rw.Lock()
	delete(c.calls, corrID)
	c.rw.Unlock()
}

// Notify -.
func (c *Client) Notify() <-chan error {
	return c.error
}

// Shutdown stops the consumer, cancels in-flight reconnect, and closes AMQP (F01).
func (c *Client) Shutdown() error {
	if !c.shuttingDown.CompareAndSwap(false, true) {
		return nil
	}

	c.reconnectMu.Lock()
	cancel := c.reconnectCancel
	c.reconnectMu.Unlock()

	if cancel != nil {
		cancel()
	}

	c.signalStop()
	time.Sleep(c.timeout)

	if c.conn.Connection == nil {
		return nil
	}

	err := c.conn.Connection.Close()
	if err != nil {
		return fmt.Errorf("rmq_rpc client - Client - Shutdown - c.Connection.Close: %w", err)
	}

	return nil
}
