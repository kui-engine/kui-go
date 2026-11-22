//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_02_accordion.go; edit that file, not this copy.
// Host logic of examples/native/component/02_accordion/main.cpp: clicking header N closes every
// section and reopens section N unless it was already open.
package main

import (
	"fmt"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	setClass := func(id, class string) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			_ = el.SetClassName(class)
			el.Close()
		}
	}
	toggle := func(index int) {
		wasOpen := false
		if header, err := doc.GetElementByID(fmt.Sprintf("h%d", index)); err == nil && header != nil {
			class, _ := header.ClassName()
			wasOpen = strings.Contains(class, "open")
			header.Close()
		}
		for i := 0; i < 4; i++ {
			setClass(fmt.Sprintf("h%d", i), "ah")
			setClass(fmt.Sprintf("b%d", i), "ab")
		}
		if !wasOpen {
			setClass(fmt.Sprintf("h%d", index), "ah open")
			setClass(fmt.Sprintf("b%d", index), "ab open")
		}
	}
	for i := 0; i < 4; i++ {
		header, err := doc.GetElementByID(fmt.Sprintf("h%d", i))
		if err != nil || header == nil {
			return fmt.Errorf("没有 #h%d: %v", i, err)
		}
		ctx.Keep(header)
		index := i
		if _, err := header.AddEventListener("click", func(*kui.Event) { toggle(index) }); err != nil {
			return err
		}
	}
	return nil
}
