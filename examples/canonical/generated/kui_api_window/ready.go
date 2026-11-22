// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_window.go; edit that file, not this copy.
// API tour: the native window (examples/api_tours/window). The host changes title, size and
// position, listens for resizes (callback and "resize" event listeners), cycles maximize and
// fullscreen, shapes the frame, sets opacity / z-order / cursor / IME position, installs a native
// menu bar, binds keyboard shortcuts, drives a hidden tray icon with simulated clicks, reads the
// DevTools state and saves a screenshot -- then checks the window is back exactly as it was.
// One #tour-log row per check; #tour-result ends as "PASS n/n". Minimize, stop DevTools and the
// tray context menu are only behind (grey) buttons: they would stop rendering, cut the
// inspector's connection or open a system menu.
package main

/*
#include <stdint.h>
#ifdef _WIN32
#include <windows.h>
// A real key-down / key-up for the native window: the path a keyboard takes, where window
// shortcuts are matched. SendMessageW is looked up at run time (user32 is already loaded).
static int tour_send_key(uintptr_t hwnd, unsigned vk) {
	typedef LRESULT (WINAPI *SendFn)(HWND, UINT, WPARAM, LPARAM);
	HMODULE user32 = GetModuleHandleW(L"user32.dll");
	SendFn send = user32 ? (SendFn)(void*)GetProcAddress(user32, "SendMessageW") : NULL;
	if (!send || !hwnd) return 0;
	send((HWND)hwnd, WM_KEYDOWN, (WPARAM)vk, (LPARAM)1);
	send((HWND)hwnd, WM_KEYUP, (WPARAM)vk, (LPARAM)0xC0000001u);
	return 1;
}
#else
static int tour_send_key(uintptr_t hwnd, unsigned vk) { (void)hwnd; (void)vk; return 0; }
#endif
*/
import "C"

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

const vkF7 = 0x76

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------
// check:    a documented behaviour; a failure makes the tour FAIL.
// knownBug: a documented behaviour the engine currently gets wrong (reported bug).
// skip:     not available on this platform; logged, not counted.

