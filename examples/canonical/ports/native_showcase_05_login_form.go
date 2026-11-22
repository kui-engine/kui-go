//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_showcase_05_login_form.go; edit that file, not this copy.
// Host logic of examples/native/showcase/05_login_form/main.cpp: Sign In shows an error in #msg
// when the email or password is empty, otherwise a welcome for the email.
//
// The C++ original reads both inputs with getAttribute("value"), which is the content attribute
// and stays empty while the user types, so it always reports the error. This port reads the live
// values (Element.Value).
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	login, err := doc.GetElementByID("loginBtn")
	if err != nil || login == nil {
		return fmt.Errorf("没有 #loginBtn: %v", err)
	}
	ctx.Keep(login)

	value := func(id string) string {
		input, err := doc.GetElementByID(id)
		if err != nil || input == nil {
			return ""
		}
		defer input.Close()
		text, _ := input.Value()
		return text
	}
	_, err = login.AddEventListener("click", func(*kui.Event) {
		email, pass := value("email"), value("pass")
		msg, err := doc.GetElementByID("msg")
		if err != nil || msg == nil {
			return
		}
		defer msg.Close()
		if email == "" || pass == "" {
			_ = msg.SetClassName("msg msg-err show")
			_ = msg.SetTextContent("Please fill in all fields.")
		} else {
			_ = msg.SetClassName("msg msg-ok show")
			_ = msg.SetTextContent("Login successful! Welcome, " + email)
		}
	})
	return err
}
