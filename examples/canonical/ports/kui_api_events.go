//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_events.go; edit that file, not this copy.
// API tour: DOM events from the host (examples/api_tours/events). On start the host registers
// listeners, dispatches synthetic events (Element.DispatchEvent) and injects real mouse and
// keyboard input (Document.DispatchMouse / DispatchKey, the same dispatch path as the OS), then
// checks what every listener saw: propagation order, current target, StopPropagation, flags and
// PreventDefault, listener removal, mouse button and coordinates, related targets, key / code /
// modifiers, focus events and the whole drag-and-drop data API. One row per check goes into
// #tour-log; #tour-result ends as "PASS n/n". Afterwards the panels react to real input: every
// listener writes what it read into the panel's output line.
//
// Binding note: Event.Flags and Event.Modifiers are the direct wrappers of kui_event_get_flags /
// kui_event_key_get_modifiers (FlagValues / ModifierKeys return the same values as structs).
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

// ---- Small helpers ------------------------------------------------------------------------

var (
	doc *kui.Document
	ctx *runner.Context
)

func with(id string, fn func(*kui.Element)) {
	if element, err := doc.GetElementByID(id); err == nil && element != nil {
		fn(element)
		element.Close()
	}
}

func near(a, b float32) bool { return math.Abs(float64(a-b)) <= 0.51 }

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

func idOf(e *kui.Element, err error) string {
	if err != nil || e == nil {
		return ""
	}
	defer e.Close()
	id, _ := e.ID()
	return id
}

func setText(id, text string)   { with(id, func(e *kui.Element) { _ = e.SetTextContent(text) }) }
func setClass(id, class string) { with(id, func(e *kui.Element) { _ = e.SetClassName(class) }) }

func rectOf(id string) (r kui.Rect) {
	with(id, func(e *kui.Element) { r, _ = e.BoundingRect() })
	return
}

func dispatch(id, name string, bubbles, cancelable bool) (notCanceled bool, err error) {
	err = fmt.Errorf("no #%s", id)
	with(id, func(e *kui.Element) { notCanceled, err = e.DispatchEvent(name, bubbles, cancelable) })
	return
}

// listen registers fn on the element with that id; the element is kept so the listener lives.
func listen(id, name string, fn func(*kui.Event)) error {
	element, err := doc.GetElementByID(id)
	if err != nil || element == nil {
		return fmt.Errorf("没有 #%s: %v", id, err)
	}
	ctx.Keep(element)
	_, err = element.AddEventListener(name, fn)
	return err
}

// Injected real input, in viewport coordinates.
func mouse(kind kui.MouseEventType, x, y float32, button kui.MouseButton, clicks int) {
	_ = doc.DispatchMouse(kind, x, y, button, 0, clicks, 0, 0)
}

func clickAt(id string, button kui.MouseButton, clicks int) {
	r := rectOf(id)
	x, y := r.X+r.Width/2, r.Y+r.Height/2
	mouse(kui.MouseMoved, x, y, kui.MouseButtonNone, 0)
	mouse(kui.MousePressed, x, y, button, clicks)
	mouse(kui.MouseReleased, x, y, button, clicks)
}

func moveTo(id string) {
	r := rectOf(id)
	mouse(kui.MouseMoved, r.X+r.Width/2, r.Y+r.Height/2, kui.MouseButtonNone, 0)
}

func focusID(id string) { with(id, func(e *kui.Element) { _ = e.Focus() }) }
func blurID(id string)  { with(id, func(e *kui.Element) { _ = e.Blur() }) }

// ---- Propagation: one listener per layer records where the event is ----------------------

var (
	trace, targets                             []string // current targets / targets, in call order
	lastBubbles, lastCancelable, lastPrevented int      // flags read by the last listener that ran
	layers                                     = []string{"ev-outer", "ev-middle", "ev-inner"}
)

func boolInt(v bool) int { return map[bool]int{false: 0, true: 1}[v] }

func onLayer(ev *kui.Event) {
	current := strings.TrimPrefix(idOf(ev.CurrentTarget()), "ev-")
	trace = append(trace, current)
	targets = append(targets, strings.TrimPrefix(idOf(ev.Target()), "ev-"))
	if bubbles, cancelable, prevented, err := ev.Flags(); err == nil {
		lastBubbles, lastCancelable, lastPrevented = boolInt(bubbles), boolInt(cancelable), boolInt(prevented)
	}
	if element, err := ev.CurrentTarget(); err == nil && element != nil {
		_ = element.AddClass("hit")
		element.Close()
	}
}

func fireLayers(id, name string, bubbles, cancelable bool) int {
	trace, targets = nil, nil
	lastBubbles, lastCancelable, lastPrevented = -1, -1, -1
	for _, layer := range layers {
		with(layer, func(e *kui.Element) { _ = e.RemoveClass("hit") })
	}
	notCanceled, err := dispatch(id, name, bubbles, cancelable)
	if err != nil {
		return -1
	}
	return boolInt(notCanceled)
}

