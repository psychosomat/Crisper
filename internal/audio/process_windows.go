//go:build windows

package audio

import (
	"os/exec"
	"time"
)

func setSysProcAttr(cmd *exec.Cmd) {}

func killPg(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	cmd.Process.Kill()
	time.Sleep(50 * time.Millisecond)
}
