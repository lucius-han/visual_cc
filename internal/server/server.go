package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/lucius-han/visual_cc/internal/event"
)

const (
	maxConnections = 16              // S3: max simultaneous connections
	maxEventBytes  = 16 * 1024      // S2: 16 KB per event line
	readDeadline   = 2 * time.Second // S4: per-connection read deadline
)

// Server listens on a Unix socket and dispatches received events to handler.
type Server struct {
	path     string
	handler  func(event.Event)
	listener net.Listener
	mu       sync.Mutex
	stopped  bool
	sem      chan struct{} // S3: connection semaphore
}

// New creates a new Server listening on socketPath.
// The handler is called for each successfully parsed event.
func New(socketPath string, handler func(event.Event)) (*Server, error) {
	os.Remove(socketPath)
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	// S1: restrict socket file to owner only (prevent other local users from injecting events)
	if err := os.Chmod(socketPath, 0600); err != nil {
		l.Close()
		return nil, fmt.Errorf("chmod socket: %w", err)
	}
	return &Server{
		path:     socketPath,
		handler:  handler,
		listener: l,
		sem:      make(chan struct{}, maxConnections), // S3
	}, nil
}

// Start accepts connections in a loop. Call in a goroutine.
func (s *Server) Start() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			stopped := s.stopped
			s.mu.Unlock()
			if stopped {
				return
			}
			continue
		}
		// S3: reject when at capacity instead of spawning unbounded goroutines
		select {
		case s.sem <- struct{}{}:
			go func() {
				defer func() { <-s.sem }()
				s.handleConn(conn)
			}()
		default:
			conn.Close()
		}
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	// S4: read deadline prevents goroutine leak from stalled clients
	conn.SetReadDeadline(time.Now().Add(readDeadline)) //nolint:errcheck
	// S2: bounded scanner buffer — prevents 64 KB-per-connection memory exhaustion
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 4096), maxEventBytes)
	for scanner.Scan() {
		var e event.Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
			s.handler(e)
		}
	}
}

// Stop shuts down the server and removes the socket file.
func (s *Server) Stop() {
	s.mu.Lock()
	s.stopped = true
	s.mu.Unlock()
	s.listener.Close()
	os.Remove(s.path)
}
