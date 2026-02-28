package server_test

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/lucius-han/visual_cc/internal/event"
	"github.com/lucius-han/visual_cc/internal/server"
)

func TestServer_ReceivesEvent(t *testing.T) {
	sockPath := "/tmp/visual_cc_test.sock"
	os.Remove(sockPath)

	received := make(chan event.Event, 1)
	srv, err := server.New(sockPath, func(e event.Event) {
		received <- e
	})
	if err != nil {
		t.Fatal(err)
	}
	go srv.Start()
	defer srv.Stop()

	time.Sleep(10 * time.Millisecond)

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	e := event.Event{
		SessionID: "test-session",
		Type:      event.TypePreToolUse,
		ToolName:  "Bash",
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(e)
	fmt.Fprintf(conn, "%s\n", data)
	conn.Close()

	select {
	case got := <-received:
		if got.ToolName != "Bash" {
			t.Errorf("expected Bash, got %s", got.ToolName)
		}
		if got.SessionID != "test-session" {
			t.Errorf("expected test-session, got %s", got.SessionID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout: no event received")
	}
}

func TestServer_Stop_CleansUpSocket(t *testing.T) {
	sockPath := "/tmp/visual_cc_stop_test.sock"
	os.Remove(sockPath)

	srv, err := server.New(sockPath, func(e event.Event) {})
	if err != nil {
		t.Fatal(err)
	}
	go srv.Start()
	time.Sleep(10 * time.Millisecond)

	// socket file should exist
	if _, err := os.Stat(sockPath); os.IsNotExist(err) {
		t.Fatal("socket file should exist after Start")
	}

	srv.Stop()
	time.Sleep(10 * time.Millisecond)

	// socket file should be cleaned up
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Error("socket file should be removed after Stop")
	}
}
