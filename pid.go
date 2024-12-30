package screenshot

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// KillByPGIDAndCleanup - kills CDP process and removes randomly created profile directory.
func (c *Config) KillByPGIDAndCleanup(cmd *exec.Cmd) {
	if c != nil && c.RandomProfileDir {
		_, err := os.Stat(c.ProfileDir)
		if err == nil {
			defer func(path string) {
				if strings.HasPrefix(path, os.TempDir()) {
					_ = os.RemoveAll(path)
				}
			}(c.ProfileDir)
		}
	}

	if cmd == nil || cmd.Process == nil {
		return
	}

	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return
	}

	err = syscall.Kill(-pgid, syscall.SIGTERM)
	if err != nil {
		return
	}

	time.Sleep(500 * time.Millisecond) // cooldown, make time for linux to actually kill process group

	_ = syscall.Kill(-pgid, syscall.SIGKILL)
}
