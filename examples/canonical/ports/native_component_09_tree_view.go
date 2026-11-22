//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_09_tree_view.go; edit that file, not this copy.
// Host logic of examples/native/component/09_tree_view/main.cpp: clicking folder label fN toggles
// "open" on its children list cN; clicking a file (i0..i5) writes "Selected: <file>".
package main

import (
	"fmt"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	for f := 0; f < 3; f++ {
		label, err := doc.GetElementByID(fmt.Sprintf("f%d", f))
		if err != nil || label == nil {
			return fmt.Errorf("没有 #f%d: %v", f, err)
		}
		ctx.Keep(label)
		childrenID := fmt.Sprintf("c%d", f)
		_, err = label.AddEventListener("click", func(*kui.Event) {
			children, err := doc.GetElementByID(childrenID)
			if err != nil || children == nil {
				return
			}
			defer children.Close()
			class, _ := children.ClassName()
			if strings.Contains(class, "open") {
				_ = children.SetClassName("children")
			} else {
				_ = children.SetClassName("children open")
			}
		})
		if err != nil {
			return err
		}
	}
	for i := 0; i < 6; i++ {
		file, err := doc.GetElementByID(fmt.Sprintf("i%d", i))
		if err != nil || file == nil {
			continue // the original skips missing files
		}
		ctx.Keep(file)
		_, err = file.AddEventListener("click", func(event *kui.Event) {
			target, err := event.Target()
			if err != nil || target == nil {
				return
			}
			defer target.Close()
			text, _ := target.TextContent()
			if status, err := doc.GetElementByID("status"); err == nil && status != nil {
				_ = status.SetTextContent("Selected: " + text)
				status.Close()
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}
