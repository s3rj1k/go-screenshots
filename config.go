package screenshot

import (
	"time"
)

// Config - options for URL screenshot function.
type Config struct {
	CMD  string
	Host string

	Port int

	WindowWidth   int
	WindowHight   int
	BottomPadding int

	JpegQuality int

	URL            string
	AcceptLanguage string
	UserAgent      string
	ProfileDir     string

	Wait     time.Duration
	Deadline time.Duration

	RandomProfileDir bool

	FullPage bool
	IsJPEG   bool
}

// DefaultConfig - creates structure with default values.
func DefaultConfig() Config {
	return Config{
		CMD:  "/usr/bin/google-chrome-stable",
		Host: "127.0.0.1",

		Port: 0,

		WindowWidth:   1920,
		WindowHight:   1080,
		BottomPadding: 0,

		JpegQuality: 95,

		URL:            "https://google.com",
		AcceptLanguage: "*",
		UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36",

		Wait:     5 * time.Second,
		Deadline: 300 * time.Second,

		RandomProfileDir: true,

		FullPage: true,
		IsJPEG:   false,
	}
}
