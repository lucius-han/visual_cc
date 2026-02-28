package tui

import (
	"fmt"
	"strings"

	"github.com/lucius-han/visual_cc/internal/event"
)

// RenderEvent returns a styled string representation of an event for the log view.
// isChild=true renders with indentation for subagent events.
// Returns empty string for unknown event types.
func RenderEvent(e event.Event, isChild bool) string {
	ts := styleTime.Render(e.Timestamp.Format("15:04:05"))
	var sb strings.Builder

	if isChild {
		return renderChildEvent(e, ts)
	}

	switch e.Type {
	case event.TypePreToolUse:
		if e.ToolName == "Agent" {
			return renderAgentStart(e, ts)
		}
		icon := stylePreTool.Render("●")
		name := stylePreTool.Render(e.ToolName)
		sb.WriteString(fmt.Sprintf("  %s  %s %s\n", ts, icon, name))
		if input := formatInput(e); input != "" {
			sb.WriteString(styleIndent.Render("│ "+input) + "\n")
		}

	case event.TypePostToolUse:
		if e.ToolName == "Agent" {
			return renderAgentEnd(e, ts)
		}
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
		msg := e.Message
		if msg == "" {
			msg = e.ToolOutput
		}
		badge := styleBadgeUnread.Render("[미확인]")
		sb.WriteString(fmt.Sprintf("  %s  %s %-32s %s\n", ts, icon, styleNotif.Render(msg), badge))

	default:
		return ""
	}

	return sb.String()
}

func renderAgentStart(e event.Event, ts string) string {
	icon := styleAgent.Render("⊕")
	name := styleAgent.Render("Agent")
	subtype := ""
	if v, ok := e.ToolInput["subagent_type"]; ok {
		subtype = styleBadgePend.Render(fmt.Sprintf("[%v]", v))
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  %s  %s %s %s\n", ts, icon, name, subtype))
	for _, key := range []string{"description", "prompt"} {
		if v, ok := e.ToolInput[key]; ok {
			desc := firstLine(fmt.Sprintf("%v", v), 55)
			sb.WriteString(styleIndent.Render("│ "+desc) + "\n")
			break
		}
	}
	return sb.String()
}

func renderAgentEnd(e event.Event, ts string) string {
	icon := stylePostOK.Render("✓")
	name := stylePostOK.Render("Agent")
	dur := styleDuration.Width(8).Render(fmt.Sprintf("%dms", e.DurationMs))
	return fmt.Sprintf("  %s  %s %-16s %s\n", ts, icon, name, dur)
}

func renderChildEvent(e event.Event, ts string) string {
	prefix := styleChildPrefix.Render("   ├─")
	var sb strings.Builder

	switch e.Type {
	case event.TypePreToolUse:
		icon := styleChildEvent.Render("●")
		name := styleChildEvent.Render(e.ToolName)
		sb.WriteString(fmt.Sprintf("%s %s  %s %s\n", prefix, ts, icon, name))
		if input := formatInput(e); input != "" {
			sb.WriteString(styleChildEvent.Render("   │       │ "+input) + "\n")
		}
	case event.TypePostToolUse:
		if e.IsError {
			icon := stylePostErr.Render("✗")
			name := stylePostErr.Render(e.ToolName)
			sb.WriteString(fmt.Sprintf("%s %s  %s %-16s\n", prefix, ts, icon, name))
		} else {
			icon := styleChildEvent.Render("✓")
			name := styleChildEvent.Render(e.ToolName)
			dur := styleChildEvent.Render(fmt.Sprintf("%dms", e.DurationMs))
			sb.WriteString(fmt.Sprintf("%s %s  %s %-16s %s\n", prefix, ts, icon, name, dur))
		}
	default:
		return styleChildEvent.Render(fmt.Sprintf("   │  %s  %s\n", ts, string(e.Type)))
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
