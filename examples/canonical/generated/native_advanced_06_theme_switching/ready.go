// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_advanced_06_theme_switching.go; edit that file, not this copy.
// Host logic of examples/native/advanced/06_theme_switching/main.cpp: the Run Demo button writes
// a success message into the output box and turns it green.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	action, err := doc.GetElementByID("actionBtn")
	if err != nil || action == nil {
		return fmt.Errorf("没有 #actionBtn: %v", err)
	}
	ctx.Keep(action)

	_, err = action.AddEventListener("click", func(*kui.Event) {
		output, err := doc.GetElementByID("output")
		if err != nil || output == nil {
			return
		}
		defer output.Close()
		_ = output.SetTextContent("Theme System demo executed successfully!")
		_ = output.SetStyleProperty("background", "#dcfce7", false)
		_ = output.SetStyleProperty("color", "#166534", false)
	})
	return err
}
