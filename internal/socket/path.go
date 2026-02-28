package socket

import (
	"fmt"
	"os/user"
)

// DefaultPath returns a per-user Unix socket path to prevent cross-user
// collisions and reduce symlink race exposure in /tmp (S5).
func DefaultPath() string {
	u, err := user.Current()
	if err != nil {
		return "/tmp/visual_cc.sock"
	}
	return fmt.Sprintf("/tmp/visual_cc-%s.sock", u.Uid)
}
