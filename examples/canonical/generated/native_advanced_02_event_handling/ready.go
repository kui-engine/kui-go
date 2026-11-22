// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_advanced_02_event_handling.go; edit that file, not this copy.
// Host logic of examples/native/advanced/02_event_handling/main.cpp: click / mouseenter /
// mouseleave on the target each append a log entry (click with its client coordinates) and
// scroll the log to the bottom.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	addLog := func(kind, detail string) {
		log, err := doc.GetElementByID("log")
		if err != nil || log == nil {
			return
		}
		defer log.Close()
		entry, err := doc.CreateElement("div")
		if err != nil || entry == nil {
			return
		}
		defer entry.Close()
		_ = entry.SetClassName("log-entry")
		_ = entry.SetInnerHTML(`<span class="type">` + kind + `</span> ` + detail)
		_ = log.AppendChild(entry)
		if metrics, err := log.ScrollMetrics(); err == nil {
			_ = log.ScrollTo(metrics.X, metrics.Height)
		}
	}

	target, err := doc.GetElementByID("target")
	if err != nil || target == nil {
		return fmt.Errorf("没有 #target: %v", err)
	}
	ctx.Keep(target)
	if _, err := target.AddEventListener("click", func(event *kui.Event) {
		point, _ := event.MouseClient()
		addLog("click", fmt.Sprintf("x=%d y=%d", int(point.X), int(point.Y)))
	}); err != nil {
		return err
	}
	if _, err := target.AddEventListener("mouseenter", func(*kui.Event) { addLog("mouseenter", "") }); err != nil {
		return err
	}
	_, err = target.AddEventListener("mouseleave", func(*kui.Event) { addLog("mouseleave", "") })
	return err
}