func bindLayers() error {
	for _, layer := range layers {
		for _, name := range []string{"ping", "halt", "veto"} {
			if err := listen(layer, name, onLayer); err != nil {
				return err
			}
		}
	}
	// "halt": the middle layer stops the event after its own listeners.
	if err := listen("ev-middle", "halt", func(ev *kui.Event) { _ = ev.StopPropagation() }); err != nil {
		return err
	}
	// "veto": the inner layer cancels the default action.
	return listen("ev-inner", "veto", func(ev *kui.Event) { _ = ev.PreventDefault() })
}

func checkPropagation(t *tour) {
	fireLayers("ev-inner", "ping", true, false)
	t.check("prop.bubble_order", strings.Join(trace, ",") == "inner,middle,outer", fmt.Sprintf("current targets '%s'", strings.Join(trace, ",")))
	t.check("prop.target", strings.Join(targets, ",") == "inner,inner,inner", fmt.Sprintf("targets '%s'", strings.Join(targets, ",")))

	fireLayers("ev-inner", "ping", false, false)
	t.check("prop.non_bubbling", strings.Join(trace, ",") == "inner" && lastBubbles == 0,
		fmt.Sprintf("current targets '%s', bubbles %d", strings.Join(trace, ","), lastBubbles))

	fireLayers("ev-middle", "ping", true, false)
	t.check("prop.from_middle", strings.Join(trace, ",") == "middle,outer" && strings.Join(targets, ",") == "middle,middle",
		fmt.Sprintf("current targets '%s', targets '%s'", strings.Join(trace, ","), strings.Join(targets, ",")))

	fireLayers("ev-inner", "halt", true, false)
	t.check("prop.stop_propagation", strings.Join(trace, ",") == "inner,middle", fmt.Sprintf("current targets '%s'", strings.Join(trace, ",")))

	notCanceled := fireLayers("ev-inner", "veto", true, true)
	t.check("cancel.prevent_default", notCanceled == 0 && lastPrevented == 1 && lastCancelable == 1 && lastBubbles == 1 &&
		strings.Join(trace, ",") == "inner,middle,outer",
		fmt.Sprintf("not canceled %d, flags at outer b%d c%d p%d, trace '%s'", notCanceled, lastBubbles, lastCancelable, lastPrevented, strings.Join(trace, ",")))

	// PreventDefault on a non-cancelable event changes nothing.
	notCanceled = fireLayers("ev-inner", "veto", true, false)
	t.check("cancel.not_cancelable", notCanceled == 1 && lastPrevented == 0 && lastCancelable == 0,
		fmt.Sprintf("not canceled %d, flags at outer c%d p%d", notCanceled, lastCancelable, lastPrevented))

	// A detached element cannot be dispatched at (kui_c.h).
	var err error = errors.New("no element")
	if loose, createErr := doc.CreateElement("div"); createErr == nil && loose != nil {
		_, err = loose.DispatchEvent("ping", true, false)
		loose.Close()
	}
	t.knownBug("dispatch.detached", status(err) == kui.ErrInvalidArg,
		"dispatch_event on an element that is not in the document returns KUI_OK", fmt.Sprintf("status %d", status(err)))
}

// ---- Listener management: ids, removing one, removing all of one name -------------------

var (
	tickTarget *kui.Element // one wrapper owns every "tick" listener, so ids can be removed
	tickTrace  []string
	tickIDs    []uint64
	tickSerial int
)

func addTick() uint64 {
	tickSerial++
	name := fmt.Sprint(tickSerial)
	id, _ := tickTarget.AddEventListener("tick", func(*kui.Event) { tickTrace = append(tickTrace, name) })
	return id
}

func fireTick(name string) {
	tickTrace = nil
	_, _ = tickTarget.DispatchEvent(name, false, false)
}

func checkListeners(t *tour) {
	first, second, third := addTick(), addTick(), addTick()
	fireTick("tick")
	t.check("listen.registration_order", strings.Join(tickTrace, ",") == "1,2,3" && first != 0 && first != second && second != third,
		fmt.Sprintf("ran '%s', ids %d %d %d", strings.Join(tickTrace, ","), first, second, third))

	err := tickTarget.RemoveEventListener("tick", second)
	fireTick("tick")
	t.check("listen.remove_event_listener", err == nil && strings.Join(tickTrace, ",") == "1,3",
		fmt.Sprintf("error %v, ran '%s'", err, strings.Join(tickTrace, ",")))

	_, _ = tickTarget.AddEventListener("tock", func(*kui.Event) { tickTrace = append(tickTrace, "tock") })
	err = tickTarget.RemoveEventListeners("tick")
	fireTick("tick")
	t.check("listen.remove_event_listeners", err == nil && len(tickTrace) == 0,
		fmt.Sprintf("error %v, ran '%s'", err, strings.Join(tickTrace, ",")))
	fireTick("tock")
	t.check("listen.remove_keeps_other_names", strings.Join(tickTrace, ",") == "tock", fmt.Sprintf("ran '%s'", strings.Join(tickTrace, ",")))
	_ = tickTarget.RemoveEventListeners("tock")
	tickSerial = 0
}

