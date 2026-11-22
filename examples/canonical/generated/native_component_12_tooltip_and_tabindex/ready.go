// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_12_tooltip_and_tabindex.go; edit that file, not this copy.
// Host logic of examples/native/component/12_tooltip_and_tabindex/main.cpp: logs focus / blur of
// every .fcard and Tab / Shift+Tab keydowns at the root into #log (newest line first). The title
// tooltips and the Tab focus walk are engine-native and need no host code.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	logEl, err := doc.GetElementByID("log")
	if err != nil || logEl == nil {
		return fmt.Errorf("没有 #log: %v", err)
	}
	ctx.Keep(logEl)
	logLine := func(line string) {
		old, _ := logEl.TextContent()
		_ = logEl.SetTextContent(line + "\n" + old)
	}

	cards, err := doc.QuerySelectorAll(".fcard")
	if err != nil {
		return err
	}
	for _, card := range cards {
		ctx.Keep(card)
		name, _ := card.Attribute("data-name")
		if _, err := card.AddEventListener("focus", func(*kui.Event) { logLine("[focus] " + name) }); err != nil {
			return err
		}
		if _, err := card.AddEventListener("blur", func(*kui.Event) { logLine("[blur]  " + name) }); err != nil {
			return err
		}
	}

	root, err := doc.Root()
	if err != nil || root == nil {
		return fmt.Errorf("没有根元素: %v", err)
	}
	ctx.Keep(root)
	_, err = root.AddEventListener("keydown", func(event *kui.Event) {
		if tag, err := event.TypeTag(); err != nil || tag != kui.EventKeyboard {
			return
		}
		if key, err := event.Key(); err != nil || key != "Tab" {
			return
		}
		if mods, err := event.ModifierKeys(); err == nil && mods.Shift {
			logLine("[keydown] Shift+Tab -> focus moves back")
		} else {
			logLine("[keydown] Tab -> focus advances")
		}
	})
	return err
}
