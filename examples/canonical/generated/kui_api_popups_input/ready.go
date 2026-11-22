// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_popups_input.go; edit that file, not this copy.
// API tour: popups, toasts and input pickers (examples/api_tours/popups_input). At start-up the
// host shows and dismisses toasts, tooltips and context menus, reads the popup layer's state
// back after each step, configures the tooltip theme / delay and the default context menu, and
// drives the <input> picker panel (open, page, time, colour, commit, close, enable) checking the
// picker state and the input's value (read with the script engine). #tour-result ends as
// "PASS n/n"; the panel buttons repeat the calls with a visible effect.
package main

import (
	"errors"
	"fmt"
	"math"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------
// check:    a documented behaviour; a failure makes the tour FAIL.
// knownBug: a documented behaviour the engine currently gets wrong (reported bug). A failure is
//           logged as "KNOWN BUG" and kept out of the verdict; once it passes it counts as ok.

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

// ---- Small helpers ----------------------------------------------------------------------

func near(a, b, tolerance float64) bool { return math.Abs(a-b) <= tolerance }

// status is the KuiStatus behind an error (0 for nil).
func status(err error) kui.Status {
	var kerr *kui.Error
	if errors.As(err, &kerr) {
		return kerr.Status
	}
	if err != nil {
		return -1
	}
	return 0
}

func statusName(err error) string {
	switch status(err) {
	case 0:
		return "OK"
	case kui.ErrInvalidArg:
		return "INVALID_ARG"
	case kui.ErrNotFound:
		return "NOT_FOUND"
	case kui.ErrNotSupported:
		return "NOT_SUPPORTED"
	}
	return "ERROR"
}

func modeName(mode kui.PickerMode) string {
	switch mode {
	case kui.PickerDate:
		return "date"
	case kui.PickerTime:
		return "time"
	case kui.PickerDateTimeLocal:
		return "datetime-local"
	case kui.PickerColor:
		return "color"
	}
	return "none"
}

func b2i(value bool) int {
	if value {
		return 1
	}
	return 0
}

type host struct {
	doc    *kui.Document
	t      *tour
	pop    *kui.Popups
	picker *kui.InputPicker

	toastBase   int
	lastToast   uint32
	toastSerial int

	defaultTheme kui.TooltipTheme
	purpleOn     bool
	delayStep    int
	tipID        uint32

	menuID     uint32
	contextOwn bool

	timeChanges, timeInputs int
}

var purple = kui.TooltipTheme{Background: 0x7c3aedff, Foreground: 0xffffffff, Border: 0x5b21b6ff, Radius: 8}

func (h *host) rectOf(id string) kui.Rect {
	element, err := h.doc.GetElementByID(id)
	if err != nil || element == nil {
		return kui.Rect{}
	}
	defer element.Close()
	rect, _ := element.BoundingRect()
	return rect
}

func (h *host) setText(id, text string) {
	if element, err := h.doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

func (h *host) hasType(kind kui.PopupType) int {
	has, err := h.pop.Has(kind)
	if err != nil {
		return -1
	}
	return b2i(has)
}

func (h *host) anyPopup() int {
	any, err := h.pop.Any()
	if err != nil {
		return -1
	}
	return b2i(any)
}

func (h *host) toastCount() int {
	count, err := h.doc.ToastCount()
	if err != nil {
		return -1
	}
	return count
}

// inputValue reads the input's current value (what the user typed): Element.Value.
func (h *host) inputValue(id string) string {
	element, err := h.doc.GetElementByID(id)
	if err != nil || element == nil {
		return ""
	}
	defer element.Close()
	value, _ := element.Value()
	return value
}

// ---- Toasts: show two, dismiss one by id, then all; the count is read back --------------

func (h *host) showToast(message string, level kui.ToastLevel) uint32 {
	id, _ := h.doc.ShowToast(kui.ToastOptions{Message: message, Title: "KUI API tour", Level: level,
		DurationMS: 0, Closable: true})
	return id
}

func (h *host) checkToasts() {
	h.toastBase = h.toastCount()
	first := h.showToast("First toast (stays until dismissed)", kui.ToastInfo)
	second := h.showToast("Second toast", kui.ToastSuccess)
	after := h.toastCount()
	h.t.check("toast.get_toast_count", h.toastBase >= 0 && after == h.toastBase+2 && first != 0 && second != 0 && first != second,
		fmt.Sprintf("count %d -> %d, ids %d %d", h.toastBase, after, first, second))
	err := h.doc.DismissToast(first)
	h.t.check("toast.dismiss_toast", err == nil, fmt.Sprint(err))
	err = h.doc.DismissAllToasts()
	h.t.check("toast.dismiss_all_toasts", err == nil, fmt.Sprint(err))
	// Dismissed toasts play an exit animation and still count; the final phase checks 0.
}

// ---- Tooltip: delay and theme read back, a tooltip shown at an element's rect -----------

func (h *host) checkTooltip() {
	delay, _ := h.pop.TooltipDelay()
	_ = h.pop.SetTooltipDelay(150)
	after, _ := h.pop.TooltipDelay()
	bad := h.pop.SetTooltipDelay(-5)
	h.t.check("tooltip.delay", near(delay, 500, 0.01) && near(after, 150, 0.01) && status(bad) == kui.ErrInvalidArg,
		fmt.Sprintf("default %.1f, after set(150) %.1f, set(-5) %v", delay, after, bad))
	_ = h.pop.SetTooltipDelay(500)

	h.defaultTheme, _ = h.pop.TooltipTheme()
	_ = h.pop.SetTooltipTheme(purple)
	theme, _ := h.pop.TooltipTheme()
	broken := purple
	broken.Radius = -1
	bad = h.pop.SetTooltipTheme(broken)
	h.t.check("tooltip.theme", theme == purple && status(bad) == kui.ErrInvalidArg,
		fmt.Sprintf("read back %+v, radius -1 %v", theme, bad))
	_ = h.pop.SetTooltipTheme(h.defaultTheme)

	anchor := h.rectOf("tip-anchor")
	id, err := h.pop.ShowTooltip("Tooltip from the host", anchor, kui.PopupBottom)
	shown := h.hasType(kui.PopupTooltip)
	h.t.check("tooltip.show_tooltip", err == nil && id != 0 && shown == 1 && anchor.Width > 0,
		fmt.Sprintf("err %v id %d has tooltip %d anchor %.0fx%.0f", err, id, shown, anchor.Width, anchor.Height))
	_, bad = h.pop.ShowTooltip("x", anchor, kui.PopupPosition(7))
	h.t.check("tooltip.bad_position", status(bad) == kui.ErrInvalidArg, fmt.Sprint(bad))
	_ = h.pop.Dismiss(id)
	shown = h.hasType(kui.PopupTooltip)
	err = h.pop.Dismiss(987654)
	h.t.check("popup.dismiss_popup", shown == 0 && err == nil,
		fmt.Sprintf("has tooltip after dismiss %d, unknown id %v", shown, err))
}

// ---- Context menu: default menu switch, a host menu shown and dismissed by type ---------

var menuItems = []kui.MenuItem{
	{Text: "Red", ID: 1},
	{Text: "Green", ID: 2},
	{Separator: true},
	{Text: "Disabled", ID: 3, Disabled: true},
}

func (h *host) onMenuChoice(id uint32) {
	name := "?"
	switch id {
	case 1:
		name = "Red"
	case 2:
		name = "Green"
	}
	h.setText("menu-out", fmt.Sprintf("chose item %d (%s)", id, name))
}

// showMenu opens the host menu at #menu-spot's top-left corner.
func (h *host) showMenu() (uint32, error) {
	spot := h.rectOf("menu-spot")
	return h.pop.ShowContextMenu(spot.X, spot.Y, menuItems, h.onMenuChoice)
}

func (h *host) checkMenu() {
	initial, _ := h.pop.DefaultContextMenuEnabled()
	_ = h.pop.SetDefaultContextMenuEnabled(true)
	on, _ := h.pop.DefaultContextMenuEnabled()
	_ = h.pop.SetDefaultContextMenuEnabled(false)
	off, _ := h.pop.DefaultContextMenuEnabled()
	h.t.check("menu.default_enabled", !initial && on && !off,
		fmt.Sprintf("initial %v, after on %v, after off %v", initial, on, off))

	id, err := h.showMenu()
	open := h.hasType(kui.PopupContextMenu)
	h.t.check("menu.show_context_menu", err == nil && id != 0 && open == 1,
		fmt.Sprintf("err %v id %d has menu %d", err, id, open))
	_, bad := h.pop.ShowContextMenu(10, 10, nil, h.onMenuChoice)
	h.t.check("menu.no_items", status(bad) == kui.ErrInvalidArg, fmt.Sprint(bad))
	_ = h.pop.DismissType(kui.PopupContextMenu)
	open = h.hasType(kui.PopupContextMenu)
	// The Go binding owns the callback handle and frees it from the menu's release callback.
	h.t.check("menu.release_fn", open == 0, fmt.Sprintf("has menu %d", open))
}

// ---- Popup layer: both kinds at once, removed by type, by id and all at once ------------

func (h *host) showBoth() {
	h.tipID, _ = h.pop.ShowTooltip("Tooltip from the host", h.rectOf("tip-anchor"), kui.PopupBottom)
	h.menuID, _ = h.showMenu()
}

func (h *host) checkPopupLayer() {
	h.showBoth()
	any, tip, menu, custom := h.anyPopup(), h.hasType(kui.PopupTooltip), h.hasType(kui.PopupContextMenu), h.hasType(kui.PopupCustom)
	h.t.check("popup.has_popups", any == 1 && tip == 1 && menu == 1 && custom == 0,
		fmt.Sprintf("any %d tooltip %d menu %d custom %d", any, tip, menu, custom))
	_ = h.pop.DismissType(kui.PopupTooltip)
	tip, menu = h.hasType(kui.PopupTooltip), h.hasType(kui.PopupContextMenu)
	h.t.check("popup.dismiss_popups_of_type", tip == 0 && menu == 1, fmt.Sprintf("tooltip %d menu %d", tip, menu))
	_ = h.pop.Dismiss(h.menuID)
	any = h.anyPopup()
	h.t.check("popup.dismiss_popup-menu", any == 0, fmt.Sprintf("any %d", any))
	h.showBoth()
	_ = h.pop.DismissAll()
	any = h.anyPopup()
	bad := h.pop.DismissType(kui.PopupType(9))
	h.t.check("popup.dismiss_all_popups", any == 0 && status(bad) == kui.ErrInvalidArg,
		fmt.Sprintf("any %d, bad type %v", any, bad))
}

// ---- Input pickers: open per input type, page / time / colour, commit and close ---------

func (h *host) pickerState() kui.InputPickerState {
	state, _ := h.picker.State()
	return state
}

// pickerTarget is the id of the picker's target input ("" when it is closed).
func (h *host) pickerTarget() (string, error) {
	target, err := h.picker.Target()
	if err != nil || target == nil {
		return "", err
	}
	defer target.Close()
	id, _ := target.Attribute("id")
	return id, nil
}

func (h *host) openPicker(id string, mode kui.PickerMode) error {
	element, err := h.doc.GetElementByID(id)
	if err != nil || element == nil {
		return errors.New("no #" + id)
	}
	defer element.Close()
	return element.OpenInputPicker(mode)
}

func (h *host) checkPickerDate() {
	state := h.pickerState()
	target, targetErr := h.pickerTarget()
	h.t.check("picker.initial_state", !state.Open && state.Enabled && target == "" && targetErr == nil,
		fmt.Sprintf("open %v enabled %v target %q (%v)", state.Open, state.Enabled, target, targetErr))

	err := h.openPicker("in-date", kui.PickerNone)
	state = h.pickerState()
	target, targetErr = h.pickerTarget()
	panel, _ := h.doc.QuerySelector("kui-picker")
	h.t.check("picker.open_date", err == nil && state.Open && state.Mode == kui.PickerDate &&
		state.VisibleYear == 2025 && state.VisibleMonth == 3 && panel != nil,
		fmt.Sprintf("err %v open %v mode %s page %d-%d panel %v", err, state.Open, modeName(state.Mode),
			state.VisibleYear, state.VisibleMonth, panel != nil))
	if panel != nil {
		panel.Close()
	}
	h.t.check("picker.get_target", targetErr == nil && target == "in-date", fmt.Sprintf("target %q (%v)", target, targetErr))

	_ = h.picker.SetMonth(2026, 7)
	state = h.pickerState()
	bad := h.picker.SetMonth(2026, 13)
	h.t.check("picker.set_month", state.VisibleYear == 2026 && state.VisibleMonth == 7 && status(bad) == kui.ErrInvalidArg,
		fmt.Sprintf("page %d-%d, month 13 %v", state.VisibleYear, state.VisibleMonth, bad))

	_ = h.picker.Close()
	state = h.pickerState()
	target, targetErr = h.pickerTarget()
	value := h.inputValue("in-date")
	h.t.check("picker.close", !state.Open && target == "" && targetErr == nil && value == "2025-03-21",
		fmt.Sprintf("open %v target %q value %q", state.Open, target, value))
}

func (h *host) checkPickerTime() {
	err := h.openPicker("in-time", kui.PickerTime)
	state := h.pickerState()
	h.t.check("picker.open_time", err == nil && state.Open && state.Mode == kui.PickerTime && state.Hour == 14 && state.Minute == 30,
		fmt.Sprintf("err %v open %v mode %s time %d:%d", err, state.Open, modeName(state.Mode), state.Hour, state.Minute))
	_ = h.picker.SetTime(9, 5)
	after := h.pickerState()
	bad := h.picker.SetTime(24, 0)
	h.t.check("picker.set_time", after.Hour == 9 && after.Minute == 5 && status(bad) == kui.ErrInvalidArg,
		fmt.Sprintf("time %d:%d, 24:00 %v", after.Hour, after.Minute, bad))
	_ = h.picker.Commit()
	state = h.pickerState()
	value := h.inputValue("in-time")
	h.t.check("picker.commit_time", !state.Open && value == "09:05", fmt.Sprintf("open %v value %q", state.Open, value))
	h.t.check("picker.commit_time-events", h.timeInputs > 0 && h.timeChanges > 0,
		fmt.Sprintf("input %d change %d", h.timeInputs, h.timeChanges))
}

func (h *host) checkPickerDatetime() {
	err := h.openPicker("in-dtl", kui.PickerNone)
	state := h.pickerState()
	h.t.check("picker.open_datetime_local", err == nil && state.Mode == kui.PickerDateTimeLocal &&
		state.VisibleYear == 2025 && state.VisibleMonth == 3,
		fmt.Sprintf("err %v mode %s page %d-%d", err, modeName(state.Mode), state.VisibleYear, state.VisibleMonth))
	h.t.knownBug("picker.open_datetime_local-time", state.Hour == 14 && state.Minute == 30,
		"the datetime-local panel ignores the input's time and opens at the current time",
		fmt.Sprintf("time %d:%d, value 2025-03-21T14:30", state.Hour, state.Minute))
	_ = h.picker.SetMonth(2026, 7)
	_ = h.picker.SetTime(8, 15)
	_ = h.picker.Commit()
	value := h.inputValue("in-dtl")
	h.t.check("picker.commit_datetime_local", value == "2026-07-21T08:15", fmt.Sprintf("value %q", value))
}

func (h *host) checkPickerColor() {
	err := h.openPicker("in-color", kui.PickerColor)
	state := h.pickerState()
	// #3b82f6 = rgb(59,130,246): h 217.2, s 0.760, v 0.965.
	h.t.check("picker.open_color", err == nil && state.Mode == kui.PickerColor && near(float64(state.ColorH), 217.2, 1) &&
		near(float64(state.ColorS), 0.760, 0.01) && near(float64(state.ColorV), 0.965, 0.01),
		fmt.Sprintf("err %v mode %s hsv %.1f %.3f %.3f", err, modeName(state.Mode), state.ColorH, state.ColorS, state.ColorV))
	_ = h.picker.SetColorHSV(0, 1, 1)
	after := h.pickerState()
	bad := h.picker.SetColorHSV(400, 1, 1)
	h.t.check("picker.set_color_hsv", near(float64(after.ColorH), 0, 0.01) && near(float64(after.ColorS), 1, 0.01) &&
		near(float64(after.ColorV), 1, 0.01) && status(bad) == kui.ErrInvalidArg,
		fmt.Sprintf("hsv %.1f %.2f %.2f, h 400 %v", after.ColorH, after.ColorS, after.ColorV, bad))
	_ = h.picker.Commit()
	value := h.inputValue("in-color")
	h.t.check("picker.commit_color", value == "#ff0000", fmt.Sprintf("value %q", value))
}

func (h *host) checkPickerRefusals() {
	textErr := h.openPicker("in-text", kui.PickerNone)
	divErr := h.openPicker("not-input", kui.PickerNone)
	h.t.check("picker.no_panel", status(textErr) == kui.ErrNotSupported && status(divErr) == kui.ErrInvalidArg,
		fmt.Sprintf("text input %v, div %v", textErr, divErr))
	_ = h.picker.SetEnabled(false)
	off := h.pickerState()
	disabledErr := h.openPicker("in-date", kui.PickerNone)
	_ = h.picker.SetEnabled(true)
	on := h.pickerState()
	h.t.check("picker.set_enabled", !off.Enabled && status(disabledErr) == kui.ErrNotSupported && !off.Open && on.Enabled,
		fmt.Sprintf("disabled %v open %v (%v), re-enabled %v", !off.Enabled, off.Open, disabledErr, on.Enabled))
}

// ---- Panel buttons ----------------------------------------------------------------------

func (h *host) reportToasts(what string) {
	h.setText("toast-out", fmt.Sprintf("%s count=%d", what, h.toastCount()))
}

func (h *host) reportTooltip() {
	theme, _ := h.pop.TooltipTheme()
	delay, _ := h.pop.TooltipDelay()
	name := "default"
	if h.purpleOn {
		name = "purple"
	}
	h.setText("tip-out", fmt.Sprintf("theme=%s bg=#%06x delay=%.0f tooltip=%d", name, theme.Background>>8, delay,
		h.hasType(kui.PopupTooltip)))
}

func (h *host) showTip() {
	h.tipID, _ = h.pop.ShowTooltip("Tooltip from the host", h.rectOf("tip-anchor"), kui.PopupBottom)
	h.reportTooltip()
}

// A click anywhere dismisses open popups first, so each popup-layer button shows both kinds
// again and then applies its own removal; the report is "before -> after".
func (h *host) popupCounts() string {
	return fmt.Sprintf("popups=%d tooltip=%d menu=%d", h.anyPopup(), h.hasType(kui.PopupTooltip), h.hasType(kui.PopupContextMenu))
}

func (h *host) popupStep(remove func()) {
	h.showBoth()
	before := h.popupCounts()
	remove()
	h.setText("pop-out", before+" -> "+h.popupCounts())
}

func (h *host) reportPicker(what string) {
	state := h.pickerState()
	target, _ := h.pickerTarget()
	if target == "" {
		target = "-"
	}
	h.setText("pick-out", fmt.Sprintf("%sopen=%d enabled=%d mode=%s page=%04d-%02d target=%s", what, b2i(state.Open),
		b2i(state.Enabled), modeName(state.Mode), state.VisibleYear, state.VisibleMonth, target))
}

func (h *host) reportValues() {
	h.setText("pick-values", fmt.Sprintf("time=%s dtl=%s color=%s", h.inputValue("in-time"), h.inputValue("in-dtl"),
		h.inputValue("in-color")))
}

func (h *host) onContextMenu(event *kui.Event) {
	enabled, _ := h.pop.DefaultContextMenuEnabled()
	h.contextOwn = !enabled
	if !enabled {
		// Right-click: the built-in menu when it is enabled, the host's own menu otherwise.
		_ = event.PreventDefault()
		h.menuID, _ = h.showMenu()
	}
	_, _ = kui.SetTimeout(func() {
		kind := "default"
		if h.contextOwn {
			kind = "custom"
		}
		h.setText("menu-out", fmt.Sprintf("contextmenu: %s menu open=%d", kind, h.hasType(kui.PopupContextMenu)))
	}, 100)
}

func (h *host) buttons() map[string]func() {
	return map[string]func(){
		"btn-toast-show": func() {
			h.toastSerial++
			h.lastToast = h.showToast(fmt.Sprintf("Toast #%d from the host", h.toastSerial), kui.ToastLevel(h.toastSerial%4))
			h.reportToasts(fmt.Sprintf("shown #%d", h.toastSerial))
		},
		"btn-toast-dismiss": func() { _ = h.doc.DismissToast(h.lastToast); h.reportToasts("dismissed last") },
		"btn-toast-clear":   func() { _ = h.doc.DismissAllToasts(); h.reportToasts("dismissed all") },
		"btn-toast-count":   func() { h.reportToasts("now") },
		"btn-tip-show":      func() { h.showTip() },
		"btn-tip-hide":      func() { _ = h.pop.Dismiss(h.tipID); h.reportTooltip() },
		"btn-tip-theme": func() {
			h.purpleOn = !h.purpleOn
			if h.purpleOn {
				_ = h.pop.SetTooltipTheme(purple)
			} else {
				_ = h.pop.SetTooltipTheme(h.defaultTheme)
			}
			h.showTip() // the click closed the tooltip: show it again in the new colours
		},
		"btn-tip-delay": func() {
			delays := []float64{500, 150, 1000}
			h.delayStep = (h.delayStep + 1) % 3
			_ = h.pop.SetTooltipDelay(delays[h.delayStep])
			h.reportTooltip()
		},
		"btn-menu-show": func() {
			h.menuID, _ = h.showMenu()
			h.setText("menu-out", fmt.Sprintf("menu open=%d", h.hasType(kui.PopupContextMenu)))
		},
		"btn-menu-default": func() {
			enabled, _ := h.pop.DefaultContextMenuEnabled()
			_ = h.pop.SetDefaultContextMenuEnabled(!enabled)
			enabled, _ = h.pop.DefaultContextMenuEnabled()
			if enabled {
				h.setText("menu-out", "default=on")
			} else {
				h.setText("menu-out", "default=off")
			}
		},
		"btn-pop-both": func() { h.showBoth(); h.setText("pop-out", h.popupCounts()) },
		"btn-pop-tip-off": func() {
			h.popupStep(func() { _ = h.pop.DismissType(kui.PopupTooltip) })
		},
		"btn-pop-menu-id": func() { h.popupStep(func() { _ = h.pop.Dismiss(h.menuID) }) },
		"btn-pop-clear":   func() { h.popupStep(func() { _ = h.pop.DismissAll() }) },
		"btn-pick-date": func() {
			err := h.openPicker("in-date", kui.PickerNone)
			h.reportPicker("status=" + statusName(err) + " ")
		},
		"btn-pick-close": func() { _ = h.picker.Close(); h.reportPicker("") },
		"btn-pick-time": func() {
			if h.openPicker("in-time", kui.PickerTime) == nil {
				_ = h.picker.SetTime(18, 45)
				_ = h.picker.Commit()
			}
			h.reportPicker("")
			h.reportValues()
		},
		"btn-pick-dtl": func() {
			if h.openPicker("in-dtl", kui.PickerDateTimeLocal) == nil {
				_ = h.picker.SetMonth(2027, 1)
				_ = h.picker.SetTime(22, 10)
				_ = h.picker.Commit()
			}
			h.reportPicker("")
			h.reportValues()
		},
		"btn-pick-color": func() {
			if h.openPicker("in-color", kui.PickerColor) == nil {
				_ = h.picker.SetColorHSV(120, 1, 1)
				_ = h.picker.Commit()
			}
			h.reportPicker("")
			h.reportValues()
		},
		"btn-pick-toggle": func() {
			state := h.pickerState()
			_ = h.picker.SetEnabled(!state.Enabled)
			err := h.openPicker("in-date", kui.PickerNone)
			if err == nil {
				_ = h.picker.Close()
			}
			h.reportPicker("open(date) status=" + statusName(err) + " ")
		},
	}
}

func (h *host) listen(ctx *runner.Context, id, event string, fn func(*kui.Event)) error {
	element, err := h.doc.GetElementByID(id)
	if err != nil || element == nil {
		return fmt.Errorf("page has no #%s: %v", id, err)
	}
	ctx.Keep(element)
	_, err = element.AddEventListener(event, fn)
	return err
}

func ready(ctx *runner.Context) error {
	h := &host{doc: ctx.Document, t: &tour{doc: ctx.Document}}
	h.pop = ctx.Document.Popups()
	h.picker = ctx.Document.InputPicker()

	if err := h.listen(ctx, "in-time", "change", func(*kui.Event) { h.timeChanges++ }); err != nil {
		return err
	}
	if err := h.listen(ctx, "in-time", "input", func(*kui.Event) { h.timeInputs++ }); err != nil {
		return err
	}

	h.checkToasts()
	h.checkTooltip()
	h.checkMenu()
	h.checkPopupLayer()
	h.checkPickerDate()
	h.checkPickerTime()
	h.checkPickerDatetime()
	h.checkPickerColor()
	h.checkPickerRefusals()
	h.reportValues()
	// Toast exit animations take a few hundred milliseconds.
	if _, err := kui.SetTimeout(func() {
		count := h.toastCount()
		h.t.check("toast.count_after_dismiss", count == h.toastBase, fmt.Sprintf("count %d (base %d)", count, h.toastBase))
		h.t.finish()
	}, 1200); err != nil {
		return err
	}

	for id, fn := range h.buttons() {
		fn := fn
		if err := h.listen(ctx, id, "click", func(*kui.Event) { fn() }); err != nil {
			return err
		}
	}
	return h.listen(ctx, "menu-spot", "contextmenu", h.onContextMenu)
}
