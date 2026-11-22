//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_advanced_07_computed_styles.go; edit that file, not this copy.
// Host logic of examples/native/advanced/07_computed_styles/main.cpp: the Run Demo button writes
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
		_ = output.SetTextContent("Reading Computed Styles demo executed successfully!")
		_ = output.SetStyleProperty("background", "#dcfce7", false)
		_ = output.SetStyleProperty("color", "#166534", false)
	})
	return err
}
