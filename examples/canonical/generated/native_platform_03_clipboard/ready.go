// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_platform_03_clipboard.go; edit that file, not this copy.
// Host logic of examples/native/platform/03_clipboard/main.cpp: Copy puts #copyInput's text on
// the system clipboard; Paste shows the clipboard text in #pasteResult ("(empty)" when none).
//
// The original reads getAttribute("value"), which never sees what the user typed; this port
// copies the live value (Element.Value), which is what the original means to do.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	copyButton, err := doc.GetElementByID("copyBtn")
	if err != nil || copyButton == nil {
		return fmt.Errorf("没有 #copyBtn: %v", err)
	}
	ctx.Keep(copyButton)
	pasteButton, err := doc.GetElementByID("pasteBtn")
	if err != nil || pasteButton == nil {
		return fmt.Errorf("没有 #pasteBtn: %v", err)
	}
	ctx.Keep(pasteButton)

	if _, err := copyButton.AddEventListener("click", func(*kui.Event) {
		value := ""
		if input, err := doc.GetElementByID("copyInput"); err == nil && input != nil {
			value, _ = input.Value()
			input.Close()
		}
		_ = kui.SetClipboardText(value)
	}); err != nil {
		return err
	}
	_, err = pasteButton.AddEventListener("click", func(*kui.Event) {
		text, err := kui.ClipboardText()
		if err != nil || text == "" {
			text = "(empty)"
		}
		if result, err := doc.GetElementByID("pasteResult"); err == nil && result != nil {
			_ = result.SetTextContent(text)
			result.Close()
		}
	})
	return err
}
