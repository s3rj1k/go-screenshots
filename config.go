package screenshot

import (
	"time"
)

// Config - options for URL screenshot function.
type Config struct {
	CMD  string
	Host string

	URL            string
	AcceptLanguage string
	UserAgent      string
	ProfileDir     string

	FullPage bool
	IsJPEG   bool

	RandomProfileDir bool

	Port int

	WindowWidth int
	WindowHight int

	PaddingTop    int
	PaddingBottom int
	PaddingLeft   int
	PaddingRight  int

	JpegQuality int

	Wait     time.Duration
	Deadline time.Duration
}

// DefaultConfig - creates structure with default values.
func DefaultConfig() Config {
	return Config{
		CMD:  "/usr/bin/google-chrome-stable",
		Host: "127.0.0.1",

		Port: 0,

		WindowWidth: 1920,
		WindowHight: 1080,

		PaddingTop:    0,
		PaddingBottom: 0,
		PaddingLeft:   0,
		PaddingRight:  0,

		JpegQuality: 95,

		URL:            "https://google.com",
		AcceptLanguage: "*",
		UserAgent:      "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.119 Safari/537.36",

		Wait:     5 * time.Second,
		Deadline: 300 * time.Second,

		RandomProfileDir: true,

		FullPage: true,
		IsJPEG:   false,
	}
}
