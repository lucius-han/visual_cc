package event

import "time"

// FromHookPayload converts a Claude Code hook payload to an internal Event.
// Claude Code sends PascalCase hook_event_name ("PreToolUse"), which is normalized
// to our internal snake_case constants.
func FromHookPayload(p HookPayload, t time.Time) Event {
	return Event{
		SessionID:  p.SessionID,
		Type:       normalizeType(p.Type),
		Timestamp:  t,
		ToolName:   p.ToolName,
		ToolInput:  p.ToolInput,
		ToolOutput: p.ToolOutput,
	}
}

// normalizeType maps Claude Code's PascalCase hook_event_name to our internal constants.
func normalizeType(t Type) Type {
	switch t {
	case "PreToolUse":
		return TypePreToolUse
	case "PostToolUse":
		return TypePostToolUse
	case "Stop":
		return TypeStop
	case "Notification":
		return TypeNotification
	default:
		return t
	}
}

type Type string

const (
	TypePreToolUse   Type = "pre_tool_use"
	TypePostToolUse  Type = "post_tool_use"
	TypeStop         Type = "stop"
	TypeNotification Type = "notification"
)

// HookPayload is what Claude Code sends to hook via stdin
type HookPayload struct {
	SessionID  string         `json:"session_id"`
	Type       Type           `json:"hook_event_name"`
	ToolName   string         `json:"tool_name,omitempty"`
	ToolInput  map[string]any `json:"tool_input,omitempty"`
	ToolOutput string         `json:"tool_response,omitempty"`
}

// Event is visual_cc's internal event representation
type Event struct {
	SessionID  string         `json:"session_id"`
	Type       Type           `json:"event_type"`
	Timestamp  time.Time      `json:"timestamp"`
	ToolName   string         `json:"tool_name,omitempty"`
	ToolInput  map[string]any `json:"tool_input,omitempty"`
	ToolOutput string         `json:"tool_output,omitempty"`
	DurationMs int64          `json:"duration_ms,omitempty"`
	IsError    bool           `json:"is_error,omitempty"`
}
