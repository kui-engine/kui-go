//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_platform.go; edit that file, not this copy.
// API tour: platform services (examples/api_tours/platform). The host reads the monotonic clock,
// runs timeouts / intervals / idle callbacks / posted tasks (also from another goroutine), pumps the event loop by
// hand (PollEvents, WaitEvents), round-trips the clipboard (putting the original text back),
// queries the monitors and toggles F12 DevTools. One #tour-log row per check; #tour-result ends
// as "PASS n/n" once the timers have fired. Message box, file dialogs and OpenURL are only behind
// (grey) buttons: they are modal or open the system browser.
package main

import (
	"errors"
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------
// check: a documented behaviour; a failure makes the tour FAIL.
// skip:  not possible in this environment; logged, not counted.

type tour struct {
	doc         *kui.Document
	total       int
	failed      int
	failedNames []string
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
	if result == "fail" {
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

func (t *tour) skip(name, why string) { t.row(name, "skip", "skipped: "+why) }

func (t *tour) finish() {
	result, err := t.doc.GetElementByID("tour-result")
	if err != nil || result == nil {
		return
	}
	defer result.Close()
	text := fmt.Sprintf("PASS %d/%d", t.total, t.total)
	if t.failed > 0 {
		text = fmt.Sprintf("FAIL %d/%d: %v", t.failed, t.total, t.failedNames)
	}
	_ = result.SetTextContent(text)
	_ = result.SetAttribute("data-total", fmt.Sprint(t.total))
	_ = result.SetAttribute("data-failed", fmt.Sprint(t.failed))
	_ = result.SetAttribute("data-known", "0")
	if t.failed == 0 {
		_ = result.AddClass("pass")
	} else {
		_ = result.AddClass("fail")
	}
}

// ---- Host state -----------------------------------------------------------------------------

type host struct {
	win *kui.Window
	doc *kui.Document

	timeoutElapsed float64
	intervalTicks  int
	intervalID     uint64
	cancelledFired int
	idleRan        int
	postedRan      int
	buttonTicks    int
	buttonInterval uint64
}

func (h *host) setText(id, text string) {
	if element, err := h.doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

func now() float64 {
	value, _ := kui.Now()
	return value
}

func statusOf(err error) kui.Status {
	var kerr *kui.Error
	if errors.As(err, &kerr) {
		return kerr.Status
	}
	return 0
}

// ---- Clock, timers, idle callbacks ----------------------------------------------------------------------

func (h *host) checkClockAndTimers(t *tour) {
	version, versionErr := kui.RuntimeVersion()
	t.check("platform.runtime_version", versionErr == nil && version.Major >= 3,
		fmt.Sprintf("err %v version %d.%d.%d", versionErr, version.Major, version.Minor, version.Patch))

	first, err := kui.Now()
	second := now()
	t.check("time.now", err == nil && first > 0 && second >= first, fmt.Sprintf("err %v now %.3f then %.3f", err, first, second))

	h.timeoutElapsed = -1
	started := now()
	timer, err := kui.SetTimeout(func() { h.timeoutElapsed = now() - started }, 50)
	t.check("time.set_timeout", err == nil && timer != 0, fmt.Sprintf("err %v id %d", err, timer))
	// Clears itself on the third tick.
	h.intervalID, err = kui.SetInterval(func() {
		h.intervalTicks++
		if h.intervalTicks == 3 {
			_ = kui.ClearTimer(h.intervalID)
		}
	}, 30)
	t.check("time.set_interval", err == nil && h.intervalID != 0 && h.intervalID != timer, fmt.Sprintf("err %v id %d", err, h.intervalID))
	cancelled, _ := kui.SetTimeout(func() { h.cancelledFired++ }, 40)
	err = kui.ClearTimer(cancelled)
	t.check("time.clear_timer", err == nil, fmt.Sprint(err))
	err = kui.RequestIdle(func() { h.idleRan++ })
	t.check("time.request_idle", err == nil, fmt.Sprint(err))
}

// ---- Pumping the loop by hand: a posted task runs inside PollEvents / WaitEvents ----------

func (h *host) checkEventLoop(t *tour) {
	err := kui.PostTask(func() { h.postedRan++ })
	t.check("loop.post_to_main_thread-async", err == nil && h.postedRan == 0, fmt.Sprintf("err %v, ran %d time(s) before the loop", err, h.postedRan))
	err = kui.PollEvents()
	t.check("loop.poll_events-runs-posted-task", err == nil && h.postedRan == 1, fmt.Sprintf("err %v, posted task ran %d time(s)", err, h.postedRan))

	started := now()
	err = kui.WaitEvents(0.05)
	elapsed := now() - started
	t.check("loop.wait_events-timeout", err == nil && elapsed < 1000, fmt.Sprintf("err %v returned after %.1f ms", err, elapsed))
	// A task posted from another goroutine wakes a waiting loop at once and runs on this thread.
	before := h.postedRan
	done := make(chan struct{})
	go func() {
		_ = kui.PostTask(func() { h.postedRan++ })
		close(done)
	}()
	<-done
	started = now()
	err = kui.WaitEvents(2)
	elapsed = now() - started
	if h.postedRan == before {
		_ = kui.PollEvents()
	}
	t.check("loop.wait_events-woken-by-post", err == nil && h.postedRan == before+1 && elapsed < 1500,
		fmt.Sprintf("err %v, ran %d -> %d, waited %.1f ms of 2000", err, before, h.postedRan, elapsed))
}

// ---- Clipboard round trip (original text restored) -----------------------------------------

const sample = "KUI clipboard 你好 42"

// clipboardRoundTrip saves, writes the sample, reads it back and restores; a nil tour means a
// button press. The second result is false when the clipboard held no text (left alone).
func clipboardRoundTrip(t *tour) (string, bool) {
	if has, _ := kui.HasClipboardText(); !has {
		return "clipboard holds no text: left alone", false
	}
	original, err := kui.ClipboardText()
	if err != nil {
		return "could not read the clipboard: " + err.Error(), true
	}
	setErr := kui.SetClipboardText(sample)
	hasAfter, _ := kui.HasClipboardText()
	readBack, _ := kui.ClipboardText()
	ok := setErr == nil && hasAfter && readBack == sample
	if t != nil {
		t.check("clip.set_clipboard_text", setErr == nil && hasAfter, fmt.Sprintf("err %v has %v", setErr, hasAfter))
		t.check("clip.get_clipboard_text", readBack == sample, fmt.Sprintf("read back %q", readBack))
	}
	_ = kui.SetClipboardText(original)
	restored, _ := kui.ClipboardText()
	if t != nil {
		t.check("clip.restored", restored == original, fmt.Sprintf("original of %d bytes, now %d bytes", len(original), len(restored)))
	}
	verdict := "ok"
	if !ok {
		verdict = "FAILED"
	}
	return fmt.Sprintf("round trip %s (sample of %d bytes), original text (%d bytes) restored=%v",
		verdict, len(sample), len(original), restored == original), true
}

func checkClipboard(t *tour) {
	_, err := kui.HasClipboardText()
	t.check("clip.has_clipboard_text", err == nil, fmt.Sprint(err))
	if _, done := clipboardRoundTrip(t); !done {
		t.skip("clip.round-trip", "the clipboard holds no text now; it is not overwritten")
	}
}

// ---- Monitors ---------------------------------------------------------------------------------

func contains(m kui.MonitorInfo, x, y float32) bool {
	return x >= m.Bounds.X && y >= m.Bounds.Y && x < m.Bounds.X+m.Bounds.Width && y < m.Bounds.Y+m.Bounds.Height
}

// ---- Diagnostics: console, stack trace writer, crash reporter, input method -----------------------------

func checkDiagnostics(t *tour) {
	consoleErr := kui.SetConsoleEnabled(true)
	if consoleErr == nil {
		consoleErr = kui.SetConsoleStderrOutput(false)
	}
	if consoleErr == nil {
		consoleErr = kui.SetConsoleMinLevel(kui.LogDebug)
	}
	t.check("diag.console_configure", consoleErr == nil, fmt.Sprintf("err %v", consoleErr))

	stackInstalled, stackErr := kui.InstallStackTrace(kui.StackTraceConfig{OutputMode: kui.StackTraceStderr})
	stackUninstallErr := kui.UninstallStackTrace()
	t.check("diag.stack_trace_install_uninstall", stackErr == nil && stackUninstallErr == nil,
		fmt.Sprintf("installed %v, install err %v, uninstall err %v", stackInstalled, stackErr, stackUninstallErr))

	// Without crashpad_handler next to the executable the call succeeds and reports false.
	crashInstalled, crashErr := kui.InstallCrashReporter(kui.CrashReporterConfig{})
	crashUninstallErr := kui.UninstallCrashReporter()
	t.check("diag.crash_reporter_install_uninstall", crashErr == nil && crashUninstallErr == nil,
		fmt.Sprintf("installed %v, install err %v, uninstall err %v", crashInstalled, crashErr, crashUninstallErr))

	// Windows turns the input method off for this thread; other platforms report NOT_SUPPORTED.
	imeStatus := statusOf(kui.DisableIME())
	t.check("diag.disable_ime", imeStatus == 0 || imeStatus == kui.ErrNotSupported, fmt.Sprintf("status %d", imeStatus))
}

func (h *host) checkMonitors(t *tour) {
	monitors, err := kui.Monitors()
	t.check("mon.get_monitor_count", err == nil && len(monitors) >= 1, fmt.Sprintf("err %v count %d", err, len(monitors)))
	primaries, valid := 0, true
	var listedPrimary kui.MonitorInfo
	for _, m := range monitors {
		if m.Bounds.Width <= 0 || m.Bounds.Height <= 0 || m.WorkArea.Width <= 0 || m.WorkArea.Width > m.Bounds.Width || m.Scale <= 0 {
			valid = false
		}
		if m.Primary {
			primaries++
			listedPrimary = m
		}
	}
	t.check("mon.get_monitor", valid && primaries == 1, fmt.Sprintf("valid %v, %d primary monitor(s)", valid, primaries))

	primary, err := kui.PrimaryMonitor()
	t.check("mon.get_primary_monitor", err == nil && primary.Primary && primary.Bounds == listedPrimary.Bounds,
		fmt.Sprintf("err %v bounds %+v", err, primary.Bounds))
	cx, cy := primary.Bounds.X+primary.Bounds.Width/2, primary.Bounds.Y+primary.Bounds.Height/2
	fromPoint, err := kui.MonitorFromPoint(cx, cy)
	t.check("mon.get_monitor_from_point", err == nil && fromPoint.Bounds == primary.Bounds && fromPoint.Primary,
		fmt.Sprintf("err %v bounds %+v", err, fromPoint.Bounds))

	wx, wy, _ := h.win.Position()
	ww, wh, _ := h.win.Size()
	centreX, centreY := float32(wx)+float32(ww)/2, float32(wy)+float32(wh)/2
	fromWindow, err := kui.MonitorFromWindow(h.win)
	t.check("mon.get_monitor_from_window", err == nil && contains(fromWindow, centreX, centreY),
		fmt.Sprintf("err %v window centre (%.0f, %.0f), monitor %+v", err, centreX, centreY, fromWindow.Bounds))
}

func checkF12(t *tour) {
	on := kui.SetF12DevToolsEnabled(true)
	off := kui.SetF12DevToolsEnabled(false)
	t.check("shell.set_f12_devtools_enabled", on == nil && off == nil, fmt.Sprintf("on %v off %v", on, off))
}

// ---- Verdict once the timers have had their turn ----------------------------------------

func (h *host) finalize(t *tour) {
	t.check("time.set_timeout-fired", h.timeoutElapsed >= 45, fmt.Sprintf("fired after %.1f ms (-1 = never)", h.timeoutElapsed))
	t.check("time.set_interval-ticks", h.intervalTicks == 3, fmt.Sprintf("%d tick(s), cleared on the third", h.intervalTicks))
	t.check("time.clear_timer-cancelled", h.cancelledFired == 0, fmt.Sprintf("cleared timeout ran %d time(s)", h.cancelledFired))
	t.check("time.request_idle-ran", h.idleRan == 1, fmt.Sprintf("idle callback ran %d time(s)", h.idleRan))
	t.finish()
}

// ---- Panel buttons -----------------------------------------------------------------------

func describe(m kui.MonitorInfo) string {
	text := fmt.Sprintf("%.0fx%.0f at (%.0f, %.0f) work %.0fx%.0f scale %.2f", m.Bounds.Width, m.Bounds.Height,
		m.Bounds.X, m.Bounds.Y, m.WorkArea.Width, m.WorkArea.Height, m.Scale)
	if m.Name != "" {
		text = m.Name + ": " + text
	}
	if m.Primary {
		text += " primary"
	}
	return text
}

var filters = []string{"HTML pages|*.html;*.htm", "All files|*.*"}

func (h *host) showPath(what, path string, chosen bool, err error) {
	switch {
	case err != nil:
		h.setText("shell-out", fmt.Sprintf("%s: %v", what, err))
	case !chosen:
		h.setText("shell-out", what+": cancelled")
	default:
		h.setText("shell-out", fmt.Sprintf("%s: status 0 %s", what, path))
	}
}

func (h *host) buttons(ctx *runner.Context) error {
	actions := map[string]func(){
		"btn-timeout": func() {
			started := now()
			_, _ = kui.SetTimeout(func() {
				if now()-started >= 95 {
					h.setText("time-out", "timeout fired after >= 100 ms")
				} else {
					h.setText("time-out", "timeout fired after < 100 ms")
				}
			}, 100)
			h.setText("time-out", "waiting 100 ms...")
		},
		"btn-interval": func() {
			h.buttonTicks = 0
			h.buttonInterval, _ = kui.SetInterval(func() {
				h.buttonTicks++
				if h.buttonTicks == 3 {
					_ = kui.ClearTimer(h.buttonInterval)
				}
				h.setText("time-out", fmt.Sprintf("interval tick %d of 3", h.buttonTicks))
			}, 50)
		},
		"btn-cancel": func() {
			id, _ := kui.SetTimeout(func() { h.setText("time-out", "the cleared timer fired (bug)") }, 30)
			h.setText("time-out", fmt.Sprintf("timer set and cleared: status %d", statusOf(kui.ClearTimer(id))))
		},
		"btn-idle": func() { _ = kui.RequestIdle(func() { h.setText("time-out", "idle callback ran") }) },
		"btn-post": func() {
			go func() { _ = kui.PostTask(func() { h.setText("time-out", "posted task ran on the UI thread") }) }()
		},
		"btn-poll": func() {
			ran := false
			_ = kui.PostTask(func() { ran = true })
			_ = kui.PollEvents()
			answer := "no"
			if ran {
				answer = "yes"
			}
			h.setText("loop-out", "posted task ran inside poll_events: "+answer)
		},
		"btn-wait": func() {
			started := now()
			err := kui.WaitEvents(0.05)
			answer := "no"
			if now()-started < 1000 {
				answer = "yes"
			}
			h.setText("loop-out", fmt.Sprintf("wait_events(0.05): status %d, returned within 1 s: %s", statusOf(err), answer))
		},
		"btn-clip-read": func() {
			has, _ := kui.HasClipboardText()
			size := 0
			if has {
				text, _ := kui.ClipboardText()
				size = len(text)
			}
			flag := 0
			if has {
				flag = 1
			}
			h.setText("clip-out", fmt.Sprintf("has text=%d (%d bytes)", flag, size))
		},
		"btn-clip-trip": func() {
			summary, _ := clipboardRoundTrip(nil)
			h.setText("clip-out", summary)
		},
		"btn-mon-list": func() {
			monitors, _ := kui.Monitors()
			text := fmt.Sprintf("%d monitor(s)", len(monitors))
			for _, m := range monitors {
				text += " | " + describe(m)
			}
			h.setText("mon-out", text)
		},
		"btn-mon-window": func() {
			m, err := kui.MonitorFromWindow(h.win)
			if err != nil {
				h.setText("mon-out", "monitor lookup failed")
				return
			}
			h.setText("mon-out", "this window: "+describe(m))
		},
		// Manual: modal.
		"btn-msg": func() {
			result, _ := kui.MessageBox("KUI API tour", "Do you like message boxes?", kui.MessageBoxQuestion, kui.MessageBoxYesNo)
			answer := "no"
			if result == kui.MessageBoxResultYes {
				answer = "yes"
			}
			h.setText("shell-out", "message box answer: "+answer)
		},
		// Manual: the three file dialogs are modal.
		"btn-open": func() {
			path, chosen, err := kui.OpenFileDialog(kui.FileDialogOptions{Title: "Open a page", Filters: filters})
			h.showPath("open", path, chosen, err)
		},
		"btn-save": func() {
			path, chosen, err := kui.SaveFileDialog(kui.FileDialogOptions{Title: "Save the page as", DefaultPath: "tour.html", Filters: filters})
			h.showPath("save", path, chosen, err)
		},
		"btn-folder": func() {
			path, chosen, err := kui.SelectFolderDialog("Pick a folder")
			h.showPath("folder", path, chosen, err)
		},
		// Manual: opens the system browser.
		"btn-url": func() {
			h.setText("shell-out", fmt.Sprintf("open_url: status %d", statusOf(kui.OpenURL("https://example.com/"))))
		},
		"btn-f12-on": func() {
			if kui.SetF12DevToolsEnabled(true) == nil {
				h.setText("shell-out", "F12 enabled: F12 opens DevTools in the browser")
			}
		},
		"btn-f12-off": func() {
			if kui.SetF12DevToolsEnabled(false) == nil {
				h.setText("shell-out", "F12 disabled")
			}
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
	h := &host{win: ctx.Window, doc: ctx.Document}
	t := &tour{doc: ctx.Document}
	h.checkClockAndTimers(t)
	h.checkEventLoop(t)
	checkClipboard(t)
	h.checkMonitors(t)
	checkF12(t)
	checkDiagnostics(t)
	_, _ = kui.SetTimeout(func() { h.finalize(t) }, 700)
	return h.buttons(ctx)
}
