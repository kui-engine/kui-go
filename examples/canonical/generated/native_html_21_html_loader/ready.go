// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_html_21_html_loader.go; edit that file, not this copy.
// Host logic of examples/native/html/21_html_loader/main.cpp: once the inline page is loaded,
// #status reports it.
package main

import (
	"fmt"

	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	status, err := ctx.Document.GetElementByID("status")
	if err != nil || status == nil {
		return fmt.Errorf("没有 #status: %v", err)
	}
	defer status.Close()
	return status.SetTextContent("Inline HTML loaded successfully.")
}
