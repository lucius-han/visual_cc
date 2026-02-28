# Agent 실행 + Notification 구현 계획

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Claude Code Agent 도구 호출을 subagent 이벤트와 함께 그룹핑하여 표시하고, Notification 이벤트와 tool permission 승인/거부 상태를 TUI에서 추적한다.

**Architecture:** session_id로 subagent 이벤트 감지 → tool_use_id로 PreToolUse↔PostToolUse 매칭 → Stats에 집계 → 렌더링에 반영. event.go에 필드 추가, store.go에 추적 로직 추가, render.go 시그니처 변경(isChild 파라미터), statspanel.go에 알람 섹션 추가, model.go에 'n' 키 추가.

**Tech Stack:** Go 1.26, bubbletea, lipgloss (기존 스택 그대로)

---

## Task 1: event 모델에 ToolUseID, Message 필드 추가

**Files:**
- Modify: `internal/event/event.go`
- Modify: `internal/event/event_test.go`

**Step 1: 실패하는 테스트 작성**

`internal/event/event_test.go` 끝에 추가:
```go
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
```

**Step 2: 테스트 실행 (실패 확인)**
```bash
cd /Users/hans/IdeaProjects/visual_cc && go test ./internal/event/... -v
```
Expected: FAIL — `event.HookPayload{} unknown field ToolUseID`

**Step 3: HookPayload와 Event에 필드 추가**

`internal/event/event.go`의 `HookPayload` 구조체에 추가:
```go
ToolUseID string `json:"tool_use_id,omitempty"`
Message   string `json:"message,omitempty"`
```

`internal/event/event.go`의 `Event` 구조체에 추가:
```go
ToolUseID string `json:"tool_use_id,omitempty"`
Message   string `json:"message,omitempty"`
```

`FromHookPayload` 함수 업데이트:
```go
func FromHookPayload(p HookPayload, t time.Time) Event {
	return Event{
		SessionID:  p.SessionID,
		Type:       normalizeType(p.Type),
		Timestamp:  t,
		ToolName:   p.ToolName,
		ToolInput:  p.ToolInput,
		ToolOutput: p.ToolOutput,
		ToolUseID:  p.ToolUseID,
		Message:    p.Message,
	}
}
```

**Step 4: 테스트 통과 확인**
```bash
cd /Users/hans/IdeaProjects/visual_cc && go test ./internal/event/... -v
```
Expected: 6개 테스트 모두 PASS

**Step 5: 커밋**
```bash
cd /Users/hans/IdeaProjects/visual_cc
git add internal/event/event.go internal/event/event_test.go
git commit -m "feat: add ToolUseID and Message fields to event model"
```

---

## Task 2: Store — permission 추적 + notification 통계

**Files:**
- Modify: `internal/store/store.go`
- Modify: `internal/store/store_test.go`

**Step 1: 실패하는 테스트 작성**