// ---- Event object: a synthetic event and the accessors of the other kinds ---------------

type probeResult struct {
	typ                                   string
	tag                                   kui.EventTypeTag
	bubbles, cancelable, prevented        bool
	related                               *kui.Element
	relatedErr, mouseXY, mouseButton      error
	key, code, modifiers, dragTypes, data error
	files                                 error
}

var probe probeResult

func onProbe(ev *kui.Event) {
	probe = probeResult{tag: -1}
	probe.typ, _ = ev.Type()
	probe.tag, _ = ev.TypeTag()
	probe.bubbles, probe.cancelable, probe.prevented, _ = ev.Flags()
	related, err := ev.RelatedTarget()
	probe.related, probe.relatedErr = nil, err
	if related != nil {
		probe.related = related
		related.Close()
	}
	_, probe.mouseXY = ev.MouseClient()
	_, probe.mouseButton = ev.MouseButton()
	_, probe.key = ev.Key()
	_, probe.code = ev.Code()
	_, _, _, _, probe.modifiers = ev.Modifiers()
	_, probe.dragTypes = ev.DragTypes()
	_, _, probe.data = ev.DragData("text/plain")
	_, probe.files = ev.DragFiles()
}

func checkProbe(t *tour) {
	p := &probe
	_, _ = dispatch("probe", "probe", true, false)
	t.check("event.type", p.typ == "probe" && p.tag == kui.EventBase, fmt.Sprintf("type '%s', tag %d", p.typ, p.tag))
	t.check("event.flags", p.bubbles && !p.cancelable && !p.prevented,
		fmt.Sprintf("bubbles %v cancelable %v prevented %v", p.bubbles, p.cancelable, p.prevented))
	t.check("event.no_related_target", p.related == nil && p.relatedErr == nil, fmt.Sprintf("related %v, error %v", p.related != nil, p.relatedErr))
	t.check("event.mouse_not_supported", status(p.mouseXY) == kui.ErrNotSupported && status(p.mouseButton) == kui.ErrNotSupported,
		fmt.Sprintf("client %d, button %d", status(p.mouseXY), status(p.mouseButton)))
	t.check("event.key_not_supported", status(p.key) == kui.ErrNotSupported && status(p.code) == kui.ErrNotSupported &&
		status(p.modifiers) == kui.ErrNotSupported, fmt.Sprintf("key %d, code %d, modifiers %d", status(p.key), status(p.code), status(p.modifiers)))
	t.check("event.drag_not_supported", status(p.dragTypes) == kui.ErrNotSupported && status(p.data) == kui.ErrNotSupported &&
		status(p.files) == kui.ErrNotSupported, fmt.Sprintf("types %d, data %d, files %d", status(p.dragTypes), status(p.data), status(p.files)))
}

// ---- Mouse: button, three coordinate spaces, related targets ----------------------------

type mouseRecord struct {
	seen                           int
	button                         kui.MouseButton
	tag                            kui.EventTypeTag
	client, offset, page           kui.Point
	bubbles, cancelable, prevented bool
}

var click, down, dbl, ctxMenu mouseRecord

func writeMouseLine(typ string, r *mouseRecord) {
	with("mouse-last", func(out *kui.Element) {
		_ = out.SetTextContent(fmt.Sprintf("%s button=%d client=%.0f,%.0f offset=%.0f,%.0f page=%.0f,%.0f",
			typ, r.button, r.client.X, r.client.Y, r.offset.X, r.offset.Y, r.page.X, r.page.Y))
		_ = out.SetAttribute("data-type", typ)
		_ = out.SetAttribute("data-button", fmt.Sprint(int(r.button)))
		for name, value := range map[string]float32{"data-cx": r.client.X, "data-cy": r.client.Y, "data-ox": r.offset.X,
			"data-oy": r.offset.Y, "data-px": r.page.X, "data-py": r.page.Y} {
			_ = out.SetAttribute(name, fmt.Sprintf("%.2f", value))
		}
	})
}

func onMouse(r *mouseRecord) func(*kui.Event) {
	return func(ev *kui.Event) {
		typ, _ := ev.Type()
		// The pad keeps its own context menu: cancel the default one.
		if typ == "contextmenu" {
			_ = ev.PreventDefault()
		}
		r.seen++
		r.button, r.tag = -1, -1
		r.button, _ = ev.MouseButton()
		r.tag, _ = ev.TypeTag()
		r.client, _ = ev.MouseClient()
		r.offset, _ = ev.MouseOffset()
		r.page, _ = ev.MousePage()
		r.bubbles, r.cancelable, r.prevented, _ = ev.Flags()
		if typ != "mousedown" {
			writeMouseLine(typ, r)
		}
	}
}

