package screenshot

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"syscall"
	"time"
)

// Screenshot - makes screenshot for URL, returns raw slice bytes.
func (c *Config) Screenshot() ([]byte, error) {
	var err error

	if c.RandomProfileDir {
		c.ProfileDir, err = os.MkdirTemp(os.TempDir(), "cdp")
		if err != nil {
			return nil, fmt.Errorf("failed to get temporary directory for URL=%q: %w", c.URL, err)
		}
	} else {
		c.ProfileDir = filepath.Join(os.TempDir(), "cdp")

		err = os.MkdirAll(c.ProfileDir, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to get temporary directory for URL=%q: %w", c.URL, err)
		}
	}

	if c.Port == 0 {
		c.Port, err = GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to get free TCP port for URL=%q: %w", c.URL, err)
		}
	}

	// https://en.wikipedia.org/wiki/8K_resolution
	if c.WindowWidth > 8192 {
		c.WindowWidth = 8192
	}
	if c.WindowHeight > 8192 {
		c.WindowHeight = 8192
	}
	if c.WindowWidth < 50 {
		c.WindowWidth = 50
	}
	if c.WindowHeight < 50 {
		c.WindowHeight = 50
	}

	flags := slices.Clone(c.Flags)

	flags = append(flags, []string{
		fmt.Sprintf("--remote-debugging-port=%s", strconv.Itoa(c.Port)),
		fmt.Sprintf("--window-size=%d,%d", c.WindowWidth, c.WindowHeight),
	}...)

	if c.AcceptLanguage != "" {
		flags = append(flags, fmt.Sprintf("--lang=%s", c.AcceptLanguage))
	}

	if c.UserAgent != "" {
		flags = append(flags, fmt.Sprintf("--user-agent=%s", c.UserAgent))
	}

	if len(c.ProfileDir) > 0 {
		flags = append(flags, fmt.Sprintf("--user-data-dir=%s", c.ProfileDir))
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.ContextDeadline)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.CMD, flags...)

	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start CDP for url=%s: %s", c.URL, err.Error())
	}

	time.Sleep(500 * time.Millisecond) // cooldown, make time for linux to actually start process
	defer c.KillByPGIDAndCleanup(cmd)

	return c.CDPScreenshot(ctx)
}
