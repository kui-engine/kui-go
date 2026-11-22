//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_scrollbar_overlay_demo.go; edit that file, not this copy.
// Host logic of examples/scrollbar_overlay_demo/main.cpp: #toggle-target starts with
// -kui-scrollbar-overlay: overlay and #toggle-btn flips it between overlay and auto (updating
// #toggle-status); #btn-top / #btn-bottom scrollTo the top / bottom of #scrollzone.
//
// #btn-fling starts an inertial fling (Element.StartMomentumScroll(0, 1.5), px/ms downward).
// The original's registerBuiltinComponents runs before loadHTML and the page uses none, so it
// is not repeated here.
package main

import (
	"fmt"
	"os"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	withElement := func(id string, fn func(*kui.Element)) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			fn(el)
			el.Close()
		}
	}
	listen := func(id string, fn func(*kui.Event)) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			ctx.Keep(el)
			_, _ = el.AddEventListener("click", fn)
		}
	}

	if target, err := doc.GetElementByID("toggle-target"); err == nil && target != nil {
		_ = target.SetStyleProperty("-kui-scrollbar-overlay", "overlay", false)
		target.Close()
		overlayOn := true
		listen("toggle-btn", func(*kui.Event) {
			overlayOn = !overlayOn
			mode, note := "auto", "  (scrollbar reserves 8px)"
			if overlayOn {
				mode, note = "overlay", "  (scrollbar overlays content)"
			}
			withElement("toggle-target", func(el *kui.Element) {
				_ = el.SetStyleProperty("-kui-scrollbar-overlay", mode, false)
			})
			withElement("toggle-status", func(el *kui.Element) {
				_ = el.SetTextContent(fmt.Sprintf("current: -kui-scrollbar-overlay: %s%s", mode, note))
			})
		})
	} else {
		fmt.Fprintln(os.Stderr, "[scrollbar_overlay_demo] WARN: toggle elements missing; runtime switch disabled")
	}

	if zone, err := doc.GetElementByID("scrollzone"); err == nil && zone != nil {
		zone.Close()
		scrollZone := func(toBottom bool) {
			withElement("scrollzone", func(el *kui.Element) {
				metrics, _ := el.ScrollMetrics()
				top := float32(0)
				if toBottom {
					top = metrics.Height
				}
				// scroll-behavior:smooth on .scrollzone makes this animate.
				_ = el.ScrollTo(metrics.X, top)
			})
		}
		listen("btn-top", func(*kui.Event) { scrollZone(false) })
		listen("btn-bottom", func(*kui.Event) { scrollZone(true) })
		listen("btn-fling", func(*kui.Event) {
			withElement("scrollzone", func(el *kui.Element) { _ = el.StartMomentumScroll(0, 1.5) }) // px/ms downward
		})
	}
	return nil
}