`internal/store/store_test.go` 끝에 추가:
```go
func TestStore_Permission_ConfirmedOnPostToolUse(t *testing.T) {
	s := store.New(100)

	s.Add(event.Event{
		SessionID: "sess1", Type: event.TypePreToolUse,
		ToolName: "Bash", ToolUseID: "tu1", Timestamp: time.Now(),
	})

	stats := s.Stats()
	if stats.ConfirmedTools != 0 {
		t.Errorf("expected 0 confirmed before PostToolUse, got %d", stats.ConfirmedTools)
	}

	s.Add(event.Event{
		SessionID: "sess1", Type: event.TypePostToolUse,
		ToolName: "Bash", ToolUseID: "tu1", Timestamp: time.Now(),
	})

	stats = s.Stats()
	if stats.ConfirmedTools != 1 {
		t.Errorf("expected 1 confirmed, got %d", stats.ConfirmedTools)
	}
}

func TestStore_Permission_DeniedOnStop(t *testing.T) {
	s := store.New(100)

	s.Add(event.Event{
		SessionID: "sess1", Type: event.TypePreToolUse,
		ToolName: "Bash", ToolUseID: "tu2", Timestamp: time.Now(),
	})
	s.Add(event.Event{
		SessionID: "sess1", Type: event.TypeStop, Timestamp: time.Now(),
	})

	stats := s.Stats()
	if stats.DeniedTools != 1 {
		t.Errorf("expected 1 denied, got %d", stats.DeniedTools)
	}
	if stats.ConfirmedTools != 0 {
		t.Errorf("expected 0 confirmed, got %d", stats.ConfirmedTools)
	}
}

func TestStore_Notification_Stats(t *testing.T) {
	s := store.New(100)

	s.Add(event.Event{
		SessionID: "sess1", Type: event.TypeNotification,
		Message: "알람1", Timestamp: time.Now(),
	})
	s.Add(event.Event{
		SessionID: "sess1", Type: event.TypeNotification,
		Message: "알람2", Timestamp: time.Now(),
	})

	stats := s.Stats()
	if stats.NotifTotal != 2 {
		t.Errorf("expected NotifTotal 2, got %d", stats.NotifTotal)
	}
	if stats.NotifUnread != 2 {
		t.Errorf("expected NotifUnread 2, got %d", stats.NotifUnread)
	}

	s.MarkNotifsRead()

	stats = s.Stats()
	if stats.NotifUnread != 0 {
		t.Errorf("expected NotifUnread 0 after MarkNotifsRead, got %d", stats.NotifUnread)
	}
}

func TestStore_MainSessionID(t *testing.T) {
	s := store.New(100)

	if s.MainSessionID() != "" {
		t.Error("expected empty MainSessionID before any event")
	}

	s.Add(event.Event{SessionID: "main-sess", Type: event.TypePreToolUse, Timestamp: time.Now()})
	s.Add(event.Event{SessionID: "sub-sess", Type: event.TypePreToolUse, Timestamp: time.Now()})

	if s.MainSessionID() != "main-sess" {
		t.Errorf("expected main-sess, got %q", s.MainSessionID())
	}
	if !s.IsSubagentSession("sub-sess") {
		t.Error("expected sub-sess to be a subagent session")
	}
	if s.IsSubagentSession("main-sess") {
		t.Error("expected main-sess NOT to be a subagent session")
	}
}
```

**Step 2: 테스트 실행 (실패 확인)**
```bash
cd /Users/hans/IdeaProjects/visual_cc && go test ./internal/store/... -v
```
Expected: FAIL — `Stats has no field ConfirmedTools`, `MarkNotifsRead undefined` 등

**Step 3: store.go 업데이트**

`Stats` 구조체에 추가:
```go
ConfirmedTools int
DeniedTools    int
NotifTotal     int
NotifUnread    int
```

`Store` 구조체에 추가:
```go
mainSessionID  string
subSessions    map[string]bool  // subagent session IDs
pendingTools   map[string]bool  // tool_use_id set of pending PreToolUse
```

`New()` 함수 업데이트:
```go
func New(capacity int) *Store {
	return &Store{
		buf:         make([]event.Event, capacity),
		cap:         capacity,
		subSessions: make(map[string]bool),
		pendingTools: make(map[string]bool),
		stats:       Stats{ToolCounts: make(map[string]int), StartTime: time.Now()},
	}
}
```

`Add()` 함수 — 기존 내용 유지하고 아래 로직 추가 (if e.ToolName 블록 뒤):
```go
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
```

새 메서드 추가:
```go
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
```

`Stats()` 반환값에 새 필드 추가:
```go
return Stats{
    ToolCounts:    counts,
    TotalTokens:   s.stats.TotalTokens,
    TotalCostUSD:  s.stats.TotalCostUSD,
    StartTime:     s.stats.StartTime,
    ConfirmedTools: s.stats.ConfirmedTools,
    DeniedTools:   s.stats.DeniedTools,
    NotifTotal:    s.stats.NotifTotal,
    NotifUnread:   s.stats.NotifUnread,
}
```

`Reset()` 업데이트:
```go
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
```

**Step 4: 테스트 통과 확인**
```bash
cd /Users/hans/IdeaProjects/visual_cc && go test ./internal/store/... -v
```
Expected: 8개 테스트 모두 PASS

**Step 5: 커밋**
```bash
cd /Users/hans/IdeaProjects/visual_cc
git add internal/store/store.go internal/store/store_test.go
git commit -m "feat: add permission tracking and notification stats to store"
```

---

