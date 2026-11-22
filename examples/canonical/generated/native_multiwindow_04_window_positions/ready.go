// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_multiwindow_04_window_positions.go; edit that file, not this copy.
// Host logic of examples/native/multiwindow/04_window_positions/main.cpp: moves the window to the
// primary monitor's work area + (48, 48), then a 200 ms interval writes the live OS position into
// #position.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	window, doc := ctx.Window, ctx.Document
	if monitor, err := kui.PrimaryMonitor(); err == nil {
		_ = window.SetPosition(int(monitor.WorkArea.X+48), int(monitor.WorkArea.Y+48))
	}
	_, err := kui.SetInterval(func() {
		x, y, err := window.Position()
		if err != nil {
			return
		}
		if el, err := doc.GetElementByID("position"); err == nil && el != nil {
			_ = el.SetTextContent(fmt.Sprintf("Live window position: (%d, %d)", x, y))
			el.Close()
		}
	}, 200)
	return err
}
