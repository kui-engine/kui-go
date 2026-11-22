//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_animation_03_lottie_player.go; edit that file, not this copy.
// Host logic of examples/native/animation/03_lottie_player/main.cpp: the runner's setupLottie
// already points <lottie-player id="lottie"> at the bundled native_animation/pulse.json and plays
// it; this reports the outcome in #status the way the original does. The runner only sets src
// when it found the file, so an empty src means "not found".
package main

import (
	"fmt"

	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	text := "Bundled pulse.json was not found."
	if player, err := doc.GetElementByID("lottie"); err == nil && player != nil {
		defer player.Close()
		if src, _ := player.Attribute("src"); src != "" {
			text = "Lottie decoder rejected pulse.json"
			if complete, _ := player.LottieComplete(); complete {
				size, _ := player.LottieNaturalSize()
				text = fmt.Sprintf("Playing bundled pulse.json · %d × %d", uint32(size.Width), uint32(size.Height))
			}
		}
	}
	status, err := doc.GetElementByID("status")
	if err != nil || status == nil {
		return fmt.Errorf("没有 #status: %v", err)
	}
	defer status.Close()
	return status.SetTextContent(text)
}
