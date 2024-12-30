package screenshot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/emulation"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/security"
	"github.com/mafredri/cdp/rpcc"
)

// CDPScreenshot - low-level function that creates screenshot for URL using CDP.
func (c *Config) CDPScreenshot(ctx context.Context) ([]byte, error) {
	var (
		width, height float64
		format        string = "png"
	)

	devt := devtool.New(fmt.Sprintf("http://%s:%s", c.Host, strconv.Itoa(c.Port)))

LOOP:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)

			_, err := devt.List(ctx)
			if err == nil {
				break LOOP
			}
		}
	}

	pt, err := devt.Create(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start CDP for URL='%s': %s", c.URL, err.Error())
	}
	defer devt.Close(ctx, pt)

	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CDP for URL='%s': %s", c.URL, err.Error())
	}
	defer conn.Close()

	cdp := cdp.NewClient(conn)
	defer cdp.Browser.Close(ctx)

	// disable unused services
	services := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"Debugger", cdp.Debugger.Disable},
		{"HeapProfiler", cdp.HeapProfiler.Disable},
		{"Inspector", cdp.Inspector.Disable},
		{"LayerTree", cdp.LayerTree.Disable},
		{"Log", cdp.Log.Disable},
		{"Overlay", cdp.Overlay.Disable},
		{"Performance", cdp.Performance.Disable},
		{"Profiler", cdp.Profiler.Disable},
	}

	for _, service := range services {
		if err := service.fn(ctx); err != nil {
			return nil, fmt.Errorf("failed to disable %s for URL='%s': %s", service.name, c.URL, err.Error())
		}
	}

	err = cdp.Network.Enable(ctx, &network.EnableArgs{})
	if err != nil {
		return nil, fmt.Errorf("failed to enable network events for URL='%s': %s", c.URL, err.Error())
	}

	_ = page.NewSetAdBlockingEnabledArgs(true)

	domContent, err := cdp.Page.DOMContentEventFired(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to catch DOMContentEventFired for URL='%s': %s", c.URL, err.Error())
	}
	defer domContent.Close()

	err = cdp.Page.Enable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to enable Page events for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.DOM.Enable(ctx, &dom.EnableArgs{})
	if err != nil {
		return nil, fmt.Errorf("failed to enable DOM events for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.CSS.Enable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to enable CSS events for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Emulation.ClearDeviceMetricsOverride(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to clear Device Metrics Override for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Emulation.SetDeviceMetricsOverride(ctx, &emulation.SetDeviceMetricsOverrideArgs{
		Width:             c.WindowWidth,
		Height:            c.WindowHeight,
		DeviceScaleFactor: 1,
		Mobile:            false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set Device Metrics Overrides for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Security.SetIgnoreCertificateErrors(ctx, &security.SetIgnoreCertificateErrorsArgs{
		Ignore: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set Ignore Certificate errors option for URL='%s': %s", c.URL, err.Error())
	}

	loadEventFired, err := cdp.Page.LoadEventFired(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to catch page Load Event fired for URL='%s': %s", c.URL, err.Error())
	}
	defer loadEventFired.Close()

	nav, err := cdp.Page.Navigate(ctx, page.NewNavigateArgs(c.URL))
	if err != nil {
		return nil, fmt.Errorf("failed to Navigate to URL='%s': %s", c.URL, err.Error())
	}

	_, err = domContent.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive DOM content for URL='%s': %s", c.URL, err.Error())
	}

	_, err = loadEventFired.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive Load Event fired for URL='%s': %s", c.URL, err.Error())
	}

	if nav.ErrorText != nil {
		return nil, fmt.Errorf("failed to Navigate to URL='%s': %s", c.URL, errors.New(*nav.ErrorText))
	}

	layout, err := cdp.Page.GetLayoutMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Layout Metrics for URL='%s': %s", c.URL, err.Error())
	}

	if c.FullPage {
		width = layout.CSSContentSize.Width
		height = layout.CSSContentSize.Height
	} else {
		width = float64(c.WindowWidth)
		height = float64(c.WindowHeight)
	}

	if layout.CSSContentSize.Height > layout.CSSVisualViewport.ClientHeight {
		err = cdp.Emulation.SetDeviceMetricsOverride(ctx, &emulation.SetDeviceMetricsOverrideArgs{
			Width:             int(layout.CSSContentSize.Width),
			Height:            int(layout.CSSContentSize.Height),
			DeviceScaleFactor: 1,
			Mobile:            false,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set full page device metrics for URL='%s': %s", c.URL, err.Error())
		}

		_, err = cdp.DOM.GetDocument(ctx, &dom.GetDocumentArgs{})
		if err != nil {
			return nil, fmt.Errorf("failed to force layout recalculation for URL='%s': %s", c.URL, err.Error())
		}
	}

	done := make(chan bool)
	var lastError error

	go func() {
		defer close(done)

		loadingFinished, err := cdp.Network.LoadingFinished(ctx)
		if err != nil {
			lastError = fmt.Errorf("failed to create loading finished listener: %v", err)
			return
		}
		defer loadingFinished.Close()

		if _, err := loadingFinished.Recv(); err != nil {
			lastError = fmt.Errorf("failed waiting for network idle: %v", err)
			return
		}
	}()

	var timeoutChan <-chan time.Time
	if c.Wait > 0 {
		timeoutChan = time.After(c.Wait)
	}

	select {
	case <-done:
		if lastError != nil {
			return nil, lastError
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timeoutChan:
		// only reached if c.Wait > 0
	}

	screenshotArgs := page.NewCaptureScreenshotArgs().
		SetFormat(format).
		SetClip(
			page.Viewport{
				X:      0 + float64(c.PaddingLeft),
				Y:      0 + float64(c.PaddingTop),
				Width:  width + float64(c.PaddingRight),
				Height: height + float64(c.PaddingBottom),
				Scale:  1,
			},
		)

	err = cdp.Page.StopLoading(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to stop Page loading for URL='%s': %s", c.URL, err.Error())
	}

	scr, err := cdp.Page.CaptureScreenshot(ctx, screenshotArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to Capture Screenshot for URL='%s': %s", c.URL, err.Error())
	}

	return scr.Data, nil
}
