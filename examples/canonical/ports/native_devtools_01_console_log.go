//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_devtools_01_console_log.go; edit that file, not this copy.
// Host logic of examples/native/devtools/01_console_log/main.cpp: Log Info / Warn / Error append
// a "log li|lw|le" row to #output (Info reads "User action #<rows added so far>"), Clear empties
// it; F12 DevTools is enabled and the window title gets the live DevTools URL appended.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	_ = kui.SetF12DevToolsEnabled(true)
	// devtools_runtime.h startLiveDevTools: the runner already started DevTools; show its URL.
	if title, err := ctx.Window.Title(); err == nil {
		if url, err := ctx.Window.DevToolsURL(); err == nil && url != "" {
			_ = ctx.Window.SetTitle(title + " — " + url)
		} else {
			_ = ctx.Window.SetTitle(title + " — DevTools failed to start")
		}
	}

	logCount := 0
	addLog := func(cls, msg string) {
		if output, err := doc.GetElementByID("output"); err == nil && output != nil {
			if entry, err := doc.CreateElement("div"); err == nil && entry != nil {
				_ = entry.SetClassName("log " + cls)
				_ = entry.SetTextContent(msg)
				_ = output.AppendChild(entry)
				entry.Close()
			}
			output.Close()
		}
		logCount++
	}
	handlers := map[string]func(){
		"btnI": func() { addLog("li", fmt.Sprintf("User action #%d", logCount)) },
		"btnW": func() { addLog("lw", "Warning: potential issue") },
		"btnE": func() { addLog("le", "Error: operation failed") },
		"btnC": func() {
			if output, err := doc.GetElementByID("output"); err == nil && output != nil {
				_ = output.SetInnerHTML("")
				output.Close()
			}
		},
	}
	for id, fn := range handlers {
		button, err := doc.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(button)
		fn := fn
		if _, err := button.AddEventListener("click", func(*kui.Event) { fn() }); err != nil {
			return err
		}
	}
	return nil
}
