package tui

import (
	"fmt"
	"strings"

	"github.com/lucius-han/visual_cc/internal/event"
)

// RenderEvent returns a styled string representation of an event for the log view.
// Returns empty string for unknown event types.
func RenderEvent(e event.Event) string {
	ts := styleTime.Render(e.Timestamp.Format("15:04:05"))
	var sb strings.Builder

	switch e.Type {
	case event.TypePreToolUse:
		icon := stylePreTool.Render("●")
		name := stylePreTool.Render(e.ToolName)
		sb.WriteString(fmt.Sprintf("  %s  %s %s\n", ts, icon, name))
		if input := formatInput(e); input != "" {
			sb.WriteString(styleIndent.Render("│ "+input) + "\n")
		}

	case event.TypePostToolUse:
		if e.IsError {
			icon := stylePostErr.Render("✗")
			name := stylePostErr.Render(e.ToolName)
			dur := styleDuration.Render("error")
			sb.WriteString(fmt.Sprintf("  %s  %s %-16s %s\n", ts, icon, name, dur))
			if e.ToolOutput != "" {
				short := firstLine(e.ToolOutput, 60)
				sb.WriteString(styleIndent.Render("│ "+short) + "\n")
			}
		} else {
			icon := stylePostOK.Render("✓")
			name := stylePostOK.Render(e.ToolName)
			dur := styleDuration.Width(8).Render(fmt.Sprintf("%dms", e.DurationMs))
			sb.WriteString(fmt.Sprintf("  %s  %s %-16s %s\n", ts, icon, name, dur))
		}

	case event.TypeStop:
		icon := styleStop.Render("■")
		sb.WriteString(fmt.Sprintf("  %s  %s %s\n", ts, icon, styleStop.Render("Session ended")))

	case event.TypeNotification:
		icon := styleNotif.Render("⚡")
		sb.WriteString(fmt.Sprintf("  %s  %s %s\n", ts, icon, styleNotif.Render(e.ToolOutput)))

	default:
		return ""
	}

	return sb.String()
}

func formatInput(e event.Event) string {
	if e.ToolInput == nil {
		return ""
	}
	for _, key := range []string{"command", "file_path", "pattern", "old_string"} {
		if v, ok := e.ToolInput[key]; ok {
			return firstLine(fmt.Sprintf("%v", v), 60)
		}
	}
	return ""
}

func firstLine(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[:idx]
	}
	if len(s) > maxLen {
		return s[:maxLen] + styleDim.Render("…")
	}
	return s
}
