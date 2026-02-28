package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/lucius-han/visual_cc/internal/event"
)

const socketPath = "/tmp/visual_cc.sock"

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

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		// visual_cc not running - silently exit
		os.Exit(0)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))

	// newline-delimited JSON
	if _, err := fmt.Fprintf(conn, "%s\n", data); err != nil {
		os.Exit(0)
	}
}
