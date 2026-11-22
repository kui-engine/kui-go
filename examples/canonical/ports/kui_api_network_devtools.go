//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_network_devtools.go; edit that file, not this copy.
// API tour: HTTP client and in-process DevTools (examples/api_tours/network_devtools). The host
// fetches from the window's own DevTools server (a loopback HTTP server: /json/version) with the
// sync, worker-thread and main-thread request methods, provokes transport errors (closed port,
// malformed URL, cancel), runs raw CDP commands, injects mouse / key / IME input and reads box
// models. No internet access is needed. One #tour-log row per check; #tour-result ends as
// "PASS n/n" once every asynchronous answer is in. The panel buttons repeat the calls.
package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------
// check:    a documented behaviour; a failure makes the tour FAIL.
// knownBug: a documented behaviour the engine currently gets wrong (reported bug).

type tour struct {
	doc         *kui.Document
	total       int
	failed      int
	known       int
	failedNames []string
	pending     int  // asynchronous checks still waiting for their callback
	readyDone   bool // the synchronous part has run
	finished    bool
}

func (t *tour) row(name, result, cell string) {
	log, err := t.doc.GetElementByID("tour-log")
	if err != nil || log == nil {
		return
	}
	defer log.Close()
	row, err := t.doc.CreateElement("tr")
	if err != nil || row == nil {
		return
	}
	defer row.Close()
	_ = row.SetAttribute("data-check", name)
	_ = row.SetAttribute("data-result", result)
	if result != "ok" {
		_ = row.AddClass("bad")
	}
	for _, text := range []string{name, cell} {
		if td, err := t.doc.CreateElement("td"); err == nil && td != nil {
			_ = td.SetTextContent(text)
			_ = row.AppendChild(td)
			td.Close()
		}
	}
	_ = log.AppendChild(row)
}

func (t *tour) check(name string, ok bool, detail string) {
	t.total++
	if ok {
		t.row(name, "ok", "ok")
		return
	}
	t.failed++
	t.failedNames = append(t.failedNames, name)
	fmt.Printf("tour check failed: %s: %s\n", name, detail)
	t.row(name, "fail", "FAIL: "+detail)
}

func (t *tour) knownBug(name string, ok bool, bug, detail string) {
	if ok {
		t.total++
		t.row(name, "ok", "ok")
		return
	}
	t.known++
	fmt.Printf("tour known bug: %s: %s\n", name, detail)
	t.row(name, "known", fmt.Sprintf("KNOWN BUG: %s (%s)", bug, detail))
}

func (t *tour) finish() {
	if t.finished {
		return
	}
	t.finished = true
	result, err := t.doc.GetElementByID("tour-result")
	if err != nil || result == nil {
		return
	}
	defer result.Close()
	text := fmt.Sprintf("PASS %d/%d", t.total, t.total)
	if t.failed > 0 {
		text = fmt.Sprintf("FAIL %d/%d: %s", t.failed, t.total, strings.Join(t.failedNames, ", "))
	}
	if t.known > 0 {
		text += fmt.Sprintf(" (known bugs: %d)", t.known)
	}
	_ = result.SetTextContent(text)
	_ = result.SetAttribute("data-total", fmt.Sprint(t.total))
	_ = result.SetAttribute("data-failed", fmt.Sprint(t.failed))
	_ = result.SetAttribute("data-known", fmt.Sprint(t.known))
	if t.failed == 0 {
		_ = result.AddClass("pass")
	} else {
		_ = result.AddClass("fail")
	}
}

// settle marks one asynchronous check as answered; the verdict follows the last one.
func (t *tour) settle() {
	t.pending--
	if t.pending == 0 && t.readyDone {
		t.finish()
	}
}

// ---- Host state ---------------------------------------------------------------------------

const closedURL = "http://127.0.0.1:1/" // nothing listens on port 1

type host struct {
	doc          *kui.Document
	client       *kui.HTTPClient
	cancelClient *kui.HTTPClient
	versionURL   string
	port         uint16
	clicks       int
	wheels       int
	moveX, moveY float32
	firstKey     string
	firstCode    string
	shiftSeen    bool
}

