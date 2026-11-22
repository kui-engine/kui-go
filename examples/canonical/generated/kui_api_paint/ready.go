// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_paint.go; edit that file, not this copy.
// API tour: host painting into custom elements (examples/api_tours/paint). Before the page loads
// the host defines six custom elements with a Paint callback; each callback draws one group of
// PaintContext calls into its content box. Once frames have run, the host checks the callbacks ran
// with the element's size and every call succeeded, asks for repaints (RequestRepaint /
// InvalidateRect) and checks they arrive; #tour-result ends as "PASS n/n". What is drawn is checked
// by the case's pixel steps; the buttons change it.
//
// Binding note: PaintContext.Width / Height / DevicePixelRatio are read with kui_paint_get_info
// before each Paint call; the Go binding has no separate method for it.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------

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

// ---- What the panes draw (the buttons change these) and what the callbacks saw ------------

const (
	paneShapes = iota
	paneText
	paneImage
	paneClip
	paneXform
	paneAlpha
	panes
)

var (
	tags = [panes]string{"paint-shapes", "paint-text", "paint-image", "paint-clip", "paint-xform", "paint-alpha"}
	ids  = [panes]string{"pv-shapes", "pv-text", "pv-image", "pv-clip", "pv-xform", "pv-alpha"}

	paints       [panes]int     // Paint callback calls
	width        [panes]float32 // PaintContext size, last call
	height       [panes]float32
	dpr          [panes]float32
	badCalls     [panes]int // paint calls that returned an error
	unbalanced   error      // Restore without Save, tried once
	unbalancedOK bool

	fill       uint32  = 0xef4444ff // shapes: first rectangle (0xRRGGBBAA)
	lineStyle          = kui.DecorationSolid
	line       uint32  = 0xe11d48ff // text: decoration line colour
	pixelShift int                  // image: rotates the 2x2 buffer
	clipOn             = true
	angle      float32 = 45
	alpha      float32 = 0.5
	imagePath  string

	styleNames = []string{"solid", "double", "dotted", "dashed", "wavy"}
)

func ok(pane int, err error) {
	if err != nil {
		badCalls[pane]++
	}
}

// ---- Paint callbacks: one group of calls per pane -----------------------------------------

// Filled and stroked rectangles at fixed fractions of the width (the case samples them).
func paintShapes(p *kui.PaintContext, w, h float32) {
	ok(paneShapes, p.FillRect(0.03*w, 0.15*h, 0.20*w, 0.70*h, fill))
	ok(paneShapes, p.FillRoundedRect(0.28*w, 0.15*h, 0.20*w, 0.70*h, kui.UniformCornerRadii(16), 0x22c55eff))
	ok(paneShapes, p.StrokeRect(0.53*w, 0.15*h, 0.20*w, 0.70*h, 8, 0x3b82f6ff))
	ok(paneShapes, p.StrokeRoundedRect(0.78*w, 0.15*h, 0.19*w, 0.70*h, kui.UniformCornerRadii(14), 8, 0xa855f7ff))
}

// Bold text hanging from y (the top of its line box, not the baseline), the decoration line band at 0.6h..0.6h+8 and a wavy line below.
func paintText(p *kui.PaintContext, w, h float32) {
	ok(paneText, p.DrawText("Hello paint", 10, 0.45*h, &kui.TextStyle{FontSize: 26, FontWeight: 700}, 0x0f172aff))
	ok(paneText, p.DrawDecorationLine(10, 0.6*h, 0.9*w, 8, lineStyle, line))
	ok(paneText, p.DrawDecorationLine(10, 0.82*h, 0.9*w, 2, kui.DecorationWavy, 0x0ea5e9ff))
}

