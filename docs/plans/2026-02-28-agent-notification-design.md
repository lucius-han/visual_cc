# Agent 실행 + Notification 디자인

**날짜:** 2026-02-28
**상태:** 승인됨

## 개요

두 가지 기능을 추가한다:
1. **Agent 실행 내용**: subagent 호출을 감지하고, 내부 tool 호출을 시각적으로 그룹핑
2. **알람 발생/확인 여부**: Notification 이벤트 표시 + tool permission 승인/거부 추적

## 접근법

- **Agent 그룹핑**: session_id 기반. 메인 session_id 외 새 session_id 이벤트 = subagent child
- **Permission 추적**: tool_use_id로 PreToolUse↔PostToolUse 매칭. 미매칭 = 거부됨
- **Notification 읽음**: 미확인 카운트 추적, TUI에서 `n` 키로 읽음 처리

## 데이터 모델 변경

### event.go — 추가 필드
```go
// HookPayload 추가
ToolUseID string `json:"tool_use_id,omitempty"`
Message   string `json:"message,omitempty"`

// Event 추가
ToolUseID string `json:"tool_use_id,omitempty"`
Message   string `json:"message,omitempty"`
```

### store.go — 내부 상태 추가
```go
mainSessionID string
agentSessions map[string]string  // subagent session_id → parent tool_use_id
pendingTools  map[string]Event   // tool_use_id → PreToolUse event
notifUnread   int

// Stats 추가 필드
ConfirmedTools int
DeniedTools    int
NotifTotal     int
NotifUnread    int
```

### store.Add() 로직
- 첫 이벤트: mainSessionID 설정
- PreToolUse:
  - tool_name == "Agent": agentSessions에 tool_use_id 등록 예약
  - 모든 PreToolUse: pendingTools[tool_use_id] = event
- PostToolUse:
  - pendingTools에서 tool_use_id 제거 → ConfirmedTools++
- Notification:
  - NotifTotal++, notifUnread++
- Stop:
  - pendingTools 잔여 항목 → DeniedTools += len(pendingTools)
  - pendingTools 초기화

### 이벤트 session_id 라우팅
- session_id != mainSessionID → isChildEvent = true
- isChildEvent인 이벤트는 렌더링 시 들여쓰기

## 렌더링

### Agent 이벤트
```
13:45:01  ⊕ Agent                     [general-purpose]
          │ Dispatch implementer subagent
          ├─ 13:45:02  ● Bash   go build ./...
          ├─ 13:45:02  ✓ Bash                  142ms
13:45:04  ✓ Agent                              3.2s
```

- `⊕` = 청록색 (#00B4D8), Bold
- child 이벤트: `├─` prefix + dim 색

### Permission 뱃지
```
● Bash   rm -rf /tmp              [승인 대기]   ← PreToolUse 도착
✓ Bash                   1ms     [승인됨]      ← PostToolUse 도착
● Bash   curl http://...          [승인 대기]
■ Session ended                   [거부됨 1건]  ← Stop 시 pending 정리
```

### Notification
```
⚡ Claude 작업이 완료되었습니다    [미확인]
```
- `n` 키: 미확인 notification 모두 읽음 처리

### 통계 패널 추가 섹션
```
├─ 알람 ──────────┤
│ ⚡ 알람   3건    │
│ ✓ 승인됨  8건   │
│ ✗ 거부됨  1건   │
│ 🔴 미확인  1건  │
```

## 키 바인딩 추가
| 키 | 동작 |
|----|------|
| `n` | 미확인 notification 모두 읽음 처리 |
