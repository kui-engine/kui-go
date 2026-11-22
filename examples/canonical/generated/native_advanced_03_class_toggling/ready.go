// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_advanced_03_class_toggling.go; edit that file, not this copy.
// Host logic of examples/native/advanced/03_class_toggling/main.cpp: three buttons toggle the
// expanded / highlight / danger classes on the panel and echo its class list.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	toggle := func(className string) {
		panel, err := doc.GetElementByID("panel")
		if err != nil || panel == nil {
			return
		}
		defer panel.Close()
		_ = panel.ToggleClass(className)
		current, _ := panel.ClassName()
		if classes, err := doc.GetElementByID("classes"); err == nil && classes != nil {
			_ = classes.SetTextContent("Classes: " + current)
			classes.Close()
		}
	}
	for id, className := range map[string]string{
		"toggleExpand": "expanded", "toggleHighlight": "highlight", "toggleDanger": "danger",
	} {
		button, err := doc.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(button)
		className := className
		if _, err := button.AddEventListener("click", func(*kui.Event) { toggle(className) }); err != nil {
			return err
		}
	}
	return nil
}
