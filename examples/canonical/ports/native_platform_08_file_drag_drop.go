//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_platform_08_file_drag_drop.go; edit that file, not this copy.
// Host logic of examples/native/platform/08_file_drag_drop/main.cpp: dragenter / dragover on the
// root element highlight #zone (class "over"), dragleave removes it, and a drop of OS files
// lists their full UTF-8 paths in #list.
package main

import (
	"fmt"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	zoneOver := func(on bool) {
		if zone, err := doc.GetElementByID("zone"); err == nil && zone != nil {
			_ = zone.RemoveClass("over")
			if on {
				_ = zone.AddClass("over")
			}
			zone.Close()
		}
	}
	root, err := doc.Root() // the whole window is the drop target
	if err != nil || root == nil {
		return fmt.Errorf("没有根元素: %v", err)
	}
	ctx.Keep(root)

	highlight := func(event *kui.Event) {
		_ = event.PreventDefault() // allows the drop
		zoneOver(true)
	}
	handlers := []struct {
		name string
		fn   func(*kui.Event)
	}{
		{"dragover", highlight},
		{"dragenter", highlight},
		{"dragleave", func(*kui.Event) { zoneOver(false) }},
		{"drop", func(event *kui.Event) {
			zoneOver(false)
			if tag, err := event.TypeTag(); err != nil || tag != kui.EventDrag {
				return
			}
			files, err := event.DragFiles()
			if err != nil || len(files) == 0 {
				return
			}
			_ = event.PreventDefault()
			var html strings.Builder
			fmt.Fprintf(&html, "<div class=\"row count\">收到 %d 个项目:</div>", len(files))
			for _, path := range files {
				html.WriteString("<div class=\"row\">" + path + "</div>")
			}
			if list, err := doc.GetElementByID("list"); err == nil && list != nil {
				_ = list.SetInnerHTML(html.String())
				list.Close()
			}
		}},
	}
	for _, handler := range handlers {
		if _, err := root.AddEventListener(handler.name, handler.fn); err != nil {
			return err
		}
	}
	return nil
}
