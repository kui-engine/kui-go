//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_html_16_css_variables.go; edit that file, not this copy.
// Host logic of examples/native/html/16_css_variables/main.cpp: #themeBtn toggles the theme-dark
// class on <body>, which redefines the CSS variables.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	button, err := doc.GetElementByID("themeBtn")
	if err != nil || button == nil {
		return fmt.Errorf("没有 #themeBtn: %v", err)
	}
	ctx.Keep(button)

	dark := false
	_, err = button.AddEventListener("click", func(*kui.Event) {
		dark = !dark
		body, err := doc.Body()
		if err != nil || body == nil {
			return
		}
		defer body.Close()
		if dark {
			_ = body.AddClass("theme-dark")
		} else {
			_ = body.RemoveClass("theme-dark")
		}
	})
	return err
}
