package store

import (
	"sync"
	"time"

	"github.com/lucius-han/visual_cc/internal/event"
)

// Stats holds aggregated statistics for a session.
type Stats struct {
	ToolCounts     map[string]int
	TotalTokens    int
	TotalCostUSD   float64
	StartTime      time.Time
	ConfirmedTools int
	DeniedTools    int
	NotifTotal     int
	NotifUnread    int
}

// Store is a thread-safe ring buffer of events with aggregated statistics.
type Store struct {
	mu            sync.RWMutex
	buf           []event.Event
	cap           int
	head          int
	count         int
	stats         Stats
	mainSessionID string
	subSessions   map[string]bool // subagent session IDs
	pendingTools  map[string]bool // tool_use_id set of pending PreToolUse
}

// New creates a Store with the given ring buffer capacity.
func New(capacity int) *Store {
	return &Store{
		buf:          make([]event.Event, capacity),
		cap:          capacity,
		subSessions:  make(map[string]bool),
		pendingTools: make(map[string]bool),
		stats:        Stats{ToolCounts: make(map[string]int), StartTime: time.Now()},
	}
}

// Add appends an event to the ring buffer and updates statistics.
func (s *Store) Add(e event.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buf[s.head%s.cap] = e
	s.head++
	if s.count < s.cap {
		s.count++
	}

	if e.ToolName != "" {
		s.stats.ToolCounts[e.ToolName]++
	}

	// Track main session (first session seen)
	if s.mainSessionID == "" && e.SessionID != "" {
		s.mainSessionID = e.SessionID
	} else if e.SessionID != "" && e.SessionID != s.mainSessionID {
		s.subSessions[e.SessionID] = true
	}

	// Permission tracking
	switch e.Type {
	case event.TypePreToolUse:
		if e.ToolUseID != "" {
			s.pendingTools[e.ToolUseID] = true
		}
	case event.TypePostToolUse:
		if e.ToolUseID != "" {
			if s.pendingTools[e.ToolUseID] {
				delete(s.pendingTools, e.ToolUseID)
				s.stats.ConfirmedTools++
			}
		}
	case event.TypeNotification:
		s.stats.NotifTotal++
		s.stats.NotifUnread++
	case event.TypeStop:
		s.stats.DeniedTools += len(s.pendingTools)
		s.pendingTools = make(map[string]bool)
	}
}

// Events returns all events in the ring buffer in insertion order.
func (s *Store) Events() []event.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]event.Event, s.count)
	start := 0
	if s.head > s.cap {
		start = s.head % s.cap
	}
	for i := 0; i < s.count; i++ {
		result[i] = s.buf[(start+i)%s.cap]
	}
	return result
}

// Stats returns a copy of current aggregated statistics.
func (s *Store) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[string]int, len(s.stats.ToolCounts))
	for k, v := range s.stats.ToolCounts {
		counts[k] = v
	}
	return Stats{
		ToolCounts:     counts,
		TotalTokens:    s.stats.TotalTokens,
		TotalCostUSD:   s.stats.TotalCostUSD,
		StartTime:      s.stats.StartTime,
		ConfirmedTools: s.stats.ConfirmedTools,
		DeniedTools:    s.stats.DeniedTools,
		NotifTotal:     s.stats.NotifTotal,
		NotifUnread:    s.stats.NotifUnread,
	}
}

// Reset clears all events and statistics, resetting the start time.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buf = make([]event.Event, s.cap)
	s.head = 0
	s.count = 0
	s.mainSessionID = ""
	s.subSessions = make(map[string]bool)
	s.pendingTools = make(map[string]bool)
	s.stats = Stats{ToolCounts: make(map[string]int), StartTime: time.Now()}
}

// MarkNotifsRead resets the unread notification counter.
func (s *Store) MarkNotifsRead() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.NotifUnread = 0
}

// MainSessionID returns the first session ID seen.
func (s *Store) MainSessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mainSessionID
}

// IsSubagentSession returns true if sessionID is a known subagent session.
func (s *Store) IsSubagentSession(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.subSessions[sessionID]
}