// A 2x2 premultiplied 0xAARRGGBB buffer scaled into the left box, an image file into the right.
func paintImage(p *kui.PaintContext, w, h float32) {
	colors := []uint32{0xffff0000, 0xff00ff00, 0xff0000ff, 0xffffff00}
	pixels := make([]uint32, 4)
	for i := range pixels {
		pixels[i] = colors[(i+pixelShift)%4]
	}
	ok(paneImage, p.DrawPixels(pixels, 2, 2, 0.03*w, 0.1*h, 0.4*w, 0.8*h))
	ok(paneImage, p.DrawImage(imagePath, kui.Rect{Width: 8, Height: 8},
		kui.Rect{X: 0.55 * w, Y: 0.1 * h, Width: 0.4 * w, Height: 0.8 * h}))
}

// Top band clipped to the left half; bottom band clipped to a rounded rectangle.
func paintClip(p *kui.PaintContext, w, h float32) {
	ok(paneClip, p.Save())
	if clipOn {
		ok(paneClip, p.ClipRect(0, 0, 0.5*w, h))
	}
	ok(paneClip, p.FillRect(0, 0, w, 0.45*h, 0xf97316ff))
	ok(paneClip, p.Restore())
	ok(paneClip, p.Save())
	if clipOn {
		ok(paneClip, p.ClipRoundedRect(0.1*w, 0.55*h, 0.8*w, 0.4*h, kui.UniformCornerRadii(24)))
	}
	ok(paneClip, p.FillRect(0, 0.5*h, w, 0.5*h, 0x14b8a6ff))
	ok(paneClip, p.Restore())
}

// A 60x60 square rotated about (0.25w, 0.5h); a 10x10 square scaled x4 at (0.6w, 0.2h); after the
// last Restore a marker in the bottom-right corner (it lands there only if the transforms were undone).
func paintXform(p *kui.PaintContext, w, h float32) {
	ok(paneXform, p.Save())
	ok(paneXform, p.Translate(0.25*w, 0.5*h))
	ok(paneXform, p.Rotate(angle))
	ok(paneXform, p.FillRect(-30, -30, 60, 60, 0x6366f1ff))
	ok(paneXform, p.Restore())
	ok(paneXform, p.Save())
	ok(paneXform, p.Translate(0.6*w, 0.2*h))
	ok(paneXform, p.Scale(4, 4))
	ok(paneXform, p.FillRect(0, 0, 10, 10, 0xf59e0bff))
	ok(paneXform, p.Restore())
	ok(paneXform, p.FillRect(w-20, h-20, 16, 16, 0x16a34aff))
	if !unbalancedOK {
		// Restore without a matching Save is ErrInvalidArg (kui_c.h).
		unbalancedOK = true
		unbalanced = p.Restore()
	}
}

// Blue at alpha on the left; 0.5 inside 0.5 on the right.
func paintAlpha(p *kui.PaintContext, w, h float32) {
	ok(paneAlpha, p.Save())
	ok(paneAlpha, p.SetOpacity(alpha))
	ok(paneAlpha, p.FillRect(0.05*w, 0.15*h, 0.4*w, 0.7*h, 0x0000ffff))
	ok(paneAlpha, p.Restore())
	ok(paneAlpha, p.Save())
	ok(paneAlpha, p.SetOpacity(0.5))
	ok(paneAlpha, p.Save())
	ok(paneAlpha, p.SetOpacity(0.5))
	ok(paneAlpha, p.FillRect(0.55*w, 0.15*h, 0.4*w, 0.7*h, 0x0000ffff))
	ok(paneAlpha, p.Restore())
	ok(paneAlpha, p.Restore())
}

func painter(pane int) func(*kui.Element, *kui.PaintContext) {
	draw := []func(*kui.PaintContext, float32, float32){paintShapes, paintText, paintImage, paintClip, paintXform, paintAlpha}[pane]
	return func(_ *kui.Element, p *kui.PaintContext) {
		paints[pane]++
		width[pane], height[pane], dpr[pane] = p.Width, p.Height, p.DevicePixelRatio
		draw(p, p.Width, p.Height)
	}
}

// ---- Before the page loads: the image file and the six definitions -----------------------