var enterLine, leaveLine string
var enterBubbles = -1

func onHover(ev *kui.Event) {
	typ, _ := ev.Type()
	current := idOf(ev.CurrentTarget())
	related := idOf(ev.RelatedTarget())
	if related == "" {
		related = "-"
	}
	bubbles, _, _, _ := ev.Flags()
	if element, err := ev.CurrentTarget(); err == nil && element != nil {
		if typ == "mouseenter" {
			_ = element.AddClass("in")
		} else {
			_ = element.RemoveClass("in")
		}
		element.Close()
	}
	if typ == "mouseenter" {
		enterLine = fmt.Sprintf("#%s from #%s", current, related)
		enterBubbles = boolInt(bubbles)
	} else {
		leaveLine = fmt.Sprintf("#%s to #%s", current, related)
	}
	orDash := func(s string) string {
		if s == "" {
			return "-"
		}
		return s
	}
	setText("hover-out", fmt.Sprintf("enter: %s | leave: %s", orDash(enterLine), orDash(leaveLine)))
}

func bindMouse() error {
	for name, record := range map[string]*mouseRecord{"click": &click, "mousedown": &down, "dblclick": &dbl, "contextmenu": &ctxMenu} {
		if err := listen("mouse-pad", name, onMouse(record)); err != nil {
			return err
		}
	}
	for _, id := range []string{"hover-a", "hover-b"} {
		for _, name := range []string{"mouseenter", "mouseleave"} {
			if err := listen(id, name, onHover); err != nil {
				return err
			}
		}
	}
	return nil
}

// ---- Wheel: the scroll amount of an injected wheel event --------------------------------------------

func checkWheel(t *tour) {
	pad, _ := doc.GetElementByID("mouse-pad")
	if pad == nil {
		t.check("wheel.setup", false, "missing #mouse-pad")
		return
	}
	defer pad.Close()
	rect := rectOf("mouse-pad")
	x, y := rect.X+rect.Width/2, rect.Y+rect.Height/2
	var delta kui.Point = kui.Point{X: -1, Y: -1}
	seen := 0
	var deltaErr error
	wheelID, _ := pad.AddEventListener("wheel", func(e *kui.Event) {
		seen++
		delta, deltaErr = e.WheelDelta()
	})
	_ = doc.DispatchMouse(kui.MouseWheel, x, y, kui.MouseButtonNone, 0, 0, 3, -7.5)
	_ = pad.RemoveEventListener("wheel", wheelID)
	t.check("wheel.delta", seen == 1 && deltaErr == nil && near(delta.X, 3) && near(delta.Y, -7.5),
		fmt.Sprintf("seen %d, delta %.1f,%.1f, err %v", seen, delta.X, delta.Y, deltaErr))

	// Only a wheel event has a scroll amount.
	var probeErr error
	probeID, _ := pad.AddEventListener("wheel-probe", func(e *kui.Event) { _, probeErr = e.WheelDelta() })
	_, _ = pad.DispatchEvent("wheel-probe", false, false)
	_ = pad.RemoveEventListener("wheel-probe", probeID)
	var kerr *kui.Error
	t.check("wheel.delta_of_a_non_wheel_event", errors.As(probeErr, &kerr) && kerr.Status == kui.ErrNotSupported,
		fmt.Sprintf("err %v", probeErr))
}

// Capture phase: listeners registered with AddCaptureEventListener run while the event travels down.
func checkCapture(t *tour) {
	outer, _ := doc.GetElementByID("ev-outer")
	middle, _ := doc.GetElementByID("ev-middle")
	inner, _ := doc.GetElementByID("ev-inner")
	if outer == nil || middle == nil || inner == nil {
		t.check("capture.setup", false, "missing the propagation layers")
		return
	}
	defer outer.Close()
	defer middle.Close()
	defer inner.Close()
	order := ""
	tag := func(name string) func(*kui.Event) { return func(*kui.Event) { order += name + " " } }
	outerPlain, _ := outer.AddEventListener("capture-probe", tag("op"))
	middlePlain, _ := middle.AddEventListener("capture-probe", tag("mp"))
	innerPlain, _ := inner.AddEventListener("capture-probe", tag("ip"))
	innerCapture, _ := inner.AddCaptureEventListener("capture-probe", tag("ic"))
	outerCapture, _ := outer.AddCaptureEventListener("capture-probe", tag("oc"))
	_, _ = inner.DispatchEvent("capture-probe", true, false)
	t.check("capture.order", order == "oc ic ip mp op ", fmt.Sprintf("order '%s'", order))

	_ = outer.RemoveEventListener("capture-probe", outerCapture)
	order = ""
	_, _ = inner.DispatchEvent("capture-probe", true, false)
	t.check("capture.remove", order == "ic ip mp op ", fmt.Sprintf("order '%s'", order))

	_ = outer.RemoveEventListener("capture-probe", outerPlain)
	_ = middle.RemoveEventListener("capture-probe", middlePlain)
	_ = inner.RemoveEventListener("capture-probe", innerPlain)
	_ = inner.RemoveEventListener("capture-probe", innerCapture)
}

