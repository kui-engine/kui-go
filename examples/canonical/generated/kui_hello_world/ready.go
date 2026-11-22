// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_hello_world.go; edit that file, not this copy.
// Host logic of examples/hello_world/main.cpp: #clickBtn increments #counter; the counter turns
// #f59e0b from 5 clicks and #10b981 from 10.
package main

import (
	"fmt"
	"strconv"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	btn, err := doc.GetElementByID("clickBtn")
	if err != nil || btn == nil {
		return fmt.Errorf("没有 #clickBtn: %v", err)
	}
	ctx.Keep(btn)
	count := 0
	_, err = btn.AddEventListener("click", func(*kui.Event) {
		count++
		counter, err := doc.GetElementByID("counter")
		if err != nil || counter == nil {
			return
		}
		defer counter.Close()
		_ = counter.SetTextContent(strconv.Itoa(count))
		if count >= 10 {
			_ = counter.SetStyleProperty("color", "#10b981", false)
		} else if count >= 5 {
			_ = counter.SetStyleProperty("color", "#f59e0b", false)
		}
	})
	return err
}