// writeImage writes an 8x8 24-bit BMP filled with #14b8a6 (teal).
func writeImage() error {
	imagePath = filepath.Join(os.TempDir(), "kui_api_paint_go.bmp")
	le := func(b []byte, at int, v uint32) {
		b[at], b[at+1], b[at+2], b[at+3] = byte(v), byte(v>>8), byte(v>>16), byte(v>>24)
	}
	data := make([]byte, 54+8*24)
	data[0], data[1] = 'B', 'M'
	le(data, 2, uint32(len(data)))
	le(data, 10, 54)
	le(data, 14, 40)
	le(data, 18, 8)
	le(data, 22, 8)
	data[26], data[28] = 1, 24 // planes, bits per pixel
	le(data, 34, 8*24)
	for i := 54; i < len(data); i += 3 { // blue, green, red
		data[i], data[i+1], data[i+2] = 0xa6, 0xb8, 0x14
	}
	return os.WriteFile(imagePath, data, 0o644)
}

func beforeLoad(ctx *runner.Context) error {
	if err := writeImage(); err != nil {
		fmt.Printf("kui_api_paint: cannot write %s: %v\n", imagePath, err)
	}
	for pane := 0; pane < panes; pane++ {
		if err := ctx.Document.DefineCustomElement(tags[pane], kui.CustomElementDefinition{Paint: painter(pane)}); err != nil {
			return err
		}
	}
	return nil
}

// ---- Self-check: definitions now, callbacks and repaints once frames have run -----------

func isInvalidArg(err error) bool {
	var kerr *kui.Error
	return errors.As(err, &kerr) && kerr.Status == kui.ErrInvalidArg
}

func withElement(doc *kui.Document, id string, fn func(*kui.Element)) {
	if element, err := doc.GetElementByID(id); err == nil && element != nil {
		fn(element)
		element.Close()
	}
}

func checkDefinitions(t *tour) {
	all := true
	for _, tag := range tags {
		defined, err := t.doc.CustomElementDefined(tag)
		all = all && err == nil && defined
	}
	err := t.doc.DefineCustomElement("paint-shapes", kui.CustomElementDefinition{Paint: painter(paneShapes)})
	t.check("define_custom_element_ex", all && isInvalidArg(err),
		fmt.Sprintf("all defined %v, second definition %v", all, err))
}

func phasePainted(t *tour) {
	allPainted, sizesMatch, bad := true, true, 0
	var sizes []string
	for pane := 0; pane < panes; pane++ {
		var rect kui.Rect
		withElement(t.doc, ids[pane], func(e *kui.Element) { rect, _ = e.BoundingRect() })
		allPainted = allPainted && paints[pane] > 0
		sizesMatch = sizesMatch && dpr[pane] > 0 && rect.Width > 0 &&
			width[pane] > rect.Width-0.5 && width[pane] < rect.Width+0.5 &&
			height[pane] > rect.Height-0.5 && height[pane] < rect.Height+0.5
		sizes = append(sizes, fmt.Sprintf("%.0fx%.0f/%.0fx%.0f", width[pane], height[pane], rect.Width, rect.Height))
		bad += badCalls[pane]
	}
	t.check("paint.callbacks_run", allPainted, fmt.Sprintf("paints %v", paints))
	t.check("paint_get_info.size", sizesMatch, fmt.Sprintf("info/element %s, dpr %.2f", strings.Join(sizes, " "), dpr[0]))
	t.check("paint.calls_ok", bad == 0, fmt.Sprintf("%d calls returned an error", bad))
	t.check("paint_restore.unbalanced", unbalancedOK && isInvalidArg(unbalanced),
		fmt.Sprintf("tried %v error %v", unbalancedOK, unbalanced))

	marks := paints
	withElement(t.doc, "pv-shapes", func(e *kui.Element) { _ = e.RequestRepaint() })
	withElement(t.doc, "pv-text", func(e *kui.Element) { _ = e.InvalidateRect(0, 60, 300, 20) })
	_, _ = kui.SetTimeout(func() {
		t.check("request_repaint", paints[paneShapes] > marks[paneShapes],
			fmt.Sprintf("paints %d -> %d", marks[paneShapes], paints[paneShapes]))
		t.check("invalidate_rect", paints[paneText] > marks[paneText],
			fmt.Sprintf("paints %d -> %d", marks[paneText], paints[paneText]))
		t.finish()
	}, 300)
}

