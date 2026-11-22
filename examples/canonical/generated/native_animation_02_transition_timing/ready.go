// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_animation_02_transition_timing.go; edit that file, not this copy.
// Host logic of examples/native/animation/02_transition_timing/main.cpp: #goBtn flips the six
// dots #d1..#d6 between left: calc(100% - 20px) and left: 0, so each runs its own transition.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	button, err := doc.GetElementByID("goBtn")
	if err != nil || button == nil {
		return fmt.Errorf("没有 #goBtn: %v", err)
	}
	ctx.Keep(button)

	atEnd := false
	_, err = button.AddEventListener("click", func(*kui.Event) {
		atEnd = !atEnd
		left := "0"
		if atEnd {
			left = "calc(100% - 20px)"
		}
		for i := 1; i <= 6; i++ {
			if dot, err := doc.GetElementByID(fmt.Sprintf("d%d", i)); err == nil && dot != nil {
				_ = dot.SetStyleProperty("left", left, false)
				dot.Close()
			}
		}
	})
	return err
}
