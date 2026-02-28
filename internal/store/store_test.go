package store_test

import (
	"testing"
	"time"

	"github.com/lucius-han/visual_cc/internal/event"
	"github.com/lucius-han/visual_cc/internal/store"
)

func TestStore_Add_And_Stats(t *testing.T) {
	s := store.New(100)

	s.Add(event.Event{Type: event.TypePreToolUse, ToolName: "Bash", Timestamp: time.Now()})
	s.Add(event.Event{Type: event.TypePostToolUse, ToolName: "Bash", DurationMs: 142, Timestamp: time.Now()})
	s.Add(event.Event{Type: event.TypePreToolUse, ToolName: "Read", Timestamp: time.Now()})

	stats := s.Stats()
	if stats.ToolCounts["Bash"] != 2 {
		t.Errorf("expected Bash count 2, got %d", stats.ToolCounts["Bash"])
	}
	if stats.ToolCounts["Read"] != 1 {
		t.Errorf("expected Read count 1, got %d", stats.ToolCounts["Read"])
	}

	events := s.Events()
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

func TestStore_RingBuffer(t *testing.T) {
	s := store.New(2)
	s.Add(event.Event{ToolName: "A", Timestamp: time.Now()})
	s.Add(event.Event{ToolName: "B", Timestamp: time.Now()})
	s.Add(event.Event{ToolName: "C", Timestamp: time.Now()})

	events := s.Events()
	if len(events) != 2 {
		t.Errorf("expected 2 events (ring buffer), got %d", len(events))
	}
	if events[len(events)-1].ToolName != "C" {
		t.Errorf("expected last event C, got %s", events[len(events)-1].ToolName)
	}
}

func TestStore_Reset(t *testing.T) {
	s := store.New(100)
	s.Add(event.Event{ToolName: "Bash", Timestamp: time.Now()})
	s.Add(event.Event{ToolName: "Read", Timestamp: time.Now()})

	s.Reset()

	events := s.Events()
	if len(events) != 0 {
		t.Errorf("expected 0 events after reset, got %d", len(events))
	}
	stats := s.Stats()
	if len(stats.ToolCounts) != 0 {
		t.Errorf("expected empty tool counts after reset, got %v", stats.ToolCounts)
	}
}

func TestStore_Stats_ReturnsCopy(t *testing.T) {
	s := store.New(100)
	s.Add(event.Event{ToolName: "Bash", Timestamp: time.Now()})

	stats1 := s.Stats()
	stats1.ToolCounts["Bash"] = 999 // mutate the copy

	stats2 := s.Stats()
	if stats2.ToolCounts["Bash"] != 1 {
		t.Errorf("expected Stats() to return a copy, but internal state was mutated")
	}
}
