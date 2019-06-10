package screenshot

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/css"
	"github.com/mafredri/cdp/protocol/emulation"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/security"
	"github.com/mafredri/cdp/rpcc"
)

// CDPScreenshot - low-level function that creates screenshot for URL using CDP.
func (c *Config) CDPScreenshot(ctx context.Context) ([]byte, error) {
	var width, height float64
	var format string

	if c.IsJPEG {
		format = "jpeg"
	} else {
		format = "png"
	}

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

	err = cdp.Debugger.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Debugger for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.HeapProfiler.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Heap Profiler for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Inspector.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Inspector for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.LayerTree.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Layer Tree for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Log.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Log for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Overlay.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Overlay for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Performance.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Performance for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Profiler.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Profiler for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Network.Disable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to disable Network for URL='%s': %s", c.URL, err.Error())
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

	err = cdp.DOM.Enable(ctx)
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

	err = cdp.Emulation.SetDeviceMetricsOverride(ctx,
		&emulation.SetDeviceMetricsOverrideArgs{
			Width:             c.WindowWidth,
			Height:            c.WindowHight,
			DeviceScaleFactor: 1,
			Mobile:            false})
	if err != nil {
		return nil, fmt.Errorf("failed to set Device Metrics Overrides for URL='%s': %s", c.URL, err.Error())
	}

	err = cdp.Security.SetIgnoreCertificateErrors(ctx,
		&security.SetIgnoreCertificateErrorsArgs{
			Ignore: true,
		},
	)
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
		width = layout.ContentSize.Width
		height = layout.ContentSize.Height
	} else {
		width = float64(c.WindowWidth)
		height = float64(c.WindowHight)
	}

	if layout.ContentSize.Height > layout.VisualViewport.ClientHeight {

		var styleSheet *css.CreateStyleSheetReply

		styleSheet, err = cdp.CSS.CreateStyleSheet(ctx, css.NewCreateStyleSheetArgs(nav.FrameID))
		if err != nil {
			return nil, fmt.Errorf("failed to create CSS override for URL='%s': %s", c.URL, err.Error())
		}

		_, err = cdp.CSS.SetStyleSheetText(ctx, css.NewSetStyleSheetTextArgs(
			styleSheet.StyleSheetID,
			`html { height: auto !important; }`,
		))
		if err != nil {
			return nil, fmt.Errorf("failed to set CSS override for URL='%s': %s", c.URL, err.Error())
		}
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

	if c.IsJPEG {
		screenshotArgs.SetQuality(c.JpegQuality)
	}

	if c.Wait > c.Deadline {
		time.Sleep(c.Deadline)
	} else {
		time.Sleep(c.Wait)
	}

	err = cdp.Page.StopLoading(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to stop Page loading for URL='%s': %s", c.URL, err.Error())
	}

	scr, err := cdp.Page.CaptureScreenshot(ctx, screenshotArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to Capture Screenshot for URL='%s': %s", c.URL, err.Error())
	}

	var image []byte

	encURL := base64.StdEncoding.EncodeToString([]byte(c.URL))

	if c.IsJPEG {
		image, err = addCOMtoJPEG(scr.Data, []byte(encURL)) // consider adding time deadline here
		if err != nil {
			return nil, fmt.Errorf("failed to add JPEG comment section for URL='%s': %s", c.URL, err.Error())
		}
	} else {
		image = scr.Data
	}

	return image, nil
}
