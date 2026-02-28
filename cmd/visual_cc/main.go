package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lucius-han/visual_cc/internal/event"
	"github.com/lucius-han/visual_cc/internal/server"
	"github.com/lucius-han/visual_cc/internal/store"
	"github.com/lucius-han/visual_cc/internal/tui"
)

const (
	socketPath    = "/tmp/visual_cc.sock"
	storeCapacity = 500
)

func main() {
	s := store.New(storeCapacity)

	var program *tea.Program

	srv, err := server.New(socketPath, func(e event.Event) {
		if program != nil {
			program.Send(tui.NewEventMsg(e))
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "소켓 서버 시작 실패: %v\n", err)
		os.Exit(1)
	}
	go srv.Start()
	defer srv.Stop()

	m := tui.NewModel(s)
	program = tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI 오류: %v\n", err)
		os.Exit(1)
	}
}