## Task 3: 스타일 + 렌더링 업데이트

**Files:**
- Modify: `internal/tui/styles.go`
- Modify: `internal/tui/render.go`
- Modify: `internal/tui/render_test.go`

### Step 1: styles.go에 새 스타일 추가

`internal/tui/styles.go`의 var 블록에 추가:
```go
colorTeal  = lipgloss.Color("#00B4D8")
colorChild = lipgloss.Color("#555555")

styleAgent      = lipgloss.NewStyle().Foreground(colorTeal).Bold(true)
styleChildPrefix = lipgloss.NewStyle().Foreground(colorChild)
styleChildEvent  = lipgloss.NewStyle().Foreground(colorChild)
styleBadgeOK    = lipgloss.NewStyle().Foreground(colorGreen)
styleBadgePend  = lipgloss.NewStyle().Foreground(colorGray)
styleBadgeUnread = lipgloss.NewStyle().Foreground(colorYellow)
```

### Step 2: 실패하는 테스트 작성

`internal/tui/render_test.go` 전체를 다음으로 교체 (isChild 파라미터 추가):
```go
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
	if tui.RenderEvent(e, false) == "" {
		t.Error("expected non-empty render output for PreToolUse")
	}
}

func TestRenderEvent_PreToolUse_AsChild(t *testing.T) {
	e := event.Event{
		Type:      event.TypePreToolUse,
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "ls"},
		Timestamp: time.Now(),
	}
	line := tui.RenderEvent(e, true)
	if line == "" {
		t.Error("expected non-empty render output for child PreToolUse")
	}
}

func TestRenderEvent_AgentToolCall(t *testing.T) {
	e := event.Event{
		Type:     event.TypePreToolUse,
		ToolName: "Agent",
		ToolInput: map[string]any{
			"description":    "Task 1: 스캐폴딩",
			"subagent_type":  "general-purpose",
		},
		Timestamp: time.Now(),
	}
	line := tui.RenderEvent(e, false)
	if line == "" {
		t.Error("expected non-empty render output for Agent tool call")
	}
}

func TestRenderEvent_PostToolUse_Success(t *testing.T) {
	e := event.Event{
		Type:       event.TypePostToolUse,
		ToolName:   "Bash",
		DurationMs: 142,
		Timestamp:  time.Now(),
	}
	if tui.RenderEvent(e, false) == "" {
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
	if tui.RenderEvent(e, false) == "" {
		t.Error("expected non-empty render output for PostToolUse error")
	}
}

func TestRenderEvent_Stop(t *testing.T) {
	e := event.Event{Type: event.TypeStop, Timestamp: time.Now()}
	if tui.RenderEvent(e, false) == "" {
		t.Error("expected non-empty render output for Stop")
	}
}

func TestRenderEvent_Notification_WithMessage(t *testing.T) {
	e := event.Event{
		Type:      event.TypeNotification,
		Message:   "작업이 완료되었습니다",
		Timestamp: time.Now(),
	}
	if tui.RenderEvent(e, false) == "" {
		t.Error("expected non-empty render output for Notification")
	}
}

func TestRenderEvent_UnknownType_EmptyOutput(t *testing.T) {
	e := event.Event{Type: event.Type("unknown_type"), Timestamp: time.Now()}
	if tui.RenderEvent(e, false) != "" {
		t.Errorf("expected empty output for unknown event type")
	}
}
```

**Step 3: 테스트 실행 (실패 확인)**
```bash
cd /Users/hans/IdeaProjects/visual_cc && go test ./internal/tui/... -v
```
Expected: FAIL — `tui.RenderEvent` takes 1 argument but 2 given

**Step 4: render.go 업데이트**

`internal/tui/render.go` 전체 교체:
```go
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
	desc := ""
	for _, key := range []string{"description", "prompt"} {
		if v, ok := e.ToolInput[key]; ok {
			desc = firstLine(fmt.Sprintf("%v", v), 55)
			break
		}
	}
	if desc != "" {
		sb.WriteString(styleIndent.Render("│ "+desc) + "\n")
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
		// Other child events (stop, notification) rendered normally but dimmed
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
```

