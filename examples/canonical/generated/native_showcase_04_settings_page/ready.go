// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_showcase_04_settings_page.go; edit that file, not this copy.
// Host logic of examples/native/showcase/04_settings_page/main.cpp: clicking a switch #t0..#t2
// flips its class between "toggle on" and "toggle".
package main

import (
	"fmt"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("t%d", i)
		toggle, err := doc.GetElementByID(id)
		if err != nil || toggle == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(toggle)
		if _, err := toggle.AddEventListener("click", func(*kui.Event) {
			el, err := doc.GetElementByID(id)
			if err != nil || el == nil {
				return
			}
			defer el.Close()
			className, _ := el.ClassName()
			if strings.Contains(className, "on") { // same test as className().find("on")
				_ = el.SetClassName("toggle")
			} else {
				_ = el.SetClassName("toggle on")
			}
		}); err != nil {
			return err
		}
	}
	return nil
}
