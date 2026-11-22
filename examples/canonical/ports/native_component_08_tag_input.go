//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_08_tag_input.go; edit that file, not this copy.
// Host logic of examples/native/component/08_tag_input/main.cpp: clicking a suggestion (s0..s5)
// appends a new <div class="tag"> with the suggestion's text to #tags.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	addTag := func(text string) {
		container, err := doc.GetElementByID("tags")
		if err != nil || container == nil {
			return
		}
		defer container.Close()
		tag, err := doc.CreateElement("div")
		if err != nil || tag == nil {
			return
		}
		defer tag.Close()
		_ = tag.SetClassName("tag")
		_ = tag.SetTextContent(text)
		_ = container.AppendChild(tag)
	}
	for i := 0; i < 6; i++ {
		suggestion, err := doc.GetElementByID(fmt.Sprintf("s%d", i))
		if err != nil || suggestion == nil {
			return fmt.Errorf("没有 #s%d: %v", i, err)
		}
		ctx.Keep(suggestion)
		_, err = suggestion.AddEventListener("click", func(event *kui.Event) {
			target, err := event.Target()
			if err != nil || target == nil {
				return
			}
			defer target.Close()
			text, _ := target.TextContent()
			addTag(text)
		})
		if err != nil {
			return err
		}
	}
	return nil
}