**Step 5: 테스트 통과 확인**
```bash
cd /Users/hans/IdeaProjects/visual_cc && go test ./internal/tui/... -v
```
Expected: 8개 테스트 모두 PASS

**Step 6: 커밋**
```bash
cd /Users/hans/IdeaProjects/visual_cc
git add internal/tui/styles.go internal/tui/render.go internal/tui/render_test.go
git commit -m "feat: add agent grouping and notification rendering"
```

---

## Task 4: 통계 패널에 알람 섹션 추가

**Files:**
- Modify: `internal/tui/statspanel.go`

**Step 1: RenderStatsPanel 업데이트**

`internal/tui/statspanel.go`의 `RenderStatsPanel` 함수에서 하단 도움말 직전에 알람 섹션 추가. `for len(lines) < height-4` 블록과 최종 lines append 사이를 다음으로 교체:

```go
// 알람 섹션
lines = append(lines,
    "",
    styleDim.Render("알람"),
    styleDivider.Render(strings.Repeat("─", panelWidth-2)),
    fmt.Sprintf("⚡ 알람   %s건", formatNumber(stats.NotifTotal)),
    styleBadgeOK.Render(fmt.Sprintf("✓ 승인됨  %d건", stats.ConfirmedTools)),
)
if stats.DeniedTools > 0 {
    lines = append(lines, stylePostErr.Render(fmt.Sprintf("✗ 거부됨  %d건", stats.DeniedTools)))
} else {
    lines = append(lines, fmt.Sprintf("✗ 거부됨  %d건", stats.DeniedTools))
}
if stats.NotifUnread > 0 {
    lines = append(lines, styleBadgeUnread.Render(fmt.Sprintf("🔴 미확인  %d건", stats.NotifUnread)))
} else {
    lines = append(lines, styleDim.Render("   미확인  0건"))
}
```

기존 padding 루프 (`for len(lines) < height-4`) 제거 또는 높이 기준 조정.

최종 lines 마지막:
```go
lines = append(lines,
    styleDivider.Render(strings.Repeat("─", panelWidth-2)),
    styleHelp.Render("q quit  c clear  n read"),
)
```

**Step 2: 빌드 확인**
```bash
cd /Users/hans/IdeaProjects/visual_cc && go build ./internal/tui/...
```
Expected: 에러 없음

**Step 3: 커밋**
```bash
cd /Users/hans/IdeaProjects/visual_cc
git add internal/tui/statspanel.go
git commit -m "feat: add notification/permission section to stats panel"
```

---

## Task 5: model.go 업데이트 + 최종 wiring

**Files:**
- Modify: `internal/tui/model.go`

**Step 1: model.go 업데이트**

두 가지 변경:

1. `NewEventMsg` 핸들러에서 isChild 계산 후 RenderEvent에 전달:
```go
case NewEventMsg:
    e := event.Event(msg)
    m.store.Add(e)
    mainID := m.store.MainSessionID()
    isChild := mainID != "" && e.SessionID != "" && e.SessionID != mainID
    rendered := RenderEvent(e, isChild)
    if rendered != "" {
        m.logBuf.WriteString(rendered)
        m.viewport.SetContent(m.logBuf.String())
        if m.autoScroll {
            m.viewport.GotoBottom()
        }
    }
```

2. `"c"` 케이스 이후에 `"n"` 케이스 추가:
```go
case "n":
    m.store.MarkNotifsRead()
    return m, nil
```

3. `renderHeader`의 right 힌트 업데이트:
```go
right := styleDim.Render("q quit  ↑↓ scroll  G bottom  c clear  n read ")
```

**Step 2: 전체 빌드 및 테스트**
```bash
cd /Users/hans/IdeaProjects/visual_cc && go build ./... && go test ./... -v
```
Expected: 모든 테스트 PASS, 빌드 성공

**Step 3: 바이너리 재빌드**
```bash
cd /Users/hans/IdeaProjects/visual_cc
go build -o visual_cc ./cmd/visual_cc/
go build -o visual_cc-hook ./cmd/visual_cc-hook/
```

**Step 4: 커밋**
```bash
cd /Users/hans/IdeaProjects/visual_cc
git add internal/tui/model.go visual_cc visual_cc-hook
git commit -m "feat: wire isChild rendering and notification read key"
```