// ---- Panel buttons: change what a pane draws, then ask for its repaint -------------------

func bindButtons(ctx *runner.Context) error {
	doc := ctx.Document
	setText := func(id, text string) { withElement(doc, id, func(e *kui.Element) { _ = e.SetTextContent(text) }) }
	repaint := func(pane int) { withElement(doc, ids[pane], func(e *kui.Element) { _ = e.RequestRepaint() }) }
	// Repaints only the decoration line's band of the text pane.
	invalidateLine := func() {
		w, h := width[paneText], height[paneText]
		if w <= 0 || h <= 0 {
			w, h = 300, 120
		}
		withElement(doc, "pv-text", func(e *kui.Element) { _ = e.InvalidateRect(0, 0.6*h-2, w, 12) })
	}
	buttons := map[string]func(){
		"btn-shapes-color": func() {
			if fill == 0xef4444ff {
				fill = 0xf97316ff
			} else {
				fill = 0xef4444ff
			}
			repaint(paneShapes)
			setText("shapes-out", fmt.Sprintf("fill=#%06x", fill>>8))
		},
		"btn-paint-info": func() {
			setText("shapes-out", fmt.Sprintf("paints=%d,%d,%d,%d,%d,%d | %.0fx%.0f @%.2f", paints[0], paints[1],
				paints[2], paints[3], paints[4], paints[5], width[paneShapes], height[paneShapes], dpr[paneShapes]))
		},
		"btn-text-style": func() {
			lineStyle = (lineStyle + 1) % 5
			invalidateLine()
			setText("text-out", "style="+styleNames[lineStyle])
		},
		"btn-text-color": func() {
			if line == 0xe11d48ff {
				line = 0x2563ebff
			} else {
				line = 0xe11d48ff
			}
			invalidateLine()
			setText("text-out", fmt.Sprintf("line=#%06x", line>>8))
		},
		"btn-image-rotate": func() {
			pixelShift = (pixelShift + 1) % 4
			repaint(paneImage)
			setText("image-out", "top-left="+[]string{"#ff0000", "#00ff00", "#0000ff", "#ffff00"}[pixelShift])
		},
		"btn-clip-toggle": func() {
			clipOn = !clipOn
			repaint(paneClip)
			if clipOn {
				setText("clip-out", "clip=on")
			} else {
				setText("clip-out", "clip=off")
			}
		},
		"btn-xform-rotate": func() {
			if angle == 0 {
				angle = 45
			} else {
				angle = 0
			}
			repaint(paneXform)
			setText("xform-out", fmt.Sprintf("angle=%.0f", angle))
		},
		"btn-alpha-cycle": func() {
			switch alpha {
			case 0.5:
				alpha = 1
			case 1:
				alpha = 0.25
			default:
				alpha = 0.5
			}
			repaint(paneAlpha)
			setText("alpha-out", fmt.Sprintf("opacity=%g", alpha))
		},
	}
	for id, fn := range buttons {
		element, err := doc.GetElementByID(id)
		if err != nil || element == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(element)
		handler := fn
		if _, err := element.AddEventListener("click", func(*kui.Event) { handler() }); err != nil {
			return err
		}
	}
	return nil
}

func ready(ctx *runner.Context) error {
	t := &tour{doc: ctx.Document}
	checkDefinitions(t)
	// Paint callbacks run with frames: check them once a few have been drawn.
	if _, err := kui.SetTimeout(func() { phasePainted(t) }, 500); err != nil {
		return err
	}
	return bindButtons(ctx)
}
