package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/lucius-han/visual_cc/internal/event"
	"github.com/lucius-han/visual_cc/internal/socket"
)

func main() {
	var payload event.HookPayload
	if err := json.NewDecoder(os.Stdin).Decode(&payload); err != nil {
		// Silently exit - hook failure must not disrupt Claude Code
		os.Exit(0)
	}

	e := event.FromHookPayload(payload, time.Now())
	data, err := json.Marshal(e)
	if err != nil {
		os.Exit(0)
	}

	// S5: use per-user socket path matching the server
	conn, err := net.Dial("unix", socket.DefaultPath())
	if err != nil {
		// visual_cc not running - silently exit
		os.Exit(0)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond)) //nolint:errcheck

	// newline-delimited JSON
	if _, err := fmt.Fprintf(conn, "%s\n", data); err != nil {
		os.Exit(0)
	}
}