type tour struct {
	doc         *kui.Document
	total       int
	failed      int
	known       int
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
	if result == "fail" || result == "known" {
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

// ---- Host state -----------------------------------------------------------------------------

type geometry struct{ w, h, x, y int }

type host struct {
	win    *kui.Window
	doc    *kui.Document
	script *kui.ScriptEngine

	origTitle string
	orig      geometry
	origFrame kui.FrameState

	resizeCalls      int
	resizeW, resizeH uint32
	eventsA, eventsB int
	f7First, f7Hits  int
	ctrlKHits        int
	buttonHits       int
	trayEvents       []kui.TrayEvent
	menuChecked      bool
	menuEnabled      bool
	titleToggled     bool
}

func (h *host) setText(id, text string) {
	if element, err := h.doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

func (h *host) geometry() geometry {
	var g geometry
	g.w, g.h, _ = h.win.Size()
	g.x, g.y, _ = h.win.Position()
	return g
}

func (h *host) title() string {
	title, _ := h.win.Title()
	return title
}

func (h *host) luaNumber(code string) float64 {
	result, err := h.script.Eval(code, "tour")
	if err != nil || !result.Success || result.Value.Type != kui.ScriptString {
		return -1
	}
	value, err := strconv.ParseFloat(result.Value.String, 64)
	if err != nil {
		return -1
	}
	return value
}

func boolOf(value bool, err error) int {
	if err != nil {
		return -1
	}
	if value {
		return 1
	}
	return 0
}

func statusOf(err error) kui.Status {
	var kerr *kui.Error
	if errors.As(err, &kerr) {
		return kerr.Status
	}
	return 0
}

func near(a, b float32) bool { return math.Abs(float64(a-b)) < 0.01 }

// sendNativeKey delivers a real key to the native window (Windows only; false elsewhere).
func (h *host) sendNativeKey(vk uint) bool {
	handle, err := h.win.NativeHandle()
	if err != nil {
		return false
	}
	return C.tour_send_key(C.uintptr_t(handle), C.unsigned(vk)) != 0
}

// ---- Title, size, position, resize notifications ------------------------------------------

func (h *host) showResize() {
	h.setText("resize-out", fmt.Sprintf("on_resize %dx%d (callbacks=%d)", h.resizeW, h.resizeH, h.resizeCalls))
}

func (h *host) checkGeometry(t *tour) {
	const hostTitle = "kui_api_window: title set by the host"
	t.check("win.get_title", h.origTitle == "KUI API — Window", fmt.Sprintf("title %q", h.origTitle))
	err := h.win.SetTitle(hostTitle)
	t.check("win.set_title", err == nil && h.title() == hostTitle, fmt.Sprintf("err %v title %q", err, h.title()))
	_ = h.win.SetTitle(h.origTitle)
	t.check("win.set_title-restore", h.title() == h.origTitle, fmt.Sprintf("title %q", h.title()))

	ratio, err := h.win.DevicePixelRatio()
	luaRatio := h.luaNumber("return tostring(window.devicePixelRatio())")
	t.check("win.get_device_pixel_ratio", err == nil && ratio > 0 && near(ratio, float32(luaRatio)),
		fmt.Sprintf("err %v ratio %.3f, Lua window.devicePixelRatio() %.3f", err, ratio, luaRatio))
	// Size is the client area in device pixels; the page sees it in CSS px.
	inner := h.luaNumber("return tostring(window.innerWidth())")
	t.check("win.get_size", h.orig.w > 0 && h.orig.h > 0 && math.Abs(inner*float64(ratio)-float64(h.orig.w)) < 1.5,
		fmt.Sprintf("size %dx%d, innerWidth %.1f x ratio %.2f", h.orig.w, h.orig.h, inner, ratio))
	handle, err := h.win.NativeHandle()
	t.check("win.get_native_handle", err == nil && handle != 0, fmt.Sprintf("err %v handle %#x", err, handle))

	// Two "resize" listeners and the resize callback, then remove them one by one.
	_ = h.win.OnResize(func(w, hgt uint32) {
		h.resizeCalls++
		h.resizeW, h.resizeH = w, hgt
		h.showResize()
	})
	idA, _ := h.win.AddEventListener("resize", false, func(*kui.Event) { h.eventsA++ })
	_, _ = h.win.AddEventListener("resize", false, func(*kui.Event) { h.eventsB++ })
	err = h.win.Resize(h.orig.w-60, h.orig.h-40)
	g := h.geometry()
	t.check("win.resize", err == nil && g.w == h.orig.w-60 && g.h == h.orig.h-40,
		fmt.Sprintf("err %v size %dx%d, wanted %dx%d", err, g.w, g.h, h.orig.w-60, h.orig.h-40))
	t.check("win.on_resize", h.resizeCalls > 0 && int(h.resizeW) == g.w && int(h.resizeH) == g.h,
		fmt.Sprintf("%d callback(s), last %dx%d", h.resizeCalls, h.resizeW, h.resizeH))
	t.check("win.add_event_listener-resize", h.eventsA > 0 && h.eventsB > 0, fmt.Sprintf("listener a %d, b %d", h.eventsA, h.eventsB))

	a, b := h.eventsA, h.eventsB
	err = h.win.RemoveEventListener("resize", idA)
	_ = h.win.Resize(h.orig.w-30, h.orig.h-20)
	t.check("win.remove_event_listener", err == nil && h.eventsA == a && h.eventsB > b,
		fmt.Sprintf("err %v, removed listener %d -> %d, kept listener %d -> %d", err, a, h.eventsA, b, h.eventsB))
	a, b, calls := h.eventsA, h.eventsB, h.resizeCalls
	err = h.win.RemoveEventListeners("resize")
	_ = h.win.Resize(h.orig.w, h.orig.h)
	g = h.geometry()
	t.check("win.remove_event_listeners", err == nil && h.eventsA == a && h.eventsB == b && h.resizeCalls > calls,
		fmt.Sprintf("err %v, listeners %d/%d -> %d/%d, callback %d -> %d", err, a, b, h.eventsA, h.eventsB, calls, h.resizeCalls))
	t.check("win.resize-back", g.w == h.orig.w && g.h == h.orig.h, fmt.Sprintf("size %dx%d", g.w, g.h))

	err = h.win.SetPosition(h.orig.x+24, h.orig.y+16)
	g = h.geometry()
	t.check("win.get_position", err == nil && g.x == h.orig.x+24 && g.y == h.orig.y+16,
		fmt.Sprintf("err %v moved to (%d, %d), wanted (%d, %d)", err, g.x, g.y, h.orig.x+24, h.orig.y+16))
	_ = h.win.SetPosition(h.orig.x, h.orig.y)
}

// ---- Maximize / restore, fullscreen, focus, close flag ------------------------------------

func (h *host) checkState(t *tour) {
	minimized := boolOf(h.win.IsMinimized())
	t.check("state.is_minimized", minimized == 0, fmt.Sprintf("is_minimized %d", minimized))
	closing := boolOf(h.win.ShouldClose())
	t.check("state.should_close", closing == 0, fmt.Sprintf("should_close %d", closing))
	focused, err := h.win.HasFocus()
	t.check("state.has_focus", err == nil, fmt.Sprintf("err %v focused %v", err, focused))

	maximized := boolOf(h.win.IsMaximized())
	t.check("state.is_maximized-initial", maximized == 0, fmt.Sprintf("is_maximized %d", maximized))
	err = h.win.Maximize()
	g := h.geometry()
	maximized = boolOf(h.win.IsMaximized())
	t.check("state.maximize", err == nil && maximized == 1 && g.w >= h.orig.w,
		fmt.Sprintf("err %v is_maximized %d size %dx%d", err, maximized, g.w, g.h))
	err = h.win.Restore()
	g = h.geometry()
	maximized = boolOf(h.win.IsMaximized())
	t.check("state.restore", err == nil && maximized == 0 && g.w == h.orig.w && g.h == h.orig.h,
		fmt.Sprintf("err %v is_maximized %d size %dx%d", err, maximized, g.w, g.h))

	full := boolOf(h.win.IsFullscreen())
	t.check("state.is_fullscreen-initial", full == 0, fmt.Sprintf("is_fullscreen %d", full))
	err = h.win.SetFullscreen(true)
	g = h.geometry()
	full = boolOf(h.win.IsFullscreen())
	t.check("state.set_fullscreen", err == nil && full == 1 && g.w >= h.orig.w,
		fmt.Sprintf("err %v is_fullscreen %d size %dx%d", err, full, g.w, g.h))
	err = h.win.SetFullscreen(false)
	g = h.geometry()
	full = boolOf(h.win.IsFullscreen())
	t.check("state.set_fullscreen-exit", err == nil && full == 0 && g == h.orig,
		fmt.Sprintf("err %v is_fullscreen %d geometry %+v, was %+v", err, full, g, h.orig))
}

// ---- Frame shape: radii, frameless, shadow ----------------------------------------------

// ---- Host integration: one frame now, enable / disable, render DPR, base path ----------------------------

func (h *host) checkHost(t *tour) {
	err := h.win.ProcessFrame()
	t.check("host.process_frame", err == nil, fmt.Sprintf("err %v", err))

	// Enabling / disabling input is Windows only; other platforms report NOT_SUPPORTED.
	wasEnabled, readErr := h.win.IsEnabled()
	if statusOf(readErr) == kui.ErrNotSupported {
		setErr := h.win.SetEnabled(false)
		t.check("host.set_enabled", statusOf(setErr) == kui.ErrNotSupported, fmt.Sprintf("status %d", statusOf(setErr)))
	} else {
		_ = h.win.SetEnabled(false)
		disabled, _ := h.win.IsEnabled()
		_ = h.win.SetEnabled(true)
		now, _ := h.win.IsEnabled()
		t.check("host.set_enabled", readErr == nil && wasEnabled && !disabled && now,
			fmt.Sprintf("was %v, disabled %v, now %v", wasEnabled, disabled, now))
	}

	// A window created hidden (no flash), its first frame rendered while it is still hidden, then shown.
	radius := float32(8)
	extra, createErr := kui.NewWindow(kui.WindowProps{Title: "kui-hidden", Width: 320, Height: 200, Frameless: true,
		Sizeable: true, Hidden: true, FrameRadius: &radius})
	hidden, shown := false, false
	if createErr == nil {
		visible, _ := extra.IsVisible()
		hidden = !visible
		_ = extra.ProcessFrame()
		_ = extra.Show()
		shown, _ = extra.IsVisible()
		_ = extra.Close()
	}
	t.check("host.create_hidden_window", createErr == nil && hidden && shown,
		fmt.Sprintf("err %v, hidden at creation %v, visible after show %v", createErr, hidden, shown))

	// Setting the render ratio to the value it already has leaves the page as it is.
	ratio, _ := h.win.DevicePixelRatio()
	ratioErr := h.doc.SetRenderDevicePixelRatio(ratio)
	t.check("host.set_render_device_pixel_ratio", ratioErr == nil, fmt.Sprintf("err %v, ratio %.3f", ratioErr, ratio))

	original, _ := h.doc.BasePath()
	_ = h.doc.SetBasePath("assets/")
	changed, _ := h.doc.BasePath()
	_ = h.doc.SetBasePath(original)
	restored, _ := h.doc.BasePath()
	t.check("host.set_get_base_path", changed == "assets/" && restored == original,
		fmt.Sprintf("changed '%s', restored '%s'", changed, restored))
}

func (h *host) checkFrame(t *tour) {
	_, ok, err := h.win.FrameRadii()
	t.check("frame.get_frame_radii-none", err == nil && !ok, fmt.Sprintf("err %v set %v", err, ok))
	_ = h.win.SetFrameRadius(12)
	r, ok, err := h.win.FrameRadii()
	t.check("frame.set_frame_radius", err == nil && ok && r == kui.FrameRadii{TopLeft: 12, TopRight: 12, BottomRight: 12, BottomLeft: 12},
		fmt.Sprintf("err %v radii %+v", err, r))
	want := kui.FrameRadii{TopLeft: 16, TopRight: 4, BottomRight: 16, BottomLeft: 4}
	_ = h.win.SetFrameRadii(want)
	r, ok, err = h.win.FrameRadii()
	t.check("frame.set_frame_radii", err == nil && ok && r == want, fmt.Sprintf("err %v radii %+v", err, r))
	// A negative radius is not a valid shape: the call changes nothing.
	_ = h.win.SetFrameRadii(kui.FrameRadii{TopLeft: -1, TopRight: 4, BottomRight: 4, BottomLeft: 4})
	r, ok, err = h.win.FrameRadii()
	t.check("frame.set_frame_radii-negative-ignored", err == nil && ok && r == want, fmt.Sprintf("err %v radii %+v", err, r))
	err = h.win.ClearFrameRadius()
	_, ok, _ = h.win.FrameRadii()
	t.check("frame.clear_frame_radius", err == nil && !ok, fmt.Sprintf("err %v still set %v", err, ok))

	err = h.win.SetFrameless(true)
	state, _ := h.win.FrameState()
	g := h.geometry()
	t.check("frame.set_frameless", err == nil && state.Frameless && g.w == h.orig.w && g.h == h.orig.h,
		fmt.Sprintf("err %v state %+v client %dx%d", err, state, g.w, g.h))
	err = h.win.SetFramelessShadow(true)
	state, _ = h.win.FrameState()
	t.check("frame.set_frameless_shadow", err == nil && state.Shadow, fmt.Sprintf("err %v state %+v", err, state))
	_ = h.win.SetFramelessShadow(h.origFrame.Shadow)
	_ = h.win.SetFrameless(h.origFrame.Frameless)
	state, _ = h.win.FrameState()
	g = h.geometry()
	t.check("frame.get_frame_state-restored", state == h.origFrame && g.w == h.orig.w && g.h == h.orig.h,
		fmt.Sprintf("state %+v (was %+v) client %dx%d", state, h.origFrame, g.w, g.h))
}

// ---- Opacity, z-order, cursor, IME candidate position ----------------------------------

func (h *host) imeCaret() (float32, float32) {
	var rect kui.Rect
	if input, err := h.doc.GetElementByID("ime-input"); err == nil && input != nil {
		rect, _ = input.BoundingRect()
		input.Close()
	}
	return rect.X + 4, rect.Y + rect.Height
}

func (h *host) checkLook(t *tour) {
	err := h.win.SetOpacity(0.85)
	low := h.win.SetOpacity(-0.1)
	high := h.win.SetOpacity(1.5)
	t.check("look.set_opacity", err == nil && statusOf(low) == kui.ErrInvalidArg && statusOf(high) == kui.ErrInvalidArg &&
		h.win.SetOpacity(1) == nil, fmt.Sprintf("0.85 -> %v, -0.1 -> %v, 1.5 -> %v", err, low, high))
	err = h.win.SetAlwaysOnTop(true)
	t.check("look.set_always_on_top", err == nil && h.win.SetAlwaysOnTop(false) == nil, fmt.Sprint(err))
	err = h.win.SetCursor(kui.CursorCrosshair)
	bad := h.win.SetCursor(kui.Cursor(99))
	t.check("look.set_cursor", err == nil && statusOf(bad) == kui.ErrInvalidArg && h.win.SetCursor(kui.CursorDefault) == nil,
		fmt.Sprintf("crosshair -> %v, 99 -> %v", err, bad))
	err = h.win.SetCursorVisible(false)
	t.check("look.set_cursor_visible", err == nil && h.win.SetCursorVisible(true) == nil, fmt.Sprint(err))
	x, y := h.imeCaret()
	err = h.win.SetIMECandidatePos(x, y)
	t.check("look.set_ime_candidate_pos", err == nil, fmt.Sprintf("err %v at (%.0f, %.0f)", err, x, y))
}

// ---- Native menu bar -------------------------------------------------------------------------

const (
	menuWrap    = 901
	menuReload  = 902
	menuZoomIn  = 903
	menuZoomOut = 904
)

func (h *host) installMenu() error {
	h.menuChecked, h.menuEnabled = true, true
	return h.win.SetMenuBar([]kui.Menu{{
		Text: "&View",
		Items: []kui.MenuItem{
			{Text: "&Word wrap", ID: menuWrap, Checkable: true, Checked: true},
			{Separator: true},
			{Text: "&Reload", ID: menuReload},
			{Text: "&Zoom", Children: []kui.MenuItem{
				{Text: "Zoom &in", Shortcut: "Ctrl++", ID: menuZoomIn},
				{Text: "Zoom &out", ID: menuZoomOut},
			}},
		},
	}}, func(id uint32) { h.setText("menu-out", fmt.Sprintf("menu item %d activated", id)) })
}

func (h *host) checkMenu(t *tour) {
	err := h.win.SetMenuItemChecked(menuWrap, false)
	t.check("menu.set_checked-no-menu", statusOf(err) == kui.ErrNotFound, fmt.Sprint(err))
	err = h.installMenu()
	g := h.geometry()
	t.check("menu.set_menu_bar", err == nil, fmt.Sprint(err))
	if runtime.GOOS == "windows" {
		// A native menu bar takes its height from the client area.
		t.check("menu.set_menu_bar-client-area", g.w == h.orig.w && g.h < h.orig.h,
			fmt.Sprintf("client %dx%d, was %dx%d", g.w, g.h, h.orig.w, h.orig.h))
	} else {
		t.skip("menu.set_menu_bar-client-area", "the menu bar only takes client height on Windows")
	}
	top := h.win.SetMenuItemChecked(menuWrap, false)
	nested := h.win.SetMenuItemChecked(menuZoomOut, true)
	missing := h.win.SetMenuItemChecked(999, true)
	t.check("menu.menu_set_checked", top == nil && nested == nil && statusOf(missing) == kui.ErrNotFound,
		fmt.Sprintf("top-level %v, submenu %v, unknown id %v", top, nested, missing))
	err = h.win.SetMenuItemEnabled(menuReload, false)
	missing = h.win.SetMenuItemEnabled(999, false)
	t.check("menu.menu_set_enabled", err == nil && statusOf(missing) == kui.ErrNotFound && h.win.SetMenuItemEnabled(menuReload, true) == nil,
		fmt.Sprintf("err %v, unknown id %v", err, missing))
	err = h.win.ClearMenuBar()
	missing = h.win.SetMenuItemChecked(menuWrap, true)
	g = h.geometry()
	t.check("menu.clear_menu_bar", err == nil && statusOf(missing) == kui.ErrNotFound && g.w == h.orig.w && g.h == h.orig.h,
		fmt.Sprintf("err %v, set_checked afterwards %v, client %dx%d", err, missing, g.w, g.h))
}

// ---- Keyboard shortcuts --------------------------------------------------------------------

func (h *host) checkShortcuts(t *tour) {
	f7, err := h.win.BindShortcut("F7", func() { h.f7First++ })
	ctrlK, _ := h.win.BindShortcut("Ctrl+Shift+K", func() { h.ctrlKHits++ })
	t.check("sc.bind_shortcut", err == nil && f7 != 0 && ctrlK != 0 && ctrlK != f7, fmt.Sprintf("err %v ids %d / %d", err, f7, ctrlK))
	again, err := h.win.BindShortcut("F7", func() { h.f7Hits++ })
	t.check("sc.bind_shortcut-rebind-same-id", err == nil && again == f7, fmt.Sprintf("err %v id %d, first id %d", err, again, f7))
	_, invalid := h.win.BindShortcut("Ctrl+", func() {})
	_, empty := h.win.BindShortcut("", func() {})
	t.check("sc.bind_shortcut-invalid", statusOf(invalid) == kui.ErrInvalidArg && statusOf(empty) == kui.ErrInvalidArg,
		fmt.Sprintf("'Ctrl+' -> %v, '' -> %v", invalid, empty))

	// Synthetic DevTools input goes to the page's event dispatcher.
	_ = h.doc.DispatchKey(kui.KeyDown, "F7", "F7", 0, "", false)
	_ = h.doc.DispatchKey(kui.KeyUp, "F7", "F7", 0, "", false)
	t.knownBug("sc.devtools-key-reaches-shortcut", h.f7Hits == 1,
		"devtools_dispatch_key_event bypasses the window ShortcutManager (only the native key-down path calls tryDispatch), so injected keys never trigger bound shortcuts",
		fmt.Sprintf("F7 callback ran %d time(s)", h.f7Hits))

	hits := h.f7Hits
	if h.sendNativeKey(vkF7) {
		t.check("sc.native-key-fires", h.f7Hits == hits+1 && h.f7First == 0,
			fmt.Sprintf("replacing callback %d -> %d, replaced callback %d", hits, h.f7Hits, h.f7First))
		hits = h.f7Hits
		err = h.win.SetShortcutsEnabled(false)
		h.sendNativeKey(vkF7)
		t.check("sc.set_shortcuts_enabled-off", err == nil && h.f7Hits == hits, fmt.Sprintf("err %v hits %d -> %d", err, hits, h.f7Hits))
		err = h.win.SetShortcutsEnabled(true)
		h.sendNativeKey(vkF7)
		t.check("sc.set_shortcuts_enabled-on", err == nil && h.f7Hits == hits+1, fmt.Sprintf("err %v hits %d -> %d", err, hits, h.f7Hits))
		hits = h.f7Hits
		err = h.win.UnbindShortcut(f7)
		h.sendNativeKey(vkF7)
		t.check("sc.unbind_shortcut", err == nil && h.f7Hits == hits, fmt.Sprintf("err %v hits %d -> %d", err, hits, h.f7Hits))
		_, _ = h.win.BindShortcut("F7", func() { h.f7Hits++ })
		err = h.win.UnbindAllShortcuts()
		h.sendNativeKey(vkF7)
		t.check("sc.unbind_all_shortcuts", err == nil && h.f7Hits == hits, fmt.Sprintf("err %v hits %d -> %d", err, hits, h.f7Hits))
	} else {
		t.skip("sc.native-key-fires", "no native key injection on this platform")
		err = h.win.SetShortcutsEnabled(false)
		t.check("sc.set_shortcuts_enabled", err == nil && h.win.SetShortcutsEnabled(true) == nil, fmt.Sprint(err))
		t.check("sc.unbind_shortcut", h.win.UnbindShortcut(f7) == nil, "")
		t.check("sc.unbind_all_shortcuts", h.win.UnbindAllShortcuts() == nil, "")
	}
	t.check("sc.ctrl-shift-k-untouched", h.ctrlKHits == 0, fmt.Sprintf("Ctrl+Shift+K ran %d time(s)", h.ctrlKHits))
}

// ---- Tray icon (created hidden; simulated clicks) -------------------------------------

var trayNames = map[kui.TrayEvent]string{kui.TrayClick: "click", kui.TrayDoubleClick: "double click",
	kui.TrayRightClick: "right click", kui.TrayBalloonClick: "balloon click"}

// ensureTray creates a tray icon with the application icon; false when the platform has none.
func (h *host) ensureTray(visible bool) bool {
	if active, _ := h.win.TrayActive(); !active {
		if h.win.CreateTray("", "KUI API tour", visible) != nil {
			return false
		}
	}
	if visible {
		_ = h.win.SetTrayVisible(true)
	}
	_ = h.win.SetTrayEventHandler(func(event kui.TrayEvent) {
		h.trayEvents = append(h.trayEvents, event)
		h.setText("tray-out", fmt.Sprintf("tray callback: %s (events=%d)", trayNames[event], len(h.trayEvents)))
	})
	return true
}

func (h *host) checkTray(t *tour) {
	active, _ := h.win.TrayActive()
	visible, _ := h.win.TrayVisible()
	t.check("tray.is_active-before-create", !active && !visible, fmt.Sprintf("active %v visible %v", active, visible))
	if !h.ensureTray(false) {
		t.skip("tray.simulate_event", "this platform has no tray icon")
		return
	}
	active, _ = h.win.TrayActive()
	visible, _ = h.win.TrayVisible()
	t.check("tray.tray_is_active-hidden", active && !visible, fmt.Sprintf("active %v visible %v", active, visible))
	h.trayEvents = nil
	_ = h.win.SimulateTrayEvent(kui.TrayClick)
	_ = h.win.SimulateTrayEvent(kui.TrayDoubleClick)
	_ = h.win.SimulateTrayEvent(kui.TrayRightClick)
	err := h.win.SimulateTrayEvent(kui.TrayBalloonClick)
	want := []kui.TrayEvent{kui.TrayClick, kui.TrayDoubleClick, kui.TrayRightClick, kui.TrayBalloonClick}
	t.check("tray.tray_simulate_event", err == nil && fmt.Sprint(h.trayEvents) == fmt.Sprint(want),
		fmt.Sprintf("err %v events %v", err, h.trayEvents))
	err = h.win.SetTrayIcon("")
	active, _ = h.win.TrayActive()
	t.check("tray.tray_set_icon", err == nil && active, fmt.Sprintf("err %v active %v", err, active))
	_ = h.win.DestroyTray()
	active, _ = h.win.TrayActive()
	t.check("tray.destroy", !active, fmt.Sprintf("active %v", active))
}

// ---- DevTools state and screenshot -----------------------------------------------------

// pngSize is the size of a PNG file, 0 when missing or not a PNG.
func pngSize(path string) int {
	data, err := os.ReadFile(path)
	if err != nil || len(data) < 8 || string(data[:8]) != "\x89PNG\r\n\x1a\n" {
		return 0
	}
	return len(data)
}

func (h *host) checkDevToolsScreenshot(t *tour) {
	port, _ := h.win.DevToolsPort()
	running, statePort, clients, err := h.win.DevToolsState()
	t.check("dt.get_devtools_state", err == nil && running && port != 0 && statePort == port,
		fmt.Sprintf("err %v running %v port %d (DevToolsPort %d) clients %d", err, running, statePort, port, clients))

	path := filepath.Join(os.TempDir(), "kui_api_window_shot_go.png")
	_ = os.Remove(path)
	saved, err := h.win.SaveScreenshot(path)
	size := pngSize(path)
	t.check("dt.save_screenshot", err == nil && saved && size > 100, fmt.Sprintf("err %v saved %v, %q holds %d PNG bytes", err, saved, path, size))
	_ = os.Remove(path)
	saved, err = h.win.SaveScreenshot(filepath.Join(os.TempDir(), "kui-api-window-no-such-dir", "shot.png"))
	t.check("dt.save_screenshot-unwritable", err == nil && !saved, fmt.Sprintf("err %v saved %v", err, saved))
}

// ---- Everything back as it was ---------------------------------------------------------

func (h *host) checkRestored(t *tour) {
	g := h.geometry()
	state, _ := h.win.FrameState()
	_, radius, _ := h.win.FrameRadii()
	maximized, _ := h.win.IsMaximized()
	full, _ := h.win.IsFullscreen()
	t.check("win.state-restored", h.title() == h.origTitle && g == h.orig && !maximized && !full &&
		state == h.origFrame && !radius,
		fmt.Sprintf("title %q geometry %+v (was %+v) maximized %v fullscreen %v frame %+v radius %v",
			h.title(), g, h.orig, maximized, full, state, radius))
}

// ---- Panel buttons -----------------------------------------------------------------------

func (h *host) showGeometry() {
	g := h.geometry()
	ratio, _ := h.win.DevicePixelRatio()
	handle, _ := h.win.NativeHandle()
	hasHandle := "no"
	if handle != 0 {
		hasHandle = "yes"
	}
	h.setText("geom-out", fmt.Sprintf("size=%dx%d pos=(%d, %d) dpr=%.2f handle=%s title=%s", g.w, g.h, g.x, g.y, ratio, hasHandle, h.title()))
}

func (h *host) showState(prefix string) {
	focused, _ := h.win.HasFocus()
	h.setText("state-out", fmt.Sprintf("%smaximized=%d minimized=%d fullscreen=%d focused=%d should_close=%d", prefix,
		boolOf(h.win.IsMaximized()), boolOf(h.win.IsMinimized()), boolOf(h.win.IsFullscreen()), boolOf(focused, nil),
		boolOf(h.win.ShouldClose())))
}

func (h *host) showFrame() {
	state, _ := h.win.FrameState()
	r, ok, _ := h.win.FrameRadii()
	radii := "none"
	if ok {
		radii = fmt.Sprintf("%.0f/%.0f/%.0f/%.0f", r.TopLeft, r.TopRight, r.BottomRight, r.BottomLeft)
	}
	h.setText("frame-out", fmt.Sprintf("radii=%s frameless=%d shadow=%d sizeable=%d", radii,
		boolOf(state.Frameless, nil), boolOf(state.Shadow, nil), boolOf(state.Sizeable, nil)))
}

func (h *host) report(id string, err error, success, failure string) {
	if err != nil {
		h.setText(id, failure+": "+err.Error())
		return
	}
	h.setText(id, success)
}

func (h *host) buttons(ctx *runner.Context) error {
	actions := map[string]func(){
		"btn-geom-read": h.showGeometry,
		"btn-geom-shrink": func() {
			_ = h.win.Resize(h.orig.w-60, h.orig.h-40)
			h.showGeometry()
		},
		"btn-geom-restore": func() {
			_ = h.win.Resize(h.orig.w, h.orig.h)
			h.showGeometry()
		},
		"btn-title": func() {
			h.titleToggled = !h.titleToggled
			if h.titleToggled {
				_ = h.win.SetTitle("Title set from a button")
			} else {
				_ = h.win.SetTitle(h.origTitle)
			}
			h.showGeometry()
		},
		"btn-state": func() { h.showState("") },
		"btn-max-cycle": func() {
			_ = h.win.Maximize()
			prefix := fmt.Sprintf("while maximized: %d | after restore: ", boolOf(h.win.IsMaximized()))
			_ = h.win.Restore()
			h.showState(prefix)
		},
		"btn-fs-cycle": func() {
			_ = h.win.SetFullscreen(true)
			prefix := fmt.Sprintf("while fullscreen: %d | after exit: ", boolOf(h.win.IsFullscreen()))
			_ = h.win.SetFullscreen(false)
			h.showState(prefix)
		},
		// Manual: a minimized window stops rendering until it is restored from the taskbar.
		"btn-minimize": func() { _ = h.win.Minimize() },
		"btn-radius": func() {
			_ = h.win.SetFrameRadius(12)
			h.showFrame()
		},
		"btn-radii": func() {
			_ = h.win.SetFrameRadii(kui.FrameRadii{TopLeft: 16, TopRight: 4, BottomRight: 16, BottomLeft: 4})
			h.showFrame()
		},
		"btn-radius-clear": func() {
			_ = h.win.ClearFrameRadius()
			h.showFrame()
		},
		"btn-frameless": func() {
			state, _ := h.win.FrameState()
			_ = h.win.SetFrameless(!state.Frameless)
			_ = h.win.SetFramelessShadow(!state.Frameless)
			h.showFrame()
		},
		"btn-opacity": func() { h.report("look-out", h.win.SetOpacity(0.85), "opacity 0.85", "set_opacity failed") },
		"btn-opaque":  func() { h.report("look-out", h.win.SetOpacity(1), "opacity 1.00", "set_opacity failed") },
		"btn-top":     func() { h.report("look-out", h.win.SetAlwaysOnTop(true), "always on top", "set_always_on_top failed") },
		"btn-untop": func() {
			h.report("look-out", h.win.SetAlwaysOnTop(false), "normal z-order", "set_always_on_top failed")
		},
		"btn-cursor": func() {
			h.report("look-out", h.win.SetCursor(kui.CursorWait), "cursor: wait (until the pointer moves to other content)", "set_cursor failed")
		},
		"btn-cursor-hide": func() {
			_ = h.win.SetCursorVisible(false)
			h.setText("look-out", "cursor hidden for 1 s")
			_, _ = kui.SetTimeout(func() {
				_ = h.win.SetCursorVisible(true)
				h.setText("look-out", "cursor visible again")
			}, 1000)
		},
		"btn-ime": func() {
			if input, err := h.doc.GetElementByID("ime-input"); err == nil && input != nil {
				_ = input.Focus()
				input.Close()
			}
			x, y := h.imeCaret()
			h.setText("look-out", fmt.Sprintf("IME candidates at (%.0f, %.0f): err %v", x, y, h.win.SetIMECandidatePos(x, y)))
		},
		"btn-menu-set": func() {
			h.report("menu-out", h.installMenu(), "menu installed: View > Word wrap (checked), Reload, Zoom", "set_menu_bar failed")
		},
		"btn-menu-check": func() {
			err := h.win.SetMenuItemChecked(menuWrap, !h.menuChecked)
			if err == nil {
				h.menuChecked = !h.menuChecked
			}
			h.setText("menu-out", fmt.Sprintf("Word wrap checked=%d (status %d)", boolOf(h.menuChecked, nil), statusOf(err)))
		},
		"btn-menu-enable": func() {
			err := h.win.SetMenuItemEnabled(menuReload, !h.menuEnabled)
			if err == nil {
				h.menuEnabled = !h.menuEnabled
			}
			h.setText("menu-out", fmt.Sprintf("Reload enabled=%d (status %d)", boolOf(h.menuEnabled, nil), statusOf(err)))
		},
		"btn-menu-clear": func() { h.report("menu-out", h.win.ClearMenuBar(), "menu removed", "clear_menu_bar failed") },
		"btn-sc-bind": func() {
			h.buttonHits = 0
			id, err := h.win.BindShortcut("F7", func() {
				h.buttonHits++
				h.setText("sc-out", fmt.Sprintf("F7 fired %d time(s)", h.buttonHits))
			})
			h.report("sc-out", err, fmt.Sprintf("F7 bound (id %d)", id), "bind failed")
		},
		"btn-sc-press": func() {
			before := h.buttonHits
			if !h.sendNativeKey(vkF7) {
				h.setText("sc-out", "no native key injection on this platform: press F7")
			} else if h.buttonHits == before {
				h.setText("sc-out", fmt.Sprintf("F7 sent, no shortcut ran (fired %d time(s))", h.buttonHits))
			}
		},
		"btn-sc-disable": func() {
			_ = h.win.SetShortcutsEnabled(false)
			h.setText("sc-out", "shortcuts disabled")
		},
		"btn-sc-enable": func() {
			_ = h.win.SetShortcutsEnabled(true)
			h.setText("sc-out", "shortcuts enabled")
		},
		"btn-sc-unbind": func() {
			_ = h.win.UnbindAllShortcuts()
			h.setText("sc-out", "all shortcuts unbound")
		},
		"btn-tray-click": func() {
			if !h.ensureTray(false) {
				h.setText("tray-out", "no tray on this platform")
				return
			}
			_ = h.win.SimulateTrayEvent(kui.TrayClick)
		},
		"btn-tray-right": func() {
			if !h.ensureTray(false) {
				h.setText("tray-out", "no tray on this platform")
				return
			}
			_ = h.win.SimulateTrayEvent(kui.TrayRightClick)
		},
		// Manual: shows the tray icon and opens its context menu (a system popup).
		"btn-tray-popup": func() {
			if !h.ensureTray(true) {
				h.setText("tray-out", "no tray on this platform")
				return
			}
			_ = h.win.SetTrayContextMenu([]kui.MenuItem{{Text: "Say hello", ID: 1}, {Text: "Hide icon", ID: 2}},
				func(id uint32) { h.setText("tray-out", fmt.Sprintf("tray menu item %d", id)) })
			h.report("tray-out", h.win.PopupTrayContextMenu(), "context menu shown", "no context menu on this platform")
		},
		"btn-shot": func() {
			path := filepath.Join(os.TempDir(), "kui_api_window_button_go.png")
			saved, _ := h.win.SaveScreenshot(path)
			bytes := "0"
			if pngSize(path) > 100 {
				bytes = ">100"
			}
			_ = os.Remove(path)
			h.setText("dt-out", fmt.Sprintf("saved=%d PNG bytes=%s (%s, removed again)", boolOf(saved, nil), bytes, path))
		},
		"btn-dt-state": func() {
			running, port, clients, _ := h.win.DevToolsState()
			h.setText("dt-out", fmt.Sprintf("devtools running=%d port=%d clients=%d", boolOf(running, nil), port, clients))
		},
		// Manual: stopping the server disconnects every inspector (and the automated check).
		"btn-dt-stop": func() {
			_ = h.win.StopDevTools()
			h.setText("dt-out", "devtools stopped")
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
	h := &host{win: ctx.Window, doc: ctx.Document, script: ctx.Script}
	t := &tour{doc: ctx.Document}
	h.origTitle = h.title()
	h.orig = h.geometry()
	h.origFrame, _ = h.win.FrameState()

	h.checkGeometry(t)
	h.checkState(t)
	h.checkFrame(t)
	h.checkLook(t)
	h.checkMenu(t)
	h.checkShortcuts(t)
	h.checkTray(t)
	h.checkDevToolsScreenshot(t)
	h.checkHost(t)
	h.checkRestored(t)
	t.finish()
	return h.buttons(ctx)
}