func (h *host) setText(id, text string) {
	if element, err := h.doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

func isJSONVersion(r kui.HTTPResponse, err error) bool {
	return err == nil && r.Status == 200 && strings.Contains(r.ContentType, "application/json") &&
		strings.HasPrefix(string(r.Body), "{") && strings.Contains(string(r.Body), `"Browser":"KUI/`) && r.Error == ""
}

func (h *host) showReply(prefix string, r kui.HTTPResponse, err error) {
	contentType := r.ContentType
	if contentType == "" {
		contentType = "-"
	}
	text := fmt.Sprintf("%s status=%d type=%s bytes=%d", prefix, r.Status, contentType, len(r.Body))
	if r.Error != "" {
		text += " error=" + r.Error
	} else if err != nil {
		text += " error=" + err.Error()
	}
	h.setText("http-out", text)
	body := string(r.Body)
	if body == "" {
		body = "-"
	}
	h.setText("http-body", body)
}

// ---- HTTP against the loopback DevTools server ---------------------------------------------

// Not started by the runner (KUI_EXAMPLE_DEVTOOLS_PORT unset): start one, port 0 picks a free one.
func (h *host) devtoolsPort(window *kui.Window) {
	port, err := window.DevToolsPort()
	if err != nil || port == 0 {
		_, _ = window.StartDevTools(0)
		port, _ = window.DevToolsPort()
	}
	h.port = port
	h.versionURL = fmt.Sprintf("http://127.0.0.1:%d/json/version", port)
}

func (h *host) checkHTTP(t *tour) {
	var err error
	h.client, err = kui.NewHTTPClient()
	t.check("http.client_create", err == nil && h.client != nil, fmt.Sprint(err))
	err = h.client.SetUserAgent("kui-api-tour/1.0")
	t.check("http.set_user_agent", err == nil, fmt.Sprint(err))
	h.cancelClient, _ = kui.NewHTTPClient()

	reply, err := h.client.Fetch(h.versionURL, 3)
	t.check("http.fetch_sync", isJSONVersion(reply, err),
		fmt.Sprintf("http %d type %q error %q err %v", reply.Status, reply.ContentType, reply.Error, err))
	t.check("http.response_get_content_type", reply.ContentType == "application/json",
		fmt.Sprintf("content type %q", reply.ContentType))

	reply, err = h.client.Fetch("not a url", 3)
	t.check("http.response_get_error-bad-url", err == nil && reply.Status == 0 && reply.Error != "",
		fmt.Sprintf("http %d error %q err %v", reply.Status, reply.Error, err))
	t.check("http.response_get_body-empty-on-error", len(reply.Body) == 0, fmt.Sprintf("body of %d bytes", len(reply.Body)))

	// Main-thread callbacks may touch the DOM directly.
	if err := h.client.FetchAsyncOnMainThread(h.versionURL, 3, func(r kui.HTTPResponse, err error) {
		t.check("http.fetch_async_on_main_thread", t.readyDone && isJSONVersion(r, err),
			fmt.Sprintf("after ready %v, http %d type %q error %q err %v", t.readyDone, r.Status, r.ContentType, r.Error, err))
		h.showReply("GET /json/version (main thread):", r, err)
		t.settle()
	}); err == nil {
		t.pending++
	} else {
		t.check("http.fetch_async_on_main_thread", false, err.Error())
	}
	if err := h.client.FetchAsyncOnMainThread(closedURL, 3, func(r kui.HTTPResponse, err error) {
		t.check("http.response_get_error-closed-port", err == nil && r.Status == 0 && r.Error != "",
			fmt.Sprintf("http %d error %q err %v", r.Status, r.Error, err))
		t.settle()
	}); err == nil {
		t.pending++
	} else {
		t.check("http.response_get_error-closed-port", false, err.Error())
	}
	// Worker-thread callbacks hand the result to the UI thread with PostTask.
	if err := h.client.FetchAsync(h.versionURL, 3, func(r kui.HTTPResponse, err error) {
		_ = kui.PostTask(func() {
			t.check("http.fetch_async", isJSONVersion(r, err),
				fmt.Sprintf("http %d type %q error %q err %v", r.Status, r.ContentType, r.Error, err))
			t.settle()
		})
	}); err == nil {
		t.pending++
	} else {
		t.check("http.fetch_async", false, err.Error())
	}
	// A request that is cancelled right away still answers once, with the error set.
	if err := h.cancelClient.FetchAsync(closedURL, 3, func(r kui.HTTPResponse, err error) {
		_ = kui.PostTask(func() {
			t.check("http.cancel_all", err == nil && r.Status == 0 && r.Error != "",
				fmt.Sprintf("http %d error %q err %v", r.Status, r.Error, err))
			t.settle()
		})
	}); err == nil {
		t.pending++
		if err := h.cancelClient.CancelAll(); err != nil {
			t.check("http.cancel_all-status", false, err.Error())
		}
	} else {
		t.check("http.cancel_all", false, err.Error())
	}
}

// ---- DevTools: port, raw CDP commands --------------------------------------------------------

// ---- DevTools token check: off for tools that cannot pass the token, then on again --------------

func checkTokenAuth(t *tour, window *kui.Window) {
	offErr := window.SetDevToolsTokenAuth(false)
	onErr := window.SetDevToolsTokenAuth(true)
	t.check("dt.set_devtools_token_auth", offErr == nil && onErr == nil, fmt.Sprintf("off err %v, on err %v", offErr, onErr))
}

func (h *host) checkDevToolsCommands(t *tour) {
	envPort := os.Getenv("KUI_EXAMPLE_DEVTOOLS_PORT")
	t.check("dt.window_get_devtools_port", h.port != 0 && (envPort == "" || envPort == strconv.Itoa(int(h.port))),
		fmt.Sprintf("port %d (KUI_EXAMPLE_DEVTOOLS_PORT=%s)", h.port, envPort))

	response, err := h.doc.DispatchDevToolsCommand(`{"id":1,"method":"DOM.setAttribute","params":{"selector":"#cdp-target","name":"data-cdp","value":"via-command"}}`)
	attr := ""
	if target, e := h.doc.GetElementByID("cdp-target"); e == nil && target != nil {
		attr, _ = target.Attribute("data-cdp")
		target.Close()
	}
	t.check("dt.dispatch_command", err == nil && strings.Contains(response, `"id":1`) &&
		!strings.Contains(response, `"error"`) && attr == "via-command",
		fmt.Sprintf("err %v attr %q response %.120s", err, attr, response))

	response, err = h.doc.DispatchDevToolsCommand(`{"id":2,"method":"Nope.nothing"}`)
	t.check("dt.dispatch_command-unknown-method", err == nil && strings.Contains(response, `"error"`),
		fmt.Sprintf("err %v response %.120s", err, response))

	// The structured form: DevToolsCall builds the JSON and decodes the result.
	var metrics map[string]any
	err = h.doc.DevToolsCall("Page.getLayoutMetrics", nil, &metrics)
	t.check("dt.devtools_call", err == nil && len(metrics) > 0, fmt.Sprintf("err %v result %v", err, metrics))
}

// ---- Input injection: listeners record what arrives ----------------------------------------

func (h *host) center(id string) (float32, float32) {
	var rect kui.Rect
	if element, err := h.doc.GetElementByID(id); err == nil && element != nil {
		rect, _ = element.BoundingRect()
		element.Close()
	}
	return rect.X + rect.Width/2, rect.Y + rect.Height/2
}

func (h *host) focus(id string) {
	if element, err := h.doc.GetElementByID(id); err == nil && element != nil {
		_ = element.Focus()
		element.Close()
	}
}

func (h *host) inputValue(id string) string {
	element, err := h.doc.GetElementByID(id)
	if err != nil || element == nil {
		return ""
	}
	defer element.Close()
	value, _ := element.Value()
	return value
}

// DispatchMouse is the raw form of DispatchMouseEvent (positional arguments).
func (h *host) injectClick(id string) {
	x, y := h.center(id)
	_ = h.doc.DispatchMouse(kui.MousePressed, x, y, kui.MouseButtonLeft, 0, 1, 0, 0)
	_ = h.doc.DispatchMouseEvent(kui.MouseInjection{Kind: kui.MouseReleased, X: x, Y: y, Button: kui.MouseButtonLeft, ClickCount: 1})
}

func (h *host) injectWheel() {
	x, y := h.center("wheel-box")
	_ = h.doc.DispatchMouseEvent(kui.MouseInjection{Kind: kui.MouseWheel, X: x, Y: y, DeltaY: 120})
}

// Key down / char / key up for each letter, as a keyboard would.
func (h *host) injectText(letters string) {
	for _, letter := range letters {
		key, code := string(letter), "Key"+strings.ToUpper(string(letter))
		_ = h.doc.DispatchKey(kui.KeyDown, key, code, 0, "", false)
		_ = h.doc.DispatchKeyEvent(kui.KeyInjection{Kind: kui.KeyChar, Key: key, Code: code, Text: key})
		_ = h.doc.DispatchKeyEvent(kui.KeyInjection{Kind: kui.KeyUp, Key: key, Code: code})
	}
}

// Start, an intermediate composition, then the committed text ("ni" -> U+4F60).
func (h *host) injectIME() {
	_ = h.doc.DispatchIME(kui.IMEStart, "")
	_ = h.doc.DispatchIME(kui.IMEUpdate, "ni")
	_ = h.doc.DispatchIME(kui.IMECommit, "你")
}

func (h *host) listen(ctx *runner.Context) error {
	listeners := []struct {
		id, event string
		fn        func(*kui.Event)
	}{
		{"inj-target", "click", func(*kui.Event) {
			h.clicks++
			h.setText("inj-target", fmt.Sprintf("clicks=%d", h.clicks))
		}},
		{"wheel-box", "wheel", func(*kui.Event) {
			h.wheels++
			h.setText("wheel-box", fmt.Sprintf("wheel=%d", h.wheels))
		}},
		{"wheel-box", "mousemove", func(e *kui.Event) {
			if point, err := e.MouseClient(); err == nil {
				h.moveX, h.moveY = point.X, point.Y
			}
		}},
		{"inj-input", "keydown", func(e *kui.Event) {
			if h.firstKey == "" {
				h.firstKey, _ = e.Key()
				h.firstCode, _ = e.Code()
			}
			if mods, err := e.ModifierKeys(); err == nil && mods.Shift {
				h.shiftSeen = true
			}
		}},
	}
	for _, l := range listeners {
		element, err := h.doc.GetElementByID(l.id)
		if err != nil || element == nil {
			return fmt.Errorf("no #%s: %v", l.id, err)
		}
		ctx.Keep(element)
		if _, err := element.AddEventListener(l.event, l.fn); err != nil {
			return err
		}
	}
	return nil
}

func near(a, b float32) bool { return math.Abs(float64(a-b)) < 0.01 }

func (h *host) checkInjection(t *tour) {
	h.injectClick("inj-target")
	t.check("dt.dispatch_mouse_event-click", h.clicks == 1, fmt.Sprintf("clicks %d", h.clicks))

	x, y := h.center("wheel-box")
	err := h.doc.DispatchMouseEvent(kui.MouseInjection{Kind: kui.MouseMoved, X: x + 3, Y: y + 2})
	t.check("dt.dispatch_mouse_event-move", err == nil && near(h.moveX, x+3) && near(h.moveY, y+2),
		fmt.Sprintf("err %v moved to (%.1f, %.1f), listener saw (%.1f, %.1f)", err, x+3, y+2, h.moveX, h.moveY))
	h.injectWheel()
	t.check("dt.dispatch_mouse_event-wheel", h.wheels == 1, fmt.Sprintf("wheel events %d", h.wheels))
	err = h.doc.DispatchMouse(kui.MouseEventType(9), 0, 0, kui.MouseButtonNone, 0, 0, 0, 0)
	t.check("dt.dispatch_mouse_event-invalid-type", err != nil, "an unknown event type was accepted")

	h.focus("inj-input")
	h.injectText("kui")
	value := h.inputValue("inj-input")
	t.check("dt.dispatch_key_event", value == "kui" && h.firstKey == "k" && h.firstCode == "KeyK",
		fmt.Sprintf("value %q, first keydown key %q code %q", value, h.firstKey, h.firstCode))
	_ = h.doc.DispatchKeyEvent(kui.KeyInjection{Kind: kui.KeyDown, Key: "Shift", Code: "ShiftLeft", Modifiers: kui.InjectShift})
	_ = h.doc.DispatchKeyEvent(kui.KeyInjection{Kind: kui.KeyUp, Key: "Shift", Code: "ShiftLeft", Modifiers: kui.InjectShift})
	t.check("dt.dispatch_key_event-modifiers", h.shiftSeen, "keydown listener saw no Shift")

	h.focus("ime-input")
	h.injectIME()
	value = h.inputValue("ime-input")
	t.check("dt.dispatch_ime_event", value == "你", fmt.Sprintf("value %q", value))
}

// ---- Box model: four quads vs the bounding rect -----------------------------------------------

func quadSize(quad [4]kui.Point) (float32, float32) {
	return quad[2].X - quad[0].X, quad[2].Y - quad[0].Y
}

func closeTo(a, b float32) bool { return math.Abs(float64(a-b)) < 0.5 }

func (h *host) checkBoxModel(t *tour) {
	// BoxModel is the raw form (32 floats); ElementBoxModel names the four quads.
	quads, size, err := h.doc.BoxModel("#box-target")
	info, infoErr := h.doc.ElementBoxModel("#box-target")
	pw, _ := quadSize(info.Padding)
	bw, bh := quadSize(info.Border)
	mw, _ := quadSize(info.Margin)
	t.check("dt.get_box_model", err == nil && infoErr == nil && closeTo(size.Width, 120) && closeTo(size.Height, 40) &&
		closeTo(pw, 140) && closeTo(bw, 148) && closeTo(bh, 68) && closeTo(mw, 160) && closeTo(quads[0]-quads[16], 14),
		fmt.Sprintf("err %v content %.1fx%.1f padding %.1f border %.1fx%.1f margin %.1f", err, size.Width, size.Height, pw, bw, bh, mw))
	var rect kui.Rect
	if element, e := h.doc.GetElementByID("box-target"); e == nil && element != nil {
		rect, _ = element.BoundingRect()
		element.Close()
	}
	t.check("dt.get_box_model-vs-bounding-rect", closeTo(info.Border[0].X, rect.X) && closeTo(info.Border[0].Y, rect.Y) &&
		closeTo(bw, rect.Width) && closeTo(bh, rect.Height),
		fmt.Sprintf("border quad (%.1f, %.1f) vs rect %+v", info.Border[0].X, info.Border[0].Y, rect))
	_, _, err = h.doc.BoxModel("#no-such-box")
	var kerr *kui.Error
	t.check("dt.get_box_model-missing", asKUI(err, &kerr) && kerr.Status == kui.ErrNotFound, fmt.Sprint(err))
}

func asKUI(err error, target **kui.Error) bool {
	e, ok := err.(*kui.Error)
	if ok {
		*target = e
	}
	return ok
}

// ---- Panel buttons ----------------------------------------------------------------------------

func (h *host) buttons(ctx *runner.Context) error {
	cdp := func(command string) {
		response, err := h.doc.DispatchDevToolsCommand(command)
		if len(response) > 300 {
			response = response[:300]
		}
		h.setText("cdp-out", fmt.Sprintf("err=%v %s", err, response))
	}
	actions := map[string]func(){
		"btn-http-main": func() {
			_ = h.client.FetchAsyncOnMainThread(h.versionURL, 3, func(r kui.HTTPResponse, err error) {
				h.showReply("GET (main thread):", r, err)
			})
		},
		"btn-http-sync": func() {
			r, err := h.client.Fetch(h.versionURL, 3)
			h.showReply("GET (sync):", r, err)
		},
		"btn-http-closed": func() {
			_ = h.client.FetchAsyncOnMainThread(closedURL, 3, func(r kui.HTTPResponse, err error) {
				h.showReply("closed port:", r, err)
			})
		},
		"btn-http-bad": func() {
			r, err := h.client.Fetch("not a url", 3)
			h.showReply("bad URL:", r, err)
		},
		"btn-http-cancel": func() {
			if h.cancelClient.FetchAsync(closedURL, 3, func(r kui.HTTPResponse, err error) {
				_ = kui.PostTask(func() { h.showReply("worker + cancel_all:", r, err) })
			}) == nil {
				_ = h.cancelClient.CancelAll()
			}
		},
		"btn-cdp-attr": func() {
			cdp(`{"id":11,"method":"DOM.setAttribute","params":{"selector":"#cdp-target","name":"data-cdp","value":"button"}}`)
		},
		"btn-cdp-metrics": func() { cdp(`{"id":12,"method":"Page.getLayoutMetrics"}`) },
		"btn-cdp-unknown": func() { cdp(`{"id":13,"method":"Nope.nothing"}`) },
		"btn-inj-click": func() {
			h.injectClick("inj-target")
			h.setText("inj-out", fmt.Sprintf("injected click -> clicks=%d", h.clicks))
		},
		"btn-inj-type": func() {
			h.focus("inj-input")
			h.injectText("kui")
			h.setText("inj-out", "typed -> value="+h.inputValue("inj-input"))
		},
		"btn-inj-ime": func() {
			h.focus("ime-input")
			h.injectIME()
			h.setText("inj-out", "IME -> value="+h.inputValue("ime-input"))
		},
		"btn-inj-wheel": func() {
			h.injectWheel()
			h.setText("inj-out", fmt.Sprintf("injected wheel -> wheel=%d", h.wheels))
		},
		"btn-box": func() {
			info, _ := h.doc.ElementBoxModel("#box-target")
			var rect kui.Rect
			if element, err := h.doc.GetElementByID("box-target"); err == nil && element != nil {
				rect, _ = element.BoundingRect()
				element.Close()
			}
			pw, ph := quadSize(info.Padding)
			bw, bh := quadSize(info.Border)
			mw, mh := quadSize(info.Margin)
			h.setText("box-out", fmt.Sprintf("content=%.0fx%.0f padding=%.0fx%.0f border=%.0fx%.0f margin=%.0fx%.0f | rect=%.0fx%.0f",
				info.Width, info.Height, pw, ph, bw, bh, mw, mh, rect.Width, rect.Height))
		},
	}
	for id, action := range actions {
		button, err := h.doc.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("no #%s: %v", id, err)
		}
		ctx.Keep(button)
		action := action
		if _, err := button.AddEventListener("click", func(*kui.Event) { action() }); err != nil {
			return err
		}
	}
	return nil
}

func ready(ctx *runner.Context) error {
	h := &host{doc: ctx.Document}
	t := &tour{doc: ctx.Document}
	ctx.OnClose(func() {
		_ = h.client.Close()
		_ = h.cancelClient.Close()
	})
	if err := h.listen(ctx); err != nil {
		return err
	}
	h.devtoolsPort(ctx.Window)
	h.checkHTTP(t)
	h.checkDevToolsCommands(t)
	checkTokenAuth(t, ctx.Window)
	h.checkInjection(t)
	h.checkBoxModel(t)
	t.readyDone = true
	if t.pending == 0 {
		t.finish()
	}
	// Fails whatever asynchronous check has not answered within 2.6 s, then publishes.
	_, _ = kui.SetTimeout(func() {
		if !t.finished && t.pending > 0 {
			t.check("http.async-answers", false, fmt.Sprintf("%d request(s) never answered", t.pending))
		}
		t.finish()
	}, 2600)
	return h.buttons(ctx)
}
