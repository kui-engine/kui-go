//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_html_08_timer_animation.go; edit that file, not this copy.
// Host logic of examples/native/html/08_timer_animation/main.cpp: a 1 s interval writes the local
// time into #clock (%H:%M:%S) and #date (%Y-%m-%d %A), sets #progress's width to sec*100/59 %
// and counts "Ticks: N" in #counter.
package main

import (
	"fmt"
	"time"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	ticks := 0
	setText := func(id, text string) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			_ = el.SetTextContent(text)
			el.Close()
		}
	}
	_, err := kui.SetInterval(func() {
		now := time.Now()
		setText("clock", now.Format("15:04:05"))
		setText("date", now.Format("2006-01-02 Monday"))
		if progress, err := doc.GetElementByID("progress"); err == nil && progress != nil {
			_ = progress.SetStyleProperty("width", fmt.Sprintf("%d%%", now.Second()*100/59), false)
			progress.Close()
		}
		ticks++
		setText("counter", fmt.Sprintf("Ticks: %d", ticks))
	}, 1000)
	return err
}
