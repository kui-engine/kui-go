// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_platform_02_dpi_scaling.go; edit that file, not this copy.
// Host logic of examples/native/platform/02_dpi_scaling/main.cpp: writes the window's device pixel
// ratio into #dpr (std::to_string formatting, "%f") and the truncated percentage into #scale.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func setText(doc *kui.Document, id, text string) {
	if element, err := doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

func ready(ctx *runner.Context) error {
	dpr, err := ctx.Window.DevicePixelRatio()
	if err != nil {
		return err
	}
	setText(ctx.Document, "dpr", fmt.Sprintf("%f", dpr))
	setText(ctx.Document, "scale", fmt.Sprintf("%d%%", int(dpr*100.0)))
	return nil
}
