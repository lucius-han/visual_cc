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
