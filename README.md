# visual_cc

Claude Code 실행을 실시간으로 모니터링하는 로컬 TUI 도구.

Claude Code hooks(PreToolUse / PostToolUse / Stop / Notification)가 이벤트를 Unix socket으로 전송하고,
`visual_cc`가 소켓 서버 + bubbletea TUI를 동시에 실행하며 실시간 로그와 통계를 표시합니다.

```
 visual_cc                          q quit  ↑↓ scroll  G bottom  c clear  n read
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  13:45:01  ● Bash                              ╭─ 통계 ─────────────────╮
            │ ls -la /src                       │ ⏱  4m 32s              │
  13:45:01  ✓ Bash                    142ms     │ 🔤 0 tokens             │
                                                │ 💰 $0.0000              │
  13:45:03  ⊕ Agent [general-purpose]           ├─ Tool 호출 ────────────┤
            │ Task 3: 스타일 업데이트            │ Bash   ████ 12         │
   ├─ 13:45:04  ● Read                          │ Read   ███  8          │
   │       │ /src/styles.go                     │ Edit   █    3          │
   ├─ 13:45:04  ✓ Read              8ms         ├─ 알람 / Permission ────┤
   ├─ 13:45:05  ● Edit                          │ ⚡ 알람   2건           │
   ├─ 13:45:06  ✓ Edit             34ms         │ ✓ 승인됨  8건           │
  13:45:07  ✓ Agent              3421ms         │ ✗ 거부됨  1건           │
                                                │    미확인  0건          │
  13:45:08  ⚡ 작업이 완료되었습니다  [미확인]   ╰────────────────────────╯
```

## 버전

- **v0.2.1** — 보안 강화
  - Unix 소켓 파일 권한 `0600` (owner-only) 적용 — 타 로컬 유저 이벤트 주입 차단
  - 유저별 소켓 경로 `/tmp/visual_cc-<uid>.sock` — 다중 유저 환경 충돌 및 symlink race 방지
  - 동시 접속 수 16개 제한 (semaphore) — goroutine 폭증 DoS 방지
  - 소켓 읽기 타임아웃 2초 — 연결 후 멈춘 클라이언트 고루틴 누수 방지
  - Scanner 버퍼 상한 16KB — 메모리 소진 방지
  - ANSI escape / 제어문자 sanitize — 터미널 인젝션 방지
  - `ToolOutput` 4KB, `Message` 512B 상한 — 링 버퍼 메모리 소진 방지
  - `pendingTools` / `subSessions` 맵 크기 상한 (1000 / 100)
  - `tea.Program` 초기화를 서버 고루틴 시작 전으로 이동 — 포인터 race condition 제거
  - `install.sh` 경로 공백 감지 후 경고 종료
- **v0.2.0** — Agent 그룹핑 + Notification/Permission 추적
  - subagent 이벤트를 session_id로 감지하여 부모 Agent 호출 하위에 들여쓰기로 표시
  - Notification 이벤트 표시 + `n` 키로 읽음 처리
  - tool_use_id 기반 PreToolUse↔PostToolUse 매칭으로 permission 승인/거부 추적
  - 통계 패널에 알람/Permission 섹션 추가
- **v0.1.0** — 기본 모니터링 (PreToolUse / PostToolUse / Stop / Notification 이벤트, Tool 호출 통계)

## 요구사항

- Go 1.22 이상
- Claude Code CLI (`claude`) 설치 및 hook 지원 버전

## 설치 및 빌드

```bash
git clone https://github.com/lucius-han/visual_cc.git
cd visual_cc
go build -o visual_cc     ./cmd/visual_cc/
go build -o visual_cc-hook ./cmd/visual_cc-hook/
```

빌드 후 두 바이너리(`visual_cc`, `visual_cc-hook`)를 PATH에 등록된 디렉토리(예: `/usr/local/bin`)에 복사합니다.

```bash
sudo cp visual_cc visual_cc-hook /usr/local/bin/
```

## Claude Code Hook 등록

Claude Code가 tool을 실행할 때마다 `visual_cc-hook`을 호출하도록 설정합니다.

`~/.claude/settings.json`에 아래 내용을 추가하세요:

