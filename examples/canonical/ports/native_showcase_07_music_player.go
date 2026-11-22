//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_showcase_07_music_player.go; edit that file, not this copy.
// Host logic of examples/native/showcase/07_music_player/main.cpp: #playBtn toggles its glyph
// between play (U+25B6) and pause (U+23F8). No audio is involved in the original.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	play, err := ctx.Document.GetElementByID("playBtn")
	if err != nil || play == nil {
		return fmt.Errorf("没有 #playBtn: %v", err)
	}
	ctx.Keep(play)

	playing := false
	_, err = play.AddEventListener("click", func(*kui.Event) {
		playing = !playing
		if playing {
			_ = play.SetTextContent("⏸")
		} else {
			_ = play.SetTextContent("▶")
		}
	})
	return err
}