func checkMouse(t *tour) {
	pad := rectOf("mouse-pad")
	cx, cy := pad.X+pad.Width/2, pad.Y+pad.Height/2
	var scrollTop float32
	if root, err := doc.Root(); err == nil && root != nil {
		if s, err := root.ScrollMetrics(); err == nil {
			scrollTop = s.Y
		}
		root.Close()
	}

	clickAt("mouse-pad", kui.MouseButtonLeft, 1)
	t.check("mouse.click", click.seen == 1 && click.button == kui.MouseButtonLeft && click.tag == kui.EventMouse && click.bubbles && click.cancelable,
		fmt.Sprintf("seen %d, button %d, tag %d, bubbles %v cancelable %v", click.seen, click.button, click.tag, click.bubbles, click.cancelable))
	t.check("mouse.client_xy", near(click.client.X, cx) && near(click.client.Y, cy),
		fmt.Sprintf("client %.1f,%.1f, pad centre %.1f,%.1f", click.client.X, click.client.Y, cx, cy))
	// Offset coordinates are relative to the padding box: inside the 4px border.
	t.knownBug("mouse.offset_xy", near(click.offset.X, cx-pad.X-4) && near(click.offset.Y, cy-pad.Y-4),
		"mouse events never get offset coordinates (always 0,0)",
		fmt.Sprintf("offset %.1f,%.1f, want %.1f,%.1f", click.offset.X, click.offset.Y, cx-pad.X-4, cy-pad.Y-4))
	t.knownBug("mouse.page_xy", near(click.page.X, click.client.X) && near(click.page.Y, click.client.Y+scrollTop),
		"mouse events never get page coordinates (always 0,0)",
		fmt.Sprintf("page %.1f,%.1f, client %.1f,%.1f, document scroll %.1f", click.page.X, click.page.Y, click.client.X, click.client.Y, scrollTop))

	clickAt("mouse-pad", kui.MouseButtonRight, 1)
	t.check("mouse.right_button", down.button == kui.MouseButtonRight && ctxMenu.seen == 1 && ctxMenu.button == kui.MouseButtonRight && ctxMenu.prevented,
		fmt.Sprintf("mousedown button %d, contextmenu seen %d button %d prevented %v", down.button, ctxMenu.seen, ctxMenu.button, ctxMenu.prevented))

	// kui_c.h: click count 2 injects the second click of a double-click.
	clickAt("mouse-pad", kui.MouseButtonLeft, 2)
	t.knownBug("mouse.injected_click_count", dbl.seen == 1,
		"kui_devtools_dispatch_mouse_event ignores click_count (counts clicks by timing)", fmt.Sprintf("dblclick seen %d", dbl.seen))
	// Two clicks in a row on the same spot make a double-click.
	dbl.seen = 0
	clickAt("mouse-pad", kui.MouseButtonLeft, 1)
	clickAt("mouse-pad", kui.MouseButtonLeft, 2)
	t.check("mouse.dblclick", dbl.seen == 1 && dbl.button == kui.MouseButtonLeft, fmt.Sprintf("dblclick seen %d button %d", dbl.seen, dbl.button))

	moveTo("hover-a")
	moveTo("hover-b")
	t.check("mouse.related_target", leaveLine == "#hover-a to #hover-b" && enterLine == "#hover-b from #hover-a",
		fmt.Sprintf("leave '%s', enter '%s'", leaveLine, enterLine))
	// UI Events: mouseenter / mouseleave do not bubble.
	t.knownBug("mouse.enter_does_not_bubble", enterBubbles == 0, "mouseenter / mouseleave are dispatched as bubbling events",
		fmt.Sprintf("bubbles %d", enterBubbles))
	moveTo("probe")
}

// ---- Keyboard and focus ------------------------------------------------------------------

type keyRecord struct {
	seen                   bool
	key, code              string
	ctrl, shift, alt, meta bool
	tag                    kui.EventTypeTag
	mouse                  error
}

var key keyRecord
var blurLine, focusLine string
var focusTag kui.EventTypeTag = -1

func onKey(ev *kui.Event) {
	key = keyRecord{seen: true, tag: -1}
	key.key, _ = ev.Key()
	key.code, _ = ev.Code()
	key.ctrl, key.shift, key.alt, key.meta, _ = ev.Modifiers()
	key.tag, _ = ev.TypeTag()
	_, key.mouse = ev.MouseClient()
	var mods []string
	for _, m := range []struct {
		on   bool
		name string
	}{{key.ctrl, "ctrl"}, {key.shift, "shift"}, {key.alt, "alt"}, {key.meta, "meta"}} {
		if m.on {
			mods = append(mods, m.name)
		}
	}
	shown := strings.Join(mods, ",")
	if shown == "" {
		shown = "none"
	}
	setText("key-out", fmt.Sprintf("keydown key=%s code=%s mods=%s", key.key, key.code, shown))
}

