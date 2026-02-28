package event

import "time"

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
