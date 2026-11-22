//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_platform_06_timer_intervals.go; edit that file, not this copy.
// Host logic of examples/native/platform/06_timer_intervals/main.cpp: Start runs a 100 ms interval
// that bumps #counter and sets #bar's width to count % 100 (ignored while running); Stop clears
// it; Reset clears it and goes back to 0. Each action writes a line into #log.
package main

import (
	"fmt"
	"strconv"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	count := 0
	var timer uint64
	setText := func(id, text string) {
		if element, err := doc.GetElementByID(id); err == nil && element != nil {
			_ = element.SetTextContent(text)
			element.Close()
		}
	}
	updateUI := func() {
		setText("counter", strconv.Itoa(count))
		if bar, err := doc.GetElementByID("bar"); err == nil && bar != nil {
			_ = bar.SetStyleProperty("width", fmt.Sprintf("%d%%", count%100), false)
			bar.Close()
		}
	}
	handlers := map[string]func(){
		"startBtn": func() {
			if timer != 0 {
				return
			}
			setText("log", "Timer started")
			if id, err := kui.SetInterval(func() {
				count++
				updateUI()
			}, 100); err == nil {
				timer = id
			}
		},
		"stopBtn": func() {
			if timer != 0 {
				_ = kui.ClearTimer(timer)
				timer = 0
				setText("log", "Timer stopped")
			}
		},
		"resetBtn": func() {
			if timer != 0 {
				_ = kui.ClearTimer(timer)
				timer = 0
			}
			count = 0
			updateUI()
			setText("log", "Timer reset")
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
