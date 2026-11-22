//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_platform_07_window_events.go; edit that file, not this copy.
// Host logic of examples/native/platform/07_window_events/main.cpp: window focus / blur /
// focusin / focusout / movestart / move / moveend / dragregionpointerdown / close listeners
// update the #b-focus / #b-move badges and prepend lines to #log; the custom title-bar buttons
// minimize and close the (frameless) window.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc, window := ctx.Document, ctx.Window
	withElement := func(id string, fn func(*kui.Element)) {
		if element, err := doc.GetElementByID(id); err == nil && element != nil {
			fn(element)
			element.Close()
		}
	}
	// Newest line first: #log = line + "\n" + #log.
	logLine := func(line string) {
		withElement("log", func(log *kui.Element) {
			old, _ := log.TextContent()
			_ = log.SetTextContent(line + "\n" + old)
		})
	}
	setBadge := func(id, text string, addActive, removeActive bool) {
		withElement(id, func(badge *kui.Element) {
			_ = badge.SetTextContent(text)
			if removeActive {
				_ = badge.RemoveClass("active")
			}
			if addActive {
				_ = badge.AddClass("active")
			}
		})
	}

	events := []struct {
		name string
		fn   func()
	}{
		{"focus", func() {
			setBadge("b-focus", "app: focus", true, true)
			logLine("[focus]  application activated")
		}},
		{"blur", func() {
			setBadge("b-focus", "app: blur", false, true)
			logLine("[blur]   application deactivated")
		}},
		{"focusin", func() { logLine("[focusin]  keyboard focus gained") }},
		{"focusout", func() { logLine("[focusout] keyboard focus lost") }},
		{"movestart", func() {
			setBadge("b-move", "move: dragging", false, false)
			logLine("[movestart] window drag begin (passed OS drag threshold)")
		}},
		{"move", func() { setBadge("b-move", "move: moving", false, false) }},
		{"moveend", func() {
			setBadge("b-move", "move: idle", false, false)
			logLine("[moveend]  window drag end")
		}},
		{"dragregionpointerdown", func() { logLine("[dragregionpointerdown] title-bar pressed") }},
		{"close", func() { logLine("[close]  close requested (allowed)") }},
	}
	for _, event := range events {
		fn := event.fn
		if _, err := window.AddEventListener(event.name, false, func(*kui.Event) { fn() }); err != nil {
			return err
		}
	}

	buttons := map[string]func(){
		"btn-min":   func() { _ = window.Minimize() },
		"btn-close": func() { _ = window.CloseRequest() }, // Close would release the handle
	}
	for id, fn := range buttons {
		button, err := doc.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(button)
		fn := fn
		if _, err := button.AddEventListener("click", func(*kui.Event) { fn() }); err != nil {
			return err
		}
	}
	logLine("ready — try dragging the title bar / switching windows")
	return nil
}