func onFocusChange(ev *kui.Event) {
	typ, _ := ev.Type()
	current := idOf(ev.CurrentTarget())
	related := idOf(ev.RelatedTarget())
	if related == "" {
		related = "-"
	}
	focusTag, _ = ev.TypeTag()
	if typ == "blur" {
		blurLine = fmt.Sprintf("blur #%s to #%s", current, related)
		setText("focus-out", blurLine)
	} else {
		focusLine = fmt.Sprintf("focus #%s from #%s", current, related)
		setText("focus-out", focusLine)
	}
}

func bindKeyboard() error {
	for _, l := range []struct {
		id, name string
		fn       func(*kui.Event)
	}{{"key-pad", "keydown", onKey}, {"key-pad", "focus", onFocusChange}, {"key-pad", "blur", onFocusChange}, {"key-other", "focus", onFocusChange}} {
		if err := listen(l.id, l.name, l.fn); err != nil {
			return err
		}
	}
	return nil
}

func checkKeyboard(t *tour) {
	focusID("key-pad")
	// Ctrl + Shift, the DevTools modifier bits.
	_ = doc.DispatchKey(kui.KeyDown, "a", "KeyA", kui.InjectCtrl|kui.InjectShift, "", false)
	_ = doc.DispatchKey(kui.KeyUp, "a", "KeyA", kui.InjectCtrl|kui.InjectShift, "", false)
	t.check("key.key_code", key.seen && key.key == "a" && key.code == "KeyA", fmt.Sprintf("seen %v key '%s' code '%s'", key.seen, key.key, key.code))
	t.check("key.modifiers", key.ctrl && key.shift && !key.alt && !key.meta,
		fmt.Sprintf("ctrl %v shift %v alt %v meta %v", key.ctrl, key.shift, key.alt, key.meta))
	t.check("key.type_tag", key.tag == kui.EventKeyboard && status(key.mouse) == kui.ErrNotSupported,
		fmt.Sprintf("tag %d, mouse accessor status %d", key.tag, status(key.mouse)))

	focusID("key-other")
	t.check("focus.blur_related_target", blurLine == "blur #key-pad to #key-other", fmt.Sprintf("'%s'", blurLine))
	t.check("focus.focus_related_target", focusLine == "focus #key-other from #key-pad" && focusTag == kui.EventFocus,
		fmt.Sprintf("'%s', tag %d", focusLine, focusTag))
	blurID("key-other")
}

// ---- Drag and drop: the data API from dragstart to dragend ------------------------------

type dragRecord struct {
	seq          []string // event types, repeats folded
	startTypes   []string // after set x3 + clear x1 in dragstart
	startRead    string
	startAllowed kui.EffectAllowed
	overEffect   kui.DropEffect
	drops        int
	dropPlain    string
	dropCustom   string
	junkFound    bool
	dropTypes    []string
	dropFiles    []string
	filesErr     error
	dropAllowed  kui.EffectAllowed
	dropEffect   kui.DropEffect
	dropTag      kui.EventTypeTag
	dropAt       kui.Point
	clearedTypes []string
	endEffect    kui.DropEffect
}

var drag dragRecord

var effectNames = []string{"none", "copy", "move", "link"}
var allowedNames = []string{"none", "copy", "move", "link", "copyMove", "copyLink", "moveLink", "all", "uninitialized"}

func name(names []string, index int) string {
	if index >= 0 && index < len(names) {
		return names[index]
	}
	return "?"
}

func noteDrag(typ string) {
	if len(drag.seq) > 0 && drag.seq[len(drag.seq)-1] == typ {
		return
	}
	drag.seq = append(drag.seq, typ)
	setText("drag-seq", strings.Join(drag.seq, " > "))
}

func onDragSource(ev *kui.Event) {
	typ, _ := ev.Type()
	switch typ {
	case "dragstart":
		drag = dragRecord{}
		setClass("drop-zone", "")
		setText("drop-out", "dragging...")
		_ = ev.SetEffectAllowed(kui.EffectCopyMove)
		_ = ev.SetDragData("text/plain", "kui-chip")
		_ = ev.SetDragData("text/x-kui", "42")
		_ = ev.SetDragData("text/x-junk", "scratch")
		_ = ev.ClearDragData("text/x-junk")
		drag.startTypes, _ = ev.DragTypes()
		drag.startRead, _, _ = ev.DragData("text/plain")
		drag.startAllowed, _ = ev.EffectAllowed()
	case "dragend":
		drag.endEffect = -1
		drag.endEffect, _ = ev.DropEffect()
	}
	noteDrag(typ)
}

