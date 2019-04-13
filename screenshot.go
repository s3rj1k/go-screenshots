package screenshot

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// Screenshot - makes screenshot for URL, returns base64 encoded slice bytes and raw slice bytes.
func (c *Config) Screenshot() ([]byte, error) {

	var err error

	if c.RandomProfileDir {
		c.ProfileDir, err = ioutil.TempDir(os.TempDir(), "cdp")
		if err != nil {
			return nil, fmt.Errorf("failed to get temp dir for URL='%s': %s", c.URL, err.Error())
		}
	} else {
		c.ProfileDir = filepath.Clean(fmt.Sprintf("%s/cdp", os.TempDir()))

		err = os.MkdirAll(c.ProfileDir, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to get temp dir for URL='%s': %s", c.URL, err.Error())
		}
	}

	if c.Port == 0 {
		c.Port, err = GetFreePort()
		if err != nil {
			return nil, fmt.Errorf("failed to get free TCP port for URL='%s': %s", c.URL, err.Error())
		}
	}

	// https://en.wikipedia.org/wiki/8K_resolution
	if c.WindowWidth > 8192 {
		c.WindowWidth = 8192
	}
	if c.WindowHight > 8192 {
		c.WindowHight = 8192
	}
	if c.WindowWidth < 50 {
		c.WindowWidth = 50
	}
	if c.WindowHight < 50 {
		c.WindowHight = 50
	}

	// https://peter.sh/experiments/chromium-command-line-switches/
	var flags = []string{
		"--autoplay-policy=document-user-activation-required",
		"--disable-client-side-phishing-detection",
		"--disable-cloud-import",
		"--disable-default-apps",
		"--disable-dinosaur-easter-egg",
		"--disable-gpu",
		"--disable-logging",
		"--disable-new-tab-first-run",
		"--disable-offer-upload-credit-cards",
		"--disable-signin-promo",
		"--disable-sync",
		"--disable-translate",
		"--headless",
		"--hide-scrollbars",
		"--ignore-certificate-errors",
		"--mute-audio",
		"--no-default-browser-check",
		"--no-first-run",
		"--no-pings",
		"--no-referrers",
		"--no-sandbox",
		"--password-store=basic",
		fmt.Sprintf("--user-agent=%s", c.UserAgent),
		fmt.Sprintf("--lang=%s", c.AcceptLanguage),
		fmt.Sprintf("--remote-debugging-port=%s", strconv.Itoa(c.Port)),
		fmt.Sprintf("--window-size=%d,%d", c.WindowWidth, c.WindowHight),
	}

	if len(c.ProfileDir) > 0 {
		flags = append(flags, fmt.Sprintf("--user-data-dir=%s", c.ProfileDir))
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Deadline)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.CMD, flags...)

	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start CDP for url=%s: %s", c.URL, err.Error())
	}

	time.Sleep(500 * time.Millisecond) // cooldown, make time for linux to actually start process

	go func(cmd *exec.Cmd) {
		_ = cmd.Wait()
	}(cmd)

	image, err := c.CDPScreenshot(ctx)
	defer c.KillByPgidAndCleanup(cmd)

	return image, err
}
