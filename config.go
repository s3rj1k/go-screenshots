package screenshot

import (
	"time"
)

// Config - options for URL screenshot function.
type Config struct {
	CMD        string
	Host       string
	ProfileDir string

	URL            string
	AcceptLanguage string
	UserAgent      string

	Flags []string

	FullPage bool

	RandomProfileDir bool

	Port int

	WindowWidth  int
	WindowHeight int

	PaddingTop    int
	PaddingBottom int
	PaddingLeft   int
	PaddingRight  int

	Wait            time.Duration
	ContextDeadline time.Duration
}

// DefaultConfig - creates structure with default values.
func DefaultConfig() Config {
	return Config{
		CMD:  "/usr/bin/google-chrome-stable",
		Host: "127.0.0.1",

		RandomProfileDir: true,

		WindowWidth:  1920,
		WindowHeight: 1080,

		URL:            "https://google.com",
		AcceptLanguage: "*",

		Wait:            5 * time.Second,
		ContextDeadline: 300 * time.Second,

		// https://peter.sh/experiments/chromium-command-line-switches/
		Flags: []string{
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
		},
	}
}
