//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_10_list_view.go; edit that file, not this copy.
// Host logic of examples/native/component/10_list_view/main.cpp: clicking row N selects it (only
// that row gets "li sel") and writes "Selected row N+1" into the status bar.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	with := func(id string, fn func(*kui.Element)) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			fn(el)
			el.Close()
		}
	}
	for i := 0; i < 5; i++ {
		row, err := doc.GetElementByID(fmt.Sprintf("r%d", i))
		if err != nil || row == nil {
			return fmt.Errorf("没有 #r%d: %v", i, err)
		}
		ctx.Keep(row)
		index := i
		_, err = row.AddEventListener("click", func(*kui.Event) {
			for j := 0; j < 5; j++ {
				with(fmt.Sprintf("r%d", j), func(el *kui.Element) { _ = el.SetClassName("li") })
			}
			with(fmt.Sprintf("r%d", index), func(el *kui.Element) { _ = el.SetClassName("li sel") })
			with("status", func(el *kui.Element) { _ = el.SetTextContent(fmt.Sprintf("Selected row %d", index+1)) })
		})
		if err != nil {
			return err
		}
	}
	return nil
}
