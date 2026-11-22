//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_geometry_focus.go; edit that file, not this copy.
// API tour: layout geometry, scrolling, anchored placement, text measurement, focus and pointer
// capture (examples/api_tours/geometry_focus). On start the host calls every method once, checks
// the numbers against the CSS on the page (and against each other: offset vs bounding rect,
// scroll offset vs the content's moved rect) and logs one row per check into #tour-log;
// #tour-result ends as "PASS n/n". The panel buttons repeat the calls on the visible boxes.
//
// Mouse input for the pointer-capture checks is injected with Document.DispatchMouse, which runs
// the same dispatch path as real input.
package main

import (
	"fmt"
	"math"
	"strconv"
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

// ---- Small readers (each looks the element up and closes it again) ----------------------

var doc *kui.Document

func with(id string, fn func(*kui.Element)) {
	if element, err := doc.GetElementByID(id); err == nil && element != nil {
		fn(element)
		element.Close()
	}
}

func near(a, b, tolerance float32) bool { return float32(math.Abs(float64(a-b))) <= tolerance }

func round(v float32) float32 { return float32(math.Floor(float64(v) + 0.5)) }

func rectOf(id string) (r kui.Rect) {
	with(id, func(e *kui.Element) { r, _ = e.BoundingRect() })
	return
}

func clientOf(id string) (r kui.Rect) {
	with(id, func(e *kui.Element) { r, _ = e.ClientMetrics() })
	return
}

func offsetOf(id string) (r kui.Rect) {
	with(id, func(e *kui.Element) { r, _ = e.OffsetMetrics() })
	return
}

func scrollOf(id string) (r kui.Rect) {
	with(id, func(e *kui.Element) { r, _ = e.ScrollMetrics() })
	return
}

func is(r kui.Rect, x, y, w, h float32) bool {
	return near(r.X, x, 0.01) && near(r.Y, y, 0.01) && near(r.Width, w, 0.01) && near(r.Height, h, 0.01)
}

func box(r kui.Rect) string { return fmt.Sprintf("%.0f,%.0f %.0fx%.0f", r.X, r.Y, r.Width, r.Height) }

func focused(id string) int {
	value := -1
	with(id, func(e *kui.Element) {
		if yes, err := e.HasFocus(); err == nil {
			value = map[bool]int{false: 0, true: 1}[yes]
		}
	})
	return value
}

func captured(id string, pointer int) int {
	value := -1
	with(id, func(e *kui.Element) {
		if yes, err := e.HasPointerCapture(pointer); err == nil {
			value = map[bool]int{false: 0, true: 1}[yes]
		}
	})
	return value
}

// mouseAt injects one mouse event at the centre of an element (viewport coordinates).
func mouseAt(id string, kind kui.MouseEventType, button kui.MouseButton) {
	r := rectOf(id)
	clicks := 1
	if kind == kui.MouseMoved {
		clicks = 0
	}
	_ = doc.DispatchMouse(kind, r.X+r.Width/2, r.Y+r.Height/2, button, 0, clicks, 0, 0)
}

// ---- Box metrics: bounding rect, client (padding box), offset (border box vs offset parent) --

func checkBoxMetrics(t *tour) {
	frame, b, bb := rectOf("geo-frame"), rectOf("geo-box"), rectOf("geo-bb")
	clientBox, clientBB := clientOf("geo-box"), clientOf("geo-bb")
	offsetBox, offsetBB := offsetOf("geo-box"), offsetOf("geo-bb")

	// content-box: 200 + 2*12 padding + 2*6 border; border-box: the declared 200x50.
	t.check("geo.bounding_rect-size", near(b.Width, 236, 0.5) && near(b.Height, 76, 0.5) &&
		near(bb.Width, 200, 0.5) && near(bb.Height, 50, 0.5),
		fmt.Sprintf("content-box %.1fx%.1f (want 236x76), border-box %.1fx%.1f (want 200x50)", b.Width, b.Height, bb.Width, bb.Height))
	// Frame border 2 + frame padding 8 + box margin 10 from the frame's edge.
	t.check("geo.bounding_rect-position", near(b.X-frame.X, 20, 0.5) && near(b.Y-frame.Y, 20, 0.5),
		fmt.Sprintf("box at +%.1f,+%.1f from the frame (want +20,+20)", b.X-frame.X, b.Y-frame.Y))
	// Padding box: left/top are the border widths.
	t.check("geo.client_metrics", is(clientBox, 6, 6, 224, 64) && is(clientBB, 4, 4, 192, 42),
		fmt.Sprintf("content-box %s (want 6,6 224x64), border-box %s (want 4,4 192x42)", box(clientBox), box(clientBB)))
	// Border box from the offset parent's padding edge; the boxes' 10px margins collapse.
	t.check("geo.offset_metrics", is(offsetBox, 18, 18, 236, 76) && is(offsetBB, 18, 104, 200, 50),
		fmt.Sprintf("content-box %s (want 18,18 236x76), border-box %s (want 18,104 200x50)", box(offsetBox), box(offsetBB)))
	// offsetLeft/Top = the bounding rects' distance minus the parent's border, rounded.
	t.check("geo.offset_vs_rect", offsetBB.X == round(bb.X-frame.X-2) && offsetBB.Y == round(bb.Y-frame.Y-2) &&
		offsetBB.Width == round(bb.Width) && offsetBB.Height == round(bb.Height),
		fmt.Sprintf("offset %.0f,%.0f vs rects %.2f,%.2f", offsetBB.X, offsetBB.Y, bb.X-frame.X-2, bb.Y-frame.Y-2))
}

// ---- Scrolling: extent, ScrollTo (set_scroll_position), ScrollBy, clamping ---------------

func scrollTo(left, top float32) {
	with("scroller", func(e *kui.Element) { _ = e.ScrollTo(left, top) })
}
func scrollBy(dx, dy float32) { with("scroller", func(e *kui.Element) { _ = e.ScrollBy(dx, dy) }) }

func checkScrolling(t *tour) {
	outer, row0 := rectOf("scroller"), rectOf("row-0")
	client, scroll := clientOf("scroller"), scrollOf("scroller")

	t.check("scroll.extent", is(scroll, 0, 0, 600, 400), fmt.Sprintf("scroll %s (want 0,0 600x400)", box(scroll)))
	// 1px border; a rendered scrollbar may take some of the padding box.
	t.check("scroll.client", client.X == 1 && client.Y == 1 && client.Width <= outer.Width-2 &&
		client.Width > outer.Width-24 && client.Height <= 118 && client.Height > 96,
		fmt.Sprintf("client %s of a %.0fx%.0f box", box(client), outer.Width, outer.Height))

	scrollTo(0, 100)
	scroll, moved := scrollOf("scroller"), rectOf("row-0")
	t.check("scroll.set_scroll_position", near(scroll.Y, 100, 0.01) && near(scroll.X, 0, 0.01) && near(row0.Y-moved.Y, 100, 0.5),
		fmt.Sprintf("scroll %.1f,%.1f, row 0 moved up %.1f", scroll.X, scroll.Y, row0.Y-moved.Y))

	scrollBy(30, 50)
	scroll, moved = scrollOf("scroller"), rectOf("row-0")
	t.check("scroll.scroll_by", near(scroll.X, 30, 0.01) && near(scroll.Y, 150, 0.01) &&
		near(row0.X-moved.X, 30, 0.5) && near(row0.Y-moved.Y, 150, 0.5),
		fmt.Sprintf("scroll %.1f,%.1f, row 0 moved %.1f,%.1f", scroll.X, scroll.Y, row0.X-moved.X, row0.Y-moved.Y))

	// Past the end clamps to extent - client size.
	scrollTo(10000, 10000)
	scroll = scrollOf("scroller")
	t.check("scroll.clamp", near(scroll.X, 600-client.Width, 0.5) && near(scroll.Y, 400-client.Height, 0.5),
		fmt.Sprintf("scroll %.1f,%.1f (want %.0f,%.0f)", scroll.X, scroll.Y, 600-client.Width, 400-client.Height))

	scrollTo(0, 0)
	scroll, moved = scrollOf("scroller"), rectOf("row-0")
	t.check("scroll.reset", near(scroll.X, 0, 0.01) && near(scroll.Y, 0, 0.01) && near(moved.Y, row0.Y, 0.5),
		fmt.Sprintf("scroll %.1f,%.1f", scroll.X, scroll.Y))
}

// ---- Anchored placement: PositionRelativeTo with placements, flip and clamp -------------

func place(placement string, offset float32, flip, clamp bool, margin float32) {
	with("pos-anchor", func(anchor *kui.Element) {
		with("pos-popup", func(popup *kui.Element) {
			_ = popup.PositionRelativeTo(anchor, placement,
				kui.PositioningOptions{Flip: flip, ClampToViewport: clamp, Offset: offset, ViewportMargin: margin})
		})
	})
}

// ---- Scroll into view and inertial scrolling ----------------------------------------------------------

func checkScrollReveal(t *tour) {
	scroller, _ := t.doc.GetElementByID("scroller")
	last, _ := t.doc.GetElementByID("row-9")
	defer func() {
		for _, e := range []*kui.Element{scroller, last} {
			if e != nil {
				e.Close()
			}
		}
	}()
	if scroller == nil || last == nil {
		t.check("reveal.setup", false, "missing #scroller or #row-9")
		return
	}
	scrollTo(0, 0)
	_ = last.ScrollIntoView(true)
	scroll, box, row := scrollOf("scroller"), rectOf("scroller"), rectOf("row-9")
	// The last row is now inside the scroll box (1px border).
	t.check("scroll.scroll_into_view", scroll.Y > 100 && row.Y >= box.Y && row.Y+row.Height <= box.Y+box.Height+1,
		fmt.Sprintf("scroll %.1f,%.1f, row 9 at %.1f+%.1f in a box at %.1f+%.1f", scroll.X, scroll.Y, row.Y, row.Height, box.Y, box.Height))

	// Inertial scroll: a velocity (px/ms) that decays frame by frame until the container stops.
	scrollTo(0, 0)
	err := scroller.StartMomentumScroll(0, 0.1)
	var previous, nowY float32 = -1, 0
	for i := 0; i < 80; i++ {
		_ = kui.WaitEvents(0.03)
		nowY = scrollOf("scroller").Y
		if nowY > 0 && nowY == previous {
			break
		}
		previous = nowY
	}
	t.check("scroll.start_momentum_scroll", err == nil && nowY > 0 && nowY < 200, fmt.Sprintf("err %v, stopped at %.1f", err, nowY))
	scrollTo(0, 0)
}

func checkPlacement(t *tour) {
	a := rectOf("pos-anchor")

	place("bottom-start", 6, false, false, 0)
	p := rectOf("pos-popup")
	t.check("pos.bottom-start", near(p.X, a.X, 0.5) && near(p.Y, a.Y+a.Height+6, 0.5),
		fmt.Sprintf("popup %.1f,%.1f anchor %s", p.X, p.Y, box(a)))
	var left string
	with("pos-popup", func(e *kui.Element) { left, _ = e.StyleProperty("left") })
	value, _ := strconv.ParseFloat(strings.TrimSuffix(left, "px"), 32)
	t.check("pos.inline-left", strings.HasSuffix(left, "px") && near(float32(value), a.X, 0.5), fmt.Sprintf("inline left '%s'", left))

	place("right-center", 4, false, false, 0)
	p = rectOf("pos-popup")
	t.check("pos.right-center", near(p.X, a.X+a.Width+4, 0.5) && near(p.Y+p.Height/2, a.Y+a.Height/2, 0.5),
		fmt.Sprintf("popup %s", box(p)))

	place("left-end", 8, false, false, 0)
	p = rectOf("pos-popup")
	t.check("pos.left-end", near(p.X+p.Width, a.X-8, 0.5) && near(p.Y+p.Height, a.Y+a.Height, 0.5),
		fmt.Sprintf("popup %s", box(p)))

	// "top" with a viewport margin below the anchor's top cannot fit: flip puts it below.
	place("top", 5, true, false, a.Y)
	p = rectOf("pos-popup")
	t.check("pos.flip", near(p.Y, a.Y+a.Height+5, 0.5), fmt.Sprintf("popup y %.1f (want %.1f)", p.Y, a.Y+a.Height+5))

	// Far to the left with clamping: held at the viewport margin.
	place("left", 2000, false, true, 7)
	p = rectOf("pos-popup")
	t.check("pos.clamp", near(p.X, 7, 0.5), fmt.Sprintf("popup x %.1f (want 7)", p.X))

	place("bottom-start", 6, true, true, 4)
}

// ---- Text measurement --------------------------------------------------------------------

func measure(text, family string, size float32, weight uint16) (m kui.TextMetrics) {
	with("measure-sample", func(e *kui.Element) {
		m, _ = e.MeasureText(text, &kui.TextStyle{FontFamily: family, FontSize: size, FontWeight: weight})
	})
	return
}

func checkMeasure(t *tour) {
	small, mid, large := measure("Hello KUI", "", 12, 400), measure("Hello KUI", "", 16, 400), measure("Hello KUI", "", 24, 400)
	empty := measure("", "", 16, 400)
	segoe := measure("Hello KUI", "Segoe UI", 16, 400)
	span := rectOf("measure-span")

	t.check("measure.monotonic", small.Width > 0 && small.Width < mid.Width && mid.Width < large.Width,
		fmt.Sprintf("widths %.2f / %.2f / %.2f at 12 / 16 / 24px", small.Width, mid.Width, large.Width))
	t.check("measure.metrics", mid.Ascent > 0 && mid.Descent > 0 && mid.LineHeight >= mid.Ascent+mid.Descent-0.5 &&
		large.Ascent > mid.Ascent, fmt.Sprintf("ascent %.2f descent %.2f line height %.2f", mid.Ascent, mid.Descent, mid.LineHeight))
	t.check("measure.empty", empty.Width == 0 && near(empty.Ascent, mid.Ascent, 0.01),
		fmt.Sprintf("empty width %.2f ascent %.2f", empty.Width, empty.Ascent))
	// Same shaping as layout: the inline-block span with the same text, font and size.
	t.check("measure.matches_layout", near(segoe.Width, span.Width, 1.0),
		fmt.Sprintf("measured %.2f, laid-out span %.2f", segoe.Width, span.Width))
}

// ---- Focus: Focus / Blur / HasFocus, and the focus / blur events they fire ---------------

var focusEvents, blurEvents int

func focusID(id string) { with(id, func(e *kui.Element) { _ = e.Focus() }) }
func blurID(id string)  { with(id, func(e *kui.Element) { _ = e.Blur() }) }

func listen(ctx *runner.Context, id, name string, fn func(*kui.Event)) error {
	element, err := ctx.Document.GetElementByID(id)
	if err != nil || element == nil {
		return fmt.Errorf("没有 #%s: %v", id, err)
	}
	ctx.Keep(element)
	_, err = element.AddEventListener(name, fn)
	return err
}

func checkFocus(ctx *runner.Context, t *tour) error {
	if err := listen(ctx, "f-div", "focus", func(*kui.Event) { focusEvents++ }); err != nil {
		return err
	}
	if err := listen(ctx, "f-input", "blur", func(*kui.Event) { blurEvents++ }); err != nil {
		return err
	}

	focusID("f-btn")
	t.check("focus.button", focused("f-btn") == 1, fmt.Sprintf("has_focus(button) = %d", focused("f-btn")))

	focusID("f-input")
	t.check("focus.moves", focused("f-input") == 1 && focused("f-btn") == 0,
		fmt.Sprintf("input %d, button %d", focused("f-input"), focused("f-btn")))

	blurID("f-input")
	t.check("focus.blur", focused("f-input") == 0 && blurEvents == 1,
		fmt.Sprintf("has_focus(input) = %d, blur events %d", focused("f-input"), blurEvents))

	focusID("f-div")
	t.check("focus.tabindex", focused("f-div") == 1 && focusEvents == 1,
		fmt.Sprintf("has_focus(div) = %d, focus events %d", focused("f-div"), focusEvents))

	// A div without tabindex is not focusable (HTML focus()): Focus changes nothing.
	focusID("f-plain")
	t.knownBug("focus.not_focusable", focused("f-plain") == 0 && focused("f-div") == 1,
		"focus() also focuses an element that is not focusable (div without tabindex)",
		fmt.Sprintf("plain %d, tabindex div %d", focused("f-plain"), focused("f-div")))

	blurID("f-div")
	return nil
}

// ---- Pointer capture: set / has / release, the capture events, routing, implicit release -

var gotCapture, lostCapture, capturedMoves int
var captureTarget bool // gotpointercapture's current target was #cap-box

func capture(on bool) {
	with("cap-box", func(e *kui.Element) {
		if on {
			_ = e.SetPointerCapture(1)
		} else {
			_ = e.ReleasePointerCapture(1)
		}
	})
}

func checkCapture(ctx *runner.Context, t *tour) error {
	err := listen(ctx, "cap-box", "gotpointercapture", func(ev *kui.Event) {
		gotCapture++
		current, _ := ev.CurrentTarget()
		captureTarget = false
		if current != nil {
			id, _ := current.ID()
			captureTarget = id == "cap-box"
			current.Close()
		}
		with("cap-box", func(e *kui.Element) { _ = e.SetClassName("captured") })
	})
	if err == nil {
		err = listen(ctx, "cap-box", "lostpointercapture", func(*kui.Event) {
			lostCapture++
			with("cap-box", func(e *kui.Element) { _ = e.SetClassName("") })
		})
	}
	if err == nil {
		err = listen(ctx, "cap-box", "pointermove", func(*kui.Event) { capturedMoves++ })
	}
	if err != nil {
		return err
	}

	t.check("cap.initial", captured("cap-box", 1) == 0, fmt.Sprintf("has_pointer_capture(1) = %d", captured("cap-box", 1)))

	capture(true)
	t.check("cap.set", captured("cap-box", 1) == 1 && captured("cap-box", 2) == 0 && gotCapture == 1,
		fmt.Sprintf("pointer 1 %d, pointer 2 %d, gotpointercapture %d", captured("cap-box", 1), captured("cap-box", 2), gotCapture))
	// The capture events are dispatched at the capturing element.
	t.knownBug("cap.event_target", captureTarget, "gotpointercapture / lostpointercapture have no target or current target",
		fmt.Sprintf("current target is #cap-box: %v", captureTarget))

	// A move over another element is delivered to the capturing box.
	moves := capturedMoves
	mouseAt("cap-other", kui.MouseMoved, kui.MouseButtonNone)
	t.check("cap.routes_moves", capturedMoves == moves+1, fmt.Sprintf("pointermove on the box %d -> %d", moves, capturedMoves))

	capture(false)
	t.check("cap.release", captured("cap-box", 1) == 0 && lostCapture == 1,
		fmt.Sprintf("has_pointer_capture(1) = %d, lostpointercapture %d", captured("cap-box", 1), lostCapture))

	// Releasing the button ends a capture by itself (implicit release).
	capture(true)
	mouseAt("cap-other", kui.MousePressed, kui.MouseButtonLeft)
	mouseAt("cap-other", kui.MouseReleased, kui.MouseButtonLeft)
	t.check("cap.implicit_release", captured("cap-box", 1) == 0 && lostCapture == 2,
		fmt.Sprintf("has_pointer_capture(1) = %d, lostpointercapture %d", captured("cap-box", 1), lostCapture))
	mouseAt("cap-other", kui.MouseMoved, kui.MouseButtonNone)
	return nil
}

// ---- Panel buttons ----------------------------------------------------------------------

var (
	widePadding bool
	measureSize = 16
	measureBold bool
)

func setText(id, text string) { with(id, func(e *kui.Element) { _ = e.SetTextContent(text) }) }
func setStyle(id, name, value string) {
	with(id, func(e *kui.Element) { _ = e.SetStyleProperty(name, value, false) })
}

func showGeo() {
	client, offset, r := clientOf("geo-box"), offsetOf("geo-box"), rectOf("geo-box")
	setText("geo-out", fmt.Sprintf("rect %.0fx%.0f | client %.0f,%.0f %.0fx%.0f | offset %.0f,%.0f",
		r.Width, r.Height, client.X, client.Y, client.Width, client.Height, offset.X, offset.Y))
}

func showScroll() {
	s := scrollOf("scroller")
	setText("scroll-out", fmt.Sprintf("left=%.0f top=%.0f", s.X, s.Y))
}

func showPos(placement string) {
	a, p := rectOf("pos-anchor"), rectOf("pos-popup")
	setText("pos-out", fmt.Sprintf("%s: dx=%.0f dy=%.0f", placement, p.X-a.X, p.Y-a.Y))
}

func showMeasure() {
	weight, bold := uint16(400), ""
	if measureBold {
		weight, bold = 700, " bold"
	}
	m := measure("Hello KUI", "Segoe UI", float32(measureSize), weight)
	setStyle("measure-sample", "font-size", fmt.Sprintf("%dpx", measureSize))
	setStyle("measure-sample", "font-weight", fmt.Sprint(weight))
	setText("measure-out", fmt.Sprintf("size=%d%s width=%.1f ascent=%.1f descent=%.1f line=%.1f",
		measureSize, bold, m.Width, m.Ascent, m.Descent, m.LineHeight))
}

func showFocus(note string) {
	setText("focus-out", fmt.Sprintf("%s | button=%d input=%d div=%d plain=%d", note,
		focused("f-btn"), focused("f-input"), focused("f-div"), focused("f-plain")))
}

func showCapture() {
	state := "no"
	if captured("cap-box", 1) == 1 {
		state = "yes"
	}
	setText("cap-out", fmt.Sprintf("captured=%s | got=%d lost=%d moves=%d", state, gotCapture, lostCapture, capturedMoves))
}

func bindButtons(ctx *runner.Context) error {
	buttons := map[string]func(){
		"btn-geo-show": showGeo,
		"btn-geo-pad": func() {
			widePadding = !widePadding
			padding := "12px"
			if widePadding {
				padding = "20px"
			}
			setStyle("geo-box", "padding", padding)
			showGeo()
		},
		"btn-scroll-down":  func() { scrollBy(0, 40); showScroll() },
		"btn-scroll-right": func() { scrollBy(50, 0); showScroll() },
		"btn-scroll-end":   func() { scrollTo(10000, 10000); showScroll() },
		"btn-scroll-top":   func() { scrollTo(0, 0); showScroll() },
		"btn-pos-below":    func() { place("bottom-start", 6, true, true, 4); showPos("bottom-start") },
		"btn-pos-right":    func() { place("right-center", 6, true, true, 4); showPos("right-center") },
		"btn-pos-left":     func() { place("left-end", 6, true, true, 4); showPos("left-end") },
		"btn-measure-smaller": func() {
			if measureSize > 8 {
				measureSize -= 4
			}
			showMeasure()
		},
		"btn-measure-larger": func() {
			if measureSize < 40 {
				measureSize += 4
			}
			showMeasure()
		},
		"btn-measure-bold": func() { measureBold = !measureBold; showMeasure() },
		"btn-focus-input":  func() { focusID("f-input"); showFocus("focus(input)") },
		"btn-focus-div":    func() { focusID("f-div"); showFocus("focus(div)") },
		"btn-focus-plain":  func() { focusID("f-plain"); showFocus("focus(plain)") },
		"btn-focus-blur": func() {
			focusID("f-input")
			before := focused("f-input")
			blurID("f-input")
			showFocus(fmt.Sprintf("focus+blur(input) %d->%d", before, focused("f-input")))
		},
		"btn-cap-set":     func() { capture(true); showCapture() },
		"btn-cap-release": func() { capture(false); showCapture() },
		"btn-cap-state":   showCapture,
	}
	for id, fn := range buttons {
		handler := fn
		if err := listen(ctx, id, "click", func(*kui.Event) { handler() }); err != nil {
			return err
		}
	}
	return nil
}

func ready(ctx *runner.Context) error {
	doc = ctx.Document
	t := &tour{doc: ctx.Document}
	checkBoxMetrics(t)
	checkScrolling(t)
	checkScrollReveal(t)
	checkPlacement(t)
	checkMeasure(t)
	if err := checkFocus(ctx, t); err != nil {
		return err
	}
	if err := checkCapture(ctx, t); err != nil {
		return err
	}
	t.finish()
	if err := bindButtons(ctx); err != nil {
		return err
	}
	showMeasure()
	return nil
}
