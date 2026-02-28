# visual_cc

Claude Code 실행을 실시간으로 모니터링하는 로컬 TUI 도구.

Claude Code hooks(PreToolUse / PostToolUse / Stop)가 이벤트를 Unix socket으로 전송하고,
`visual_cc`가 소켓 서버 + bubbletea TUI를 동시에 실행하며 실시간 로그와 통계를 표시합니다.

```
 visual_cc                                    q quit  ↑↓ scroll  G bottom  c clear
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  13:45:01  ● Bash                              ╭─ 통계 ────────────╮
            │ ls -la /src                       │ ⏱  4m 32s         │
  13:45:01  ✓ Bash                    142ms     │ 🔤 0 tokens        │
                                                │ 💰 $0.0000         │
  13:45:03  ● Read                              ├─ Tool 호출 ───────┤
            │ /src/main.go                      │ Bash   ████ 12    │
  13:45:03  ✓ Read                      8ms     │ Read   ███  8     │
                                                │ Edit   █    3     │
  13:45:07  ● Edit                              ╰───────────────────╯
            │ /src/main.go
  13:45:08  ✓ Edit                     34ms
  13:45:10  ✗ Bash                    error
            │ permission denied
```

## 버전

- **v0.1.0** — 기본 모니터링 (PreToolUse / PostToolUse / Stop / Notification 이벤트 표시, Tool 호출 통계)

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

### 이벤트 색상

| 이벤트 | 색상 | 아이콘 |
|--------|------|--------|
| PreToolUse | 파란색 | `●` |
| PostToolUse 성공 | 초록색 | `✓` |
| PostToolUse 실패 | 빨간색 | `✗` |
| Stop | 보라색 | `■` |
| Notification | 노란색 | `⚡` |

## 아키텍처

```
Claude Code 세션
  └─ PreToolUse / PostToolUse / Stop / Notification hooks
       └─ JSON event → visual_cc-hook (stdin 읽어 Unix socket 전송)
            └─ /tmp/visual_cc.sock
                 └─ visual_cc (Go binary)
                      ├─ Socket Server  → Event Store (ring buffer, cap=500)
                      └─ TUI (bubbletea)
                           ├─ 실시간 로그 뷰 (좌측)
                           └─ 통계 패널 (우측)
```

### 프로젝트 구조

```
visual_cc/
├── cmd/
│   ├── visual_cc/         # TUI 메인 바이너리
│   └── visual_cc-hook/    # hook 이벤트 전송 바이너리
├── internal/
│   ├── event/             # HookPayload, Event 타입, FromHookPayload
│   ├── server/            # Unix socket 서버
│   ├── store/             # ring buffer + 통계 집계
│   └── tui/               # styles, render, statspanel, model (bubbletea)
├── hooks/
│   └── install.sh         # hook 등록 안내 스크립트
└── docs/plans/            # 설계 및 구현 계획 문서
```

## 개발

```bash
# 테스트
go test ./...

# 전체 빌드
go build ./...
```

## 라이선스

MIT