func onDropZone(ev *kui.Event) {
	typ, _ := ev.Type()
	noteDrag(typ)
	switch typ {
	case "dragenter", "dragover":
		// Accept the drop and ask for a copy.
		_ = ev.PreventDefault()
		_ = ev.SetDropEffect(kui.DropCopy)
		drag.overEffect, _ = ev.DropEffect()
		setClass("drop-zone", "over")
		return
	case "dragleave":
		if drag.drops == 0 {
			setClass("drop-zone", "")
		}
		return
	}
	// drop
	_ = ev.PreventDefault()
	drag.drops++
	drag.dropPlain, _, _ = ev.DragData("text/plain")
	drag.dropCustom, _, _ = ev.DragData("text/x-kui")
	_, drag.junkFound, _ = ev.DragData("text/x-junk")
	drag.dropTypes, _ = ev.DragTypes()
	drag.dropFiles, drag.filesErr = ev.DragFiles()
	drag.dropAllowed, _ = ev.EffectAllowed()
	drag.dropEffect, _ = ev.DropEffect()
	drag.dropTag, _ = ev.TypeTag()
	drag.dropAt, _ = ev.MouseClient()
	// Done with the data: clear every format.
	_ = ev.ClearAllDragData()
	drag.clearedTypes, _ = ev.DragTypes()

	files := ""
	if len(drag.dropFiles) > 0 {
		files = ": " + strings.Join(drag.dropFiles, ",")
	}
	setText("drop-out", fmt.Sprintf("dropped '%s' + '%s' | types %d | files %d%s | effect %s / allowed %s",
		drag.dropPlain, drag.dropCustom, len(drag.dropTypes), len(drag.dropFiles), files,
		name(effectNames, int(drag.dropEffect)), name(allowedNames, int(drag.dropAllowed))))
	setClass("drop-zone", "dropped")
}

func bindDrag() error {
	for _, l := range []struct{ id, name string }{{"drag-src", "dragstart"}, {"drag-src", "dragend"}} {
		if err := listen(l.id, l.name, onDragSource); err != nil {
			return err
		}
	}
	for _, name := range []string{"dragenter", "dragover", "dragleave", "drop"} {
		if err := listen("drop-zone", name, onDropZone); err != nil {
			return err
		}
	}
	return nil
}

// simulateDrag presses on the chip, moves past the drag threshold, moves onto the zone and releases.
func simulateDrag() {
	src, zone := rectOf("drag-src"), rectOf("drop-zone")
	sx, sy := src.X+src.Width/2, src.Y+src.Height/2
	zx, zy := zone.X+zone.Width/2, zone.Y+zone.Height/2
	mouse(kui.MouseMoved, sx, sy, kui.MouseButtonNone, 0)
	mouse(kui.MousePressed, sx, sy, kui.MouseButtonLeft, 1)
	mouse(kui.MouseMoved, sx+12, sy, kui.MouseButtonLeft, 0)
	mouse(kui.MouseMoved, zx, zy, kui.MouseButtonLeft, 0)
	mouse(kui.MouseReleased, zx, zy, kui.MouseButtonLeft, 1)
}

func checkDrag(t *tour) {
	zone := rectOf("drop-zone")
	d := &drag
	simulateDrag()
	seq := strings.Join(d.seq, " > ")
	t.check("drag.sequence", strings.HasPrefix(seq, "dragstart") && strings.Contains(seq, "dragenter > dragover") &&
		strings.Contains(seq, "drop") && strings.Contains(seq, "dragend"), fmt.Sprintf("sequence '%s'", seq))
	t.check("drag.set_clear_types", strings.Join(d.startTypes, ",") == "text/plain,text/x-kui" && d.startRead == "kui-chip",
		fmt.Sprintf("types '%s', read back '%s'", strings.Join(d.startTypes, ","), d.startRead))
	t.check("drag.effect_allowed", d.startAllowed == kui.EffectCopyMove && d.dropAllowed == kui.EffectCopyMove,
		fmt.Sprintf("dragstart %d, drop %d", d.startAllowed, d.dropAllowed))
	t.check("drag.drop_effect", d.overEffect == kui.DropCopy && d.dropEffect == kui.DropCopy,
		fmt.Sprintf("dragover %d, drop %d", d.overEffect, d.dropEffect))
	t.check("drag.get_data", d.drops == 1 && d.dropPlain == "kui-chip" && d.dropCustom == "42" && !d.junkFound && len(d.dropTypes) == 2,
		fmt.Sprintf("drops %d, plain '%s', custom '%s', cleared format found %v, types %d", d.drops, d.dropPlain, d.dropCustom, d.junkFound, len(d.dropTypes)))
	// An in-page drag carries no files.
	t.check("drag.files", d.filesErr == nil && len(d.dropFiles) == 0, fmt.Sprintf("files %d, error %v", len(d.dropFiles), d.filesErr))
	t.check("drag.clear_all", len(d.clearedTypes) == 0, fmt.Sprintf("types after clear %d", len(d.clearedTypes)))
	t.check("drag.drop_event", d.dropTag == kui.EventDrag && near(d.dropAt.X, zone.X+zone.Width/2) && near(d.dropAt.Y, zone.Y+zone.Height/2),
		fmt.Sprintf("tag %d, client %.1f,%.1f", d.dropTag, d.dropAt.X, d.dropAt.Y))
	// HTML DnD: dragend's dropEffect is the operation the drop target chose (copy).
	t.knownBug("drag.dragend_effect", d.endEffect == kui.DropCopy, "dragend reports dropEffect 'move' whatever the target chose",
		fmt.Sprintf("dragend effect %d", d.endEffect))
	moveTo("probe")
}

