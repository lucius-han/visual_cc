package tui_test

import (
	"testing"
	"time"

	"github.com/lucius-han/visual_cc/internal/event"
	"github.com/lucius-han/visual_cc/internal/tui"
)

func TestRenderEvent_PreToolUse(t *testing.T) {
	e := event.Event{
		Type:      event.TypePreToolUse,
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "ls -la"},
		Timestamp: time.Date(2026, 2, 28, 13, 45, 1, 0, time.UTC),
	}
	line := tui.RenderEvent(e)
	if line == "" {
		t.Error("expected non-empty render output for PreToolUse")
	}
}

func TestRenderEvent_PostToolUse_Success(t *testing.T) {
	e := event.Event{
		Type:       event.TypePostToolUse,
		ToolName:   "Bash",
		DurationMs: 142,
		Timestamp:  time.Now(),
	}
	line := tui.RenderEvent(e)
	if line == "" {
		t.Error("expected non-empty render output for PostToolUse success")
	}
}

func TestRenderEvent_PostToolUse_Error(t *testing.T) {
	e := event.Event{
		Type:       event.TypePostToolUse,
		ToolName:   "Bash",
		IsError:    true,
		ToolOutput: "permission denied",
		Timestamp:  time.Now(),
	}
	line := tui.RenderEvent(e)
	if line == "" {
		t.Error("expected non-empty render output for PostToolUse error")
	}
}

func TestRenderEvent_Stop(t *testing.T) {
	e := event.Event{
		Type:      event.TypeStop,
		Timestamp: time.Now(),
	}
	line := tui.RenderEvent(e)
	if line == "" {
		t.Error("expected non-empty render output for Stop")
	}
}

func TestRenderEvent_UnknownType_EmptyOutput(t *testing.T) {
	e := event.Event{
		Type:      event.Type("unknown_type"),
		Timestamp: time.Now(),
	}
	line := tui.RenderEvent(e)
	// unknown type should return empty string (no render)
	if line != "" {
		t.Errorf("expected empty output for unknown event type, got %q", line)
	}
}
