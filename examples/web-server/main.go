package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	screenshot "github.com/s3rj1k/go-webpage-screenshots"
)

/* # User-Agent MUST be URL Encoded
curl -s -X POST http://127.0.0.1:8888/screenshot \
  -d 'url=https://www.google.com/' \
  -d 'time-wait=5' \
  -d 'viewport-width=1920' \
  -d 'viewport-height=1080' \
  -d 'fullpage=true' \
  -d 'user-agent=Mozilla/5.0' \
  -d 'accept-language=uk-UA,uk;'
*/

const (
	defaultContextDeadline = 15 * 60
	defaultTimeWait        = 15
	maxTimeWait            = defaultContextDeadline
	defaultViewportWidth   = 1920
	maxViewportWidth       = 4096
	defaultViewportHeight  = 1080
	maxViewportHeight      = 2160
	defaultIsFullpage      = true
)

var (
	cmdBin      string
	cmdDeadLine int
	cmdPort     int
)

type screenshotForm struct {
	remoteURL      string
	userAgent      string
	acceptLanguage string
	timeWait       int
	viewportWidth  int
	viewportHeight int
	isFullpage     bool
}

func parseScreenshotForm(r *http.Request) (screenshotForm, error) {
	if err := r.ParseForm(); err != nil {
		return screenshotForm{}, fmt.Errorf("failed to parse form: %w", err)
	}

	var form screenshotForm

	formURL := strings.TrimSpace(r.PostFormValue("url"))
	formTimeWait := strings.TrimSpace(r.PostFormValue("time-wait"))
	formViewportWidth := strings.TrimSpace(r.PostFormValue("viewport-width"))
	formViewportHeight := strings.TrimSpace(r.PostFormValue("viewport-height"))
	formFullpage := strings.TrimSpace(strings.ToLower(r.PostFormValue("fullpage")))
	formUserAgent := strings.TrimSpace(r.PostFormValue("user-agent"))
	formAcceptLanguage := strings.TrimSpace(r.PostFormValue("accept-language"))

	if len(formURL) == 0 {
		return screenshotForm{}, fmt.Errorf("empty remote URL")
	}

	_, err := url.ParseRequestURI(formURL)
	if err != nil {
		return screenshotForm{}, fmt.Errorf("Invalid remote URL: %s", formURL)
	}

	form.remoteURL = formURL

	form.timeWait = defaultTimeWait
	if len(formTimeWait) != 0 {
		if i, err := strconv.Atoi(formTimeWait); err == nil {
			if i >= 0 && i < maxTimeWait {
				form.timeWait = i
			} else {
				form.timeWait = maxTimeWait
			}
		}
	}

	form.viewportWidth = defaultViewportWidth
	if len(formViewportWidth) != 0 {
		if i, err := strconv.Atoi(formViewportWidth); err == nil {
			if i > 0 && i <= maxViewportWidth {
				form.viewportWidth = i
			}
		}
	}

	form.viewportHeight = defaultViewportHeight
	if len(formViewportHeight) != 0 {
		if i, err := strconv.Atoi(formViewportHeight); err == nil {
			if i > 0 && i <= maxViewportHeight {
				form.viewportHeight = i
			}
		}
	}

	switch formFullpage {
	case "true":
		form.isFullpage = true
	case "false":
		form.isFullpage = false
	default:
		form.isFullpage = defaultIsFullpage
	}

	if len(formUserAgent) != 0 {
		form.userAgent = formUserAgent
	}

	if len(formAcceptLanguage) != 0 {
		form.acceptLanguage = formAcceptLanguage
	}

	return form, nil
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		handler.ServeHTTP(rw, r)

		path := r.URL.Path
		query := r.URL.RawQuery

		log := []string{
			fmt.Sprintf("%v", start.Format("2006/01/02 - 15:04:05")),
			fmt.Sprintf("status: %d", rw.statusCode),
			fmt.Sprintf("remote_ip: %s", r.RemoteAddr),
			fmt.Sprintf("user_agent: %s", r.UserAgent()),
			fmt.Sprintf("method: %s", r.Method),
		}

		if len(query) != 0 {
			log = append(log, fmt.Sprintf("path: %s?%s", path, query))
		} else {
			log = append(log, fmt.Sprintf("path: %s", path))
		}

		if path == "/screenshot" && r.Method == http.MethodPost {
			if form, err := parseScreenshotForm(r); err == nil {
				log = append(log, fmt.Sprintf("remote_url: %s", form.remoteURL))
				log = append(log, fmt.Sprintf("wait_time: %d sec", form.timeWait))
				log = append(log, fmt.Sprintf("fullscreen: %t", form.isFullpage))
				log = append(log, fmt.Sprintf("size: %dx%d", form.viewportWidth, form.viewportHeight))
				log = append(log, fmt.Sprintf("emulated_user_agent: %s", form.userAgent))
				log = append(log, fmt.Sprintf("accept_language: %s", form.acceptLanguage))
			}
		}

		fmt.Printf("%s\n", strings.Join(log, " | "))
	})
}

func handleScreenshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	form, err := parseScreenshotForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cdp := screenshot.DefaultConfig()
	cdp.CMD = cmdBin
	cdp.URL = form.remoteURL
	cdp.Wait = time.Duration(form.timeWait) * time.Second
	cdp.ContextDeadline = time.Duration(cmdDeadLine) * time.Second
	cdp.WindowWidth = form.viewportWidth
	cdp.WindowHeight = form.viewportHeight
	cdp.FullPage = form.isFullpage
	cdp.UserAgent = form.userAgent
	cdp.AcceptLanguage = form.acceptLanguage

	image, err := cdp.Screenshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, base64.StdEncoding.EncodeToString(image))
}

func handleDefault(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "form.html")
}

func main() {
	flag.StringVar(&cmdBin, "cdp-bin", "/usr/bin/google-chrome-stable", "path to Chrome/Chromium binary")
	flag.IntVar(&cmdDeadLine, "time-deadline", 300, "deadline in seconds")
	flag.IntVar(&cmdPort, "listen-port", 8888, "tcp port for web server")
	flag.Parse()

	mux := http.NewServeMux()

	mux.HandleFunc("/screenshot", handleScreenshot)
	mux.HandleFunc("/", handleDefault)

	handler := logRequest(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cmdPort),
		Handler: handler,
	}

	log.Printf("Starting webpage-screenshot server on port %d", cmdPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
