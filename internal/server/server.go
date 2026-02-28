package server

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"sync"

	"github.com/lucius-han/visual_cc/internal/event"
)

// Server listens on a Unix socket and dispatches received events to handler.
type Server struct {
	path     string
	handler  func(event.Event)
	listener net.Listener
	mu       sync.Mutex
	stopped  bool
}

// New creates a new Server listening on socketPath.
// The handler is called for each successfully parsed event.
func New(socketPath string, handler func(event.Event)) (*Server, error) {
	os.Remove(socketPath)
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return &Server{path: socketPath, handler: handler, listener: l}, nil
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
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
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
