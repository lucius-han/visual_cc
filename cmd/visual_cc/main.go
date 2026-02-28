package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucius-han/visual_cc/internal/event"
	"github.com/lucius-han/visual_cc/internal/server"
	"github.com/lucius-han/visual_cc/internal/socket"
	"github.com/lucius-han/visual_cc/internal/store"
	"github.com/lucius-han/visual_cc/internal/tui"
)

const storeCapacity = 500

func main() {
	s := store.New(storeCapacity)

	// S10: initialise program before starting the server goroutine so the
	// closure captures a fully-initialised pointer — no mutex needed.
	m := tui.NewModel(s)
	program := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// S5: use per-user socket path to prevent cross-user collision and symlink race
	sockPath := socket.DefaultPath()
	srv, err := server.New(sockPath, func(e event.Event) {
		program.Send(tui.NewEventMsg(e))
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "소켓 서버 시작 실패: %v\n", err)
		os.Exit(1)
	}
	go srv.Start()
	defer srv.Stop()

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI 오류: %v\n", err)
		os.Exit(1)
	}
}
