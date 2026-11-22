// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_showcase_02_chat_app.go; edit that file, not this copy.
// Host logic of examples/native/showcase/02_chat_app/main.cpp: Send appends the typed text to
// #msgs as a "msg sent" bubble and clears #msgInput; an empty input sends nothing.
//
// The C++ original reads and clears the input through the "value" content attribute, which stays
// empty while the user types, so it never sends anything. This port reads and clears the live
// value (Element.Value / Element.SetValue).
// ABI accessor.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	send, err := doc.GetElementByID("sendBtn")
	if err != nil || send == nil {
		return fmt.Errorf("没有 #sendBtn: %v", err)
	}
	ctx.Keep(send)

	_, err = send.AddEventListener("click", func(*kui.Event) {
		input, err := doc.GetElementByID("msgInput")
		if err != nil || input == nil {
			return
		}
		defer input.Close()
		text, _ := input.Value()
		if text == "" {
			return
		}
		msg, err := doc.CreateElement("div")
		if err != nil {
			return
		}
		defer msg.Close()
		_ = msg.SetClassName("msg sent")
		_ = msg.SetTextContent(text)
		if list, err := doc.GetElementByID("msgs"); err == nil && list != nil {
			_ = list.AppendChild(msg)
			list.Close()
		}
		_ = input.SetValue("")
	})
	return err
}
