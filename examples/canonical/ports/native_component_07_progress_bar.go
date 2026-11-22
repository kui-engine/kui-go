//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_07_progress_bar.go; edit that file, not this copy.
// Host logic of examples/native/component/07_progress_bar/main.cpp: Start runs a 50 ms interval
// that grows #pf1 / #lbl1 by 2% up to 100% (ignored while one is running); Reset stops it and
// goes back to 0%.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	progress := 0
	var timer uint64
	show := func() {
		value := fmt.Sprintf("%d%%", progress)
		if bar, err := doc.GetElementByID("pf1"); err == nil && bar != nil {
			_ = bar.SetStyleProperty("width", value, false)
			bar.Close()
		}
		if label, err := doc.GetElementByID("lbl1"); err == nil && label != nil {
			_ = label.SetTextContent(value)
			label.Close()
		}
	}
	start, err := doc.GetElementByID("startBtn")
	if err != nil || start == nil {
		return fmt.Errorf("没有 #startBtn: %v", err)
	}
	ctx.Keep(start)
	reset, err := doc.GetElementByID("resetBtn")
	if err != nil || reset == nil {
		return fmt.Errorf("没有 #resetBtn: %v", err)
	}
	ctx.Keep(reset)

	if _, err := start.AddEventListener("click", func(*kui.Event) {
		if timer != 0 {
			return
		}
		progress = 0
		id, err := kui.SetInterval(func() {
			progress += 2
			if progress > 100 {
				progress = 100
			}
			show()
			if progress >= 100 {
				_ = kui.ClearTimer(timer)
				timer = 0
			}
		}, 50)
		if err == nil {
			timer = id
		}
	}); err != nil {
		return err
	}
	_, err = reset.AddEventListener("click", func(*kui.Event) {
		if timer != 0 {
			_ = kui.ClearTimer(timer)
			timer = 0
		}
		progress = 0
		show()
	})
	return err
}
