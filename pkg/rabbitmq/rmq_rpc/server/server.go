package server

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	rmqrpc "github.com/sivdead/OmniBotGo/pkg/rabbitmq/rmq_rpc"
)

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 2 * time.Second
)

// CallHandler -.
type CallHandler func(*amqp.Delivery) (interface{}, error)

// Server -.
type Server struct {
	conn   *rmqrpc.Connection
	error  chan error
	stop   chan struct{}
	router map[string]CallHandler

	timeout time.Duration

	logger logger.Interface

	stopOnce        sync.Once
	errOnce         sync.Once
	shuttingDown    atomic.Bool
	reconnectMu     sync.Mutex
	reconnectCancel context.CancelFunc
}

// New -.
func New(url, serverExchange string, router map[string]CallHandler, l logger.Interface, opts ...Option) (*Server, error) {
	cfg := rmqrpc.Config{
		URL:      url,
		WaitTime: _defaultWaitTime,
		Attempts: _defaultAttempts,
	}

	s := &Server{
		conn:    rmqrpc.New(serverExchange, cfg),
		error:   make(chan error, 1),
		stop:    make(chan struct{}),
		router:  router,
		timeout: _defaultTimeout,
		logger:  l,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	err := s.conn.AttemptConnect()
	if err != nil {
		return nil, fmt.Errorf("rmq_rpc server - NewServer - s.conn.AttemptConnect: %w", err)
	}

	return s, nil
}

// Start -.
func (s *Server) Start() {
	go s.consumer()
}

func (s *Server) consumer() {
	for {
		select {
		case <-s.stop:
			return
		case d, opened := <-s.conn.Delivery:
			if !opened {
				s.reconnect()

				return
			}

			_ = d.Ack(false) //nolint:errcheck // don't need this

			s.serveCall(&d)
		}
	}
}

func (s *Server) serveCall(d *amqp.Delivery) {
	callHandler, ok := s.router[d.Type]
	if !ok {
		s.publish(d, nil, rmqrpc.ErrBadHandler.Error())

		return
	}

	response, err := callHandler(d)
	if err != nil {
		s.publish(d, nil, rmqrpc.ErrInternalServer.Error())

		s.logger.Error(err, "rmq_rpc server - Server - serveCall - callHandler")

		return
	}

	body, err := json.Marshal(response)
	if err != nil {
		s.logger.Error(err, "rmq_rpc server - Server - serveCall - json.Marshal")
	}

	s.publish(d, body, rmqrpc.Success)
}

func (s *Server) publish(d *amqp.Delivery, body []byte, status string) {
	err := s.conn.Channel.Publish(d.ReplyTo, "", false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Type:          status,
			Body:          body,
		})
	if err != nil {
		s.logger.Error(err, "rmq_rpc server - Server - publish - s.conn.Channel.Publish")
	}
}

func (s *Server) reconnect() {
	if s.shuttingDown.Load() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	s.reconnectMu.Lock()
	s.reconnectCancel = cancel
	s.reconnectMu.Unlock()

	defer func() {
		cancel()

		s.reconnectMu.Lock()
		s.reconnectCancel = nil
		s.reconnectMu.Unlock()
	}()

	err := s.conn.AttemptConnectContext(ctx)
	if err != nil {
		if s.shuttingDown.Load() || ctx.Err() != nil {
			return
		}

		s.reportFatal(err)

		return
	}

	if s.shuttingDown.Load() {
		return
	}

	go s.consumer()
}

func (s *Server) reportFatal(err error) {
	s.errOnce.Do(func() {
		select {
		case s.error <- err:
		default:
		}

		close(s.error)
	})
}

func (s *Server) signalStop() {
	s.stopOnce.Do(func() {
		close(s.stop)
	})
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.error
}

// Shutdown stops the consumer, cancels in-flight reconnect, and closes AMQP (F01).
func (s *Server) Shutdown() error {
	if !s.shuttingDown.CompareAndSwap(false, true) {
		return nil
	}

	s.reconnectMu.Lock()
	cancel := s.reconnectCancel
	s.reconnectMu.Unlock()

	if cancel != nil {
		cancel()
	}

	s.signalStop()
	time.Sleep(s.timeout)

	if s.conn.Connection == nil {
		return nil
	}

	err := s.conn.Connection.Close()
	if err != nil {
		return fmt.Errorf("rmq_rpc server - Server - Shutdown - s.Connection.Close: %w", err)
	}

	return nil
}
