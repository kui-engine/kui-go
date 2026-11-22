//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_advanced_04_custom_element.go; edit that file, not this copy.
// Host logic of examples/native/advanced/04_custom_element/main.cpp: defines kui-card and
// kui-badge with lifecycle counters before the page loads (beforeLoad), so parsing the page runs
// the callbacks as in the original; once loaded, marks the first badge with data-live="registered"
// and writes the counters into #lifecycle.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

var connected, disconnected, attributes int

func beforeLoad(ctx *runner.Context) error {
	for _, tag := range []string{"kui-card", "kui-badge"} {
		if err := ctx.Document.DefineCustomElement(tag, kui.CustomElementDefinition{
			Connected:        func(*kui.Element) { connected++ },
			Disconnected:     func(*kui.Element) { disconnected++ },
			AttributeChanged: func(*kui.Element, string, string, string) { attributes++ },
		}); err != nil {
			return err
		}
	}
	return nil
}

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	if badge, err := doc.QuerySelector("kui-badge"); err == nil && badge != nil {
		_ = badge.SetAttribute("data-live", "registered")
		badge.Close()
	}
	lifecycle, err := doc.GetElementByID("lifecycle")
	if err != nil || lifecycle == nil {
		return fmt.Errorf("没有 #lifecycle: %v", err)
	}
	defer lifecycle.Close()
	return lifecycle.SetTextContent(fmt.Sprintf(
		"Registry definitions: kui-card, kui-badge · connected=%d · disconnected=%d · attribute callbacks=%d",
		connected, disconnected, attributes))
}
