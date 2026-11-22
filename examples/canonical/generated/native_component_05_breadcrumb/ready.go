// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_05_breadcrumb.go; edit that file, not this copy.
// Host logic of examples/native/component/05_breadcrumb/main.cpp: clicking any breadcrumb link
// (bc0..bc6) writes "Navigated to: <link text>" into the status line.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	onLink := func(event *kui.Event) {
		target, err := event.Target()
		if err != nil || target == nil {
			return
		}
		defer target.Close()
		text, _ := target.TextContent()
		if status, err := doc.GetElementByID("status"); err == nil && status != nil {
			_ = status.SetTextContent("Navigated to: " + text)
			status.Close()
		}
	}
	for i := 0; i < 7; i++ {
		link, err := doc.GetElementByID(fmt.Sprintf("bc%d", i))
		if err != nil || link == nil {
			continue // the original skips missing links
		}
		ctx.Keep(link)
		if _, err := link.AddEventListener("click", onLink); err != nil {
			return err
		}
	}
	return nil
}
