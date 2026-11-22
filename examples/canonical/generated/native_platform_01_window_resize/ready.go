// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_platform_01_window_resize.go; edit that file, not this copy.
// Host logic of examples/native/platform/01_window_resize/main.cpp: #size shows the live client
// size, once at start and again on every window resize.
package main

import (
	"fmt"

	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	updateSize := func(width, height uint32) {
		size, err := doc.GetElementByID("size")
		if err != nil || size == nil {
			return
		}
		defer size.Close()
		_ = size.SetTextContent(fmt.Sprintf("Live client size: %d × %d CSS px", width, height))
	}
	width, height, err := ctx.Window.Size()
	if err != nil {
		return err
	}
	updateSize(uint32(width), uint32(height))
	return ctx.Window.OnResize(updateSize)
}