```json
{
  "hooks": {
    "PreToolUse": [
      { "hooks": [{ "type": "command", "command": "visual_cc-hook" }] }
    ],
    "PostToolUse": [
      { "hooks": [{ "type": "command", "command": "visual_cc-hook" }] }
    ],
    "Stop": [
      { "hooks": [{ "type": "command", "command": "visual_cc-hook" }] }
    ],
    "Notification": [
      { "hooks": [{ "type": "command", "command": "visual_cc-hook" }] }
    ]
  }
}
```

> 기존 hooks 설정이 있다면 해당 배열에 항목을 추가하세요.

편의 스크립트를 사용할 수도 있습니다:

```bash
bash hooks/install.sh
```

## 사용법

### TUI 실행

별도 터미널에서 `visual_cc`를 먼저 실행합니다:

```bash
visual_cc
```

이후 Claude Code를 평소처럼 사용하면 hook이 자동으로 이벤트를 전송하고 TUI에 실시간으로 표시됩니다.

### 키보드 단축키

| 키 | 동작 |
|----|------|
| `q` / `Ctrl+C` | 종료 |
| `↑` / `↓` / `k` / `j` | 스크롤 |
| `G` | 최신 이벤트로 이동 (자동 스크롤 재개) |
| `c` | 로그 및 통계 초기화 |
| `n` | 미확인 알람을 읽음으로 처리 |

### 이벤트 색상 및 아이콘

| 이벤트 | 색상 | 아이콘 |
|--------|------|--------|
| PreToolUse | 파란색 | `●` |
| PostToolUse 성공 | 초록색 | `✓` |
| PostToolUse 실패 | 빨간색 | `✗` |
| Agent 시작 | 청록색 | `⊕` |
| Subagent 이벤트 | 회색 (들여쓰기) | `├─` |
| Stop | 보라색 | `■` |
| Notification | 노란색 | `⚡` |

## 아키텍처

```
Claude Code 세션
  └─ PreToolUse / PostToolUse / Stop / Notification hooks
       └─ JSON event → visual_cc-hook (stdin 읽어 Unix socket 전송)
            └─ /tmp/visual_cc-<uid>.sock  (유저별 격리)
                 └─ visual_cc (Go binary)
                      ├─ Socket Server (최대 16 동시접속, 읽기 타임아웃 2s)
                      │    └─ Event Store (ring buffer cap=500)
                      │         ├─ session_id로 subagent 감지
                      │         └─ tool_use_id로 permission 추적
                      └─ TUI (bubbletea)
                           ├─ 실시간 로그 뷰 (좌측, Agent 그룹핑)
                           └─ 통계 패널 (우측, 알람/Permission 섹션)
```

### 프로젝트 구조

```
visual_cc/
├── cmd/
│   ├── visual_cc/         # TUI 메인 바이너리
│   └── visual_cc-hook/    # hook 이벤트 전송 바이너리
├── internal/
│   ├── event/             # HookPayload, Event 타입, FromHookPayload
│   ├── server/            # Unix socket 서버 (보안 강화)
│   ├── socket/            # 유저별 소켓 경로 헬퍼
│   ├── store/             # ring buffer + 통계/permission/notification 집계
│   └── tui/               # styles, render(Agent 그룹핑), statspanel, model
├── hooks/
│   └── install.sh         # hook 등록 안내 스크립트
└── docs/plans/            # 설계 및 구현 계획 문서
```

## 보안

visual_cc는 로컬 전용 모니터링 도구이며, 아래 보안 조치가 적용되어 있습니다:

| 항목 | 내용 |
|------|------|
| 소켓 권한 | `0600` (owner-only) — 타 유저 접근 차단 |
| 소켓 경로 | `/tmp/visual_cc-<uid>.sock` — 유저별 격리 |
| 동시 접속 | 최대 16개 — goroutine DoS 방지 |
| 읽기 타임아웃 | 2초 — 고루틴 누수 방지 |
| 이벤트 크기 | 16KB/line, ToolOutput 4KB, Message 512B 상한 |
| 출력 sanitize | ANSI escape 및 제어문자 제거 |

> **알려진 제한:** session_id / tool_use_id 기반 구분은 인증 없이 작동합니다. 단독 개발 머신 용도를 전제로 설계되었으며, 다중 신뢰할 수 없는 유저 환경에는 적합하지 않습니다.

## 개발

```bash
# 테스트
go test ./...

# race condition 검사 포함 테스트
go test -race ./...

# 전체 빌드
go build ./...
```

## 라이선스

MIT
