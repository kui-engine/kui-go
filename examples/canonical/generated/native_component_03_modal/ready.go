// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_03_modal.go; edit that file, not this copy.
// Host logic of examples/native/component/03_modal/main.cpp: the two open buttons fill and show
// the modal; Cancel / Confirm hide it and report the choice in the status line.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	with := func(id string, fn func(*kui.Element)) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			fn(el)
			el.Close()
		}
	}
	setText := func(id, text string) { with(id, func(el *kui.Element) { _ = el.SetTextContent(text) }) }
	setClass := func(id, class string) { with(id, func(el *kui.Element) { _ = el.SetClassName(class) }) }
	open := func(title, message, okClass string) func(*kui.Event) {
		return func(*kui.Event) {
			setText("mT", title)
			setText("mM", message)
			setClass("okBtn", okClass)
			setClass("ov", "overlay show")
		}
	}
	closeWith := func(status string) func(*kui.Event) {
		return func(*kui.Event) {
			setClass("ov", "overlay")
			setText("status", status)
		}
	}
	handlers := []struct {
		id string
		fn func(*kui.Event)
	}{
		{"openC", open("Confirm Action", "Are you sure you want to proceed?", "btn bp")},
		{"openD", open("Delete Item", "This cannot be undone. Delete permanently?", "btn bd")},
		{"canBtn", closeWith("Cancelled.")},
		{"okBtn", closeWith("Confirmed!")},
	}
	for _, h := range handlers {
		el, err := doc.GetElementByID(h.id)
		if err != nil || el == nil {
			return fmt.Errorf("没有 #%s: %v", h.id, err)
		}
		ctx.Keep(el)
		if _, err := el.AddEventListener("click", h.fn); err != nil {
			return err
		}
	}
	return nil
}
