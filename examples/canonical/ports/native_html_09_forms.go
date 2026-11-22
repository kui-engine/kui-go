//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_html_09_forms.go; edit that file, not this copy.
// Host logic of examples/native/html/09_forms/main.cpp: the submit button reads the name input
// and writes a greeting into the output box.
//
// The C++ original reads the input with getAttribute("value"), which is the content attribute and
// stays empty while the user types, so it always greets "(empty)". This port reads the live value
// (Element.Value) so the example does what it means to.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	submit, err := doc.GetElementByID("submitBtn")
	if err != nil || submit == nil {
		return fmt.Errorf("没有 #submitBtn: %v", err)
	}
	ctx.Keep(submit) // 登记了监听的元素要让运行器保管，Close 会移除监听

	_, err = submit.AddEventListener("click", func(*kui.Event) {
		name := ""
		if input, err := doc.GetElementByID("nameInput"); err == nil && input != nil {
			name, _ = input.Value()
			input.Close()
		}
		if name == "" {
			name = "(empty)"
		}
		output, err := doc.GetElementByID("output")
		if err != nil || output == nil {
			return
		}
		defer output.Close()
		_ = output.SetTextContent(fmt.Sprintf("Hello, %s! Form submitted.", name))
		_ = output.SetStyleProperty("background", "#dcfce7", false)
		_ = output.SetStyleProperty("color", "#166534", false)
	})
	return err
}
