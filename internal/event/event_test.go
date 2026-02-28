package event_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lucius-han/visual_cc/internal/event"
)

func TestFromHookPayload_PreToolUse(t *testing.T) {
	payload := event.HookPayload{
		SessionID: "sess1",
		Type:      event.TypePreToolUse,
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "ls -la"},
	}
	now := time.Now()
	e := event.FromHookPayload(payload, now)

	if e.SessionID != "sess1" {
		t.Errorf("expected session_id sess1, got %s", e.SessionID)
	}
	if e.ToolName != "Bash" {
		t.Errorf("expected tool_name Bash, got %s", e.ToolName)
	}
	if e.Type != event.TypePreToolUse {
		t.Errorf("expected pre_tool_use, got %s", e.Type)
	}
	if e.Timestamp != now {
		t.Errorf("expected timestamp to be set")
	}
	_, err := json.Marshal(e)
	if err != nil {
		t.Errorf("expected event to be JSON serializable: %v", err)
	}
}

func TestFromHookPayload_PostToolUse_WithOutput(t *testing.T) {
	payload := event.HookPayload{
		SessionID:  "sess2",
		Type:       event.TypePostToolUse,
		ToolName:   "Read",
		ToolOutput: "file contents here",
	}
	now := time.Now()
	e := event.FromHookPayload(payload, now)

	if e.Type != event.TypePostToolUse {
		t.Errorf("expected post_tool_use, got %s", e.Type)
	}
	if e.ToolOutput != "file contents here" {
		t.Errorf("expected tool_output to be set, got %q", e.ToolOutput)
	}
	if e.ToolName != "Read" {
		t.Errorf("expected tool_name Read, got %s", e.ToolName)
	}
}

func TestFromHookPayload_Stop_MinimalPayload(t *testing.T) {
	payload := event.HookPayload{
		SessionID: "sess3",
		Type:      event.TypeStop,
	}
	now := time.Now()
	e := event.FromHookPayload(payload, now)

	if e.Type != event.TypeStop {
		t.Errorf("expected stop, got %s", e.Type)
	}
	if e.SessionID != "sess3" {
		t.Errorf("expected session_id sess3, got %s", e.SessionID)
	}
	if e.ToolName != "" {
		t.Errorf("expected empty tool_name, got %s", e.ToolName)
	}
	if e.ToolInput != nil {
		t.Errorf("expected nil tool_input, got %v", e.ToolInput)
	}
}

func TestFromHookPayload_ToolInput_Contents(t *testing.T) {
	input := map[string]any{"command": "ls -la", "timeout": float64(30)}
	payload := event.HookPayload{
		SessionID: "sess4",
		Type:      event.TypePreToolUse,
		ToolName:  "Bash",
		ToolInput: input,
	}
	e := event.FromHookPayload(payload, time.Now())

	if e.ToolInput == nil {
		t.Fatal("expected tool_input to be set")
	}
	if e.ToolInput["command"] != "ls -la" {
		t.Errorf("expected command 'ls -la', got %v", e.ToolInput["command"])
	}
}

func TestFromHookPayload_ToolUseID(t *testing.T) {
	payload := event.HookPayload{
		SessionID: "sess1",
		Type:      event.TypePreToolUse,
		ToolName:  "Bash",
		ToolUseID: "toolu_abc123",
		ToolInput: map[string]any{"command": "ls"},
	}
	e := event.FromHookPayload(payload, time.Now())
	if e.ToolUseID != "toolu_abc123" {
		t.Errorf("expected ToolUseID toolu_abc123, got %q", e.ToolUseID)
	}
}

func TestFromHookPayload_NotificationMessage(t *testing.T) {
	payload := event.HookPayload{
		SessionID: "sess1",
		Type:      event.TypeNotification,
		Message:   "작업이 완료되었습니다",
	}
	e := event.FromHookPayload(payload, time.Now())
	if e.Message != "작업이 완료되었습니다" {
		t.Errorf("expected Message set, got %q", e.Message)
	}
}