// ---- Panel buttons ----------------------------------------------------------------------

func showProp(label string, notCanceled int) {
	shown := strings.Join(trace, ",")
	if shown == "" {
		shown = "-"
	}
	setText("prop-out", fmt.Sprintf("%s: %s | bubbles=%d cancelable=%d prevented=%d not_canceled=%d",
		label, shown, lastBubbles, lastCancelable, lastPrevented, notCanceled))
}

func showListeners(note string) {
	text := fmt.Sprintf("listeners=%d", len(tickIDs))
	if note != "" {
		text += " | " + note
	}
	setText("lm-out", text)
}

func bindButtons() error {
	tags := []string{"base", "mouse", "wheel", "pointer", "keyboard", "focus", "input", "drag"}
	na := func(err error) string {
		if status(err) == kui.ErrNotSupported {
			return "n/a"
		}
		return "?"
	}
	buttons := map[string]func(){
		"btn-prop-bubble": func() { showProp("ping", fireLayers("ev-inner", "ping", true, false)) },
		"btn-prop-flat":   func() { showProp("ping (no bubbles)", fireLayers("ev-inner", "ping", false, false)) },
		"btn-prop-stop":   func() { showProp("halt", fireLayers("ev-inner", "halt", true, false)) },
		"btn-prop-cancel": func() { showProp("veto", fireLayers("ev-inner", "veto", true, true)) },
		"btn-lm-add":      func() { tickIDs = append(tickIDs, addTick()); showListeners("") },
		"btn-lm-remove": func() {
			if n := len(tickIDs); n > 0 {
				_ = tickTarget.RemoveEventListener("tick", tickIDs[n-1])
				tickIDs = tickIDs[:n-1]
			}
			showListeners("")
		},
		"btn-lm-clear": func() {
			_ = tickTarget.RemoveEventListeners("tick")
			tickIDs, tickSerial = nil, 0
			showListeners("")
		},
		"btn-lm-fire": func() {
			fireTick("tick")
			ran := strings.Join(tickTrace, ",")
			class := "fired"
			if ran == "" {
				ran, class = "none", ""
			}
			setClass("lm-target", class)
			showListeners("ran: " + ran)
		},
		"btn-probe": func() {
			_, _ = dispatch("probe", "probe", true, false)
			p := &probe
			related := "?"
			if p.related == nil && p.relatedErr == nil {
				related = "none"
			}
			setText("probe-out", fmt.Sprintf("type=%s tag=%s bubbles=%d cancelable=%d | related=%s mouse=%s key=%s drag=%s",
				p.typ, name(tags, int(p.tag)), boolInt(p.bubbles), boolInt(p.cancelable), related, na(p.mouseXY), na(p.key), na(p.dragTypes)))
		},
		// The injected drag must not run inside this click's own dispatch: start it from a timer.
		"btn-dnd-simulate": func() {
			setText("drag-seq", "-")
			setText("drop-out", "-")
			_, _ = kui.SetTimeout(simulateDrag, 10)
		},
	}
	for id, fn := range buttons {
		handler := fn
		if err := listen(id, "click", func(*kui.Event) { handler() }); err != nil {
			return err
		}
	}
	return nil
}

func ready(c *runner.Context) error {
	ctx, doc = c, c.Document
	var err error
	if tickTarget, err = doc.GetElementByID("lm-target"); err != nil || tickTarget == nil {
		return fmt.Errorf("没有 #lm-target: %v", err)
	}
	ctx.Keep(tickTarget)
	for _, bind := range []func() error{bindLayers, func() error { return listen("probe", "probe", onProbe) }, bindMouse, bindKeyboard, bindDrag} {
		if err := bind(); err != nil {
			return err
		}
	}

	t := &tour{doc: doc}
	checkPropagation(t)
	checkListeners(t)
	checkProbe(t)
	checkMouse(t)
	checkKeyboard(t)
	checkDrag(t)
	checkWheel(t)
	t.finish()
	return bindButtons()
}
