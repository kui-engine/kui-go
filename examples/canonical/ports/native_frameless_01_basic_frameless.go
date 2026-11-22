//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_frameless_01_basic_frameless.go; edit that file, not this copy.
// Host logic of examples/native/frameless/01_basic_frameless/main.cpp: the custom title bar
// buttons minimize, maximize / restore and close the frameless window.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	window := ctx.Window
	actions := map[string]func(){
		"minBtn": func() { _ = window.Minimize() },
		"maxBtn": func() {
			if maximized, _ := window.IsMaximized(); maximized {
				_ = window.Restore()
			} else {
				_ = window.Maximize()
			}
		},
		"closeBtn": func() { _ = window.Close() },
	}
	for id, action := range actions {
		button, err := ctx.Document.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(button)
		action := action
		if _, err := button.AddEventListener("click", func(*kui.Event) { action() }); err != nil {
			return err
		}
	}
	return nil
}
