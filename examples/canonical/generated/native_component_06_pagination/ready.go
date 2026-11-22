// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_06_pagination.go; edit that file, not this copy.
// Host logic of examples/native/component/06_pagination/main.cpp: page buttons 1..5 and the
// prev / next arrows move the current page and refresh the buttons, content and info line.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	with := func(id string, fn func(*kui.Element)) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			fn(el)
			el.Close()
		}
	}
	setClass := func(id, class string) { with(id, func(el *kui.Element) { _ = el.SetClassName(class) }) }
	page := 1
	update := func() {
		for i := 1; i <= 5; i++ {
			class := "pg"
			if i == page {
				class = "pg active"
			}
			setClass(fmt.Sprintf("pg%d", i), class)
		}
		with("content", func(el *kui.Element) {
			_ = el.SetInnerHTML(fmt.Sprintf("<p>Page %d: Items %d-%d of 50.</p>", page, (page-1)*10+1, page*10))
		})
		with("info", func(el *kui.Element) { _ = el.SetTextContent(fmt.Sprintf("Page %d of 5", page)) })
		prev, next := "pg", "pg"
		if page <= 1 {
			prev = "pg disabled"
		}
		if page >= 5 {
			next = "pg disabled"
		}
		setClass("prev", prev)
		setClass("next", next)
	}
	listen := func(id string, fn func()) error {
		el, err := doc.GetElementByID(id)
		if err != nil || el == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(el)
		_, err = el.AddEventListener("click", func(*kui.Event) { fn() })
		return err
	}
	for i := 1; i <= 5; i++ {
		target := i
		if err := listen(fmt.Sprintf("pg%d", i), func() { page = target; update() }); err != nil {
			return err
		}
	}
	if err := listen("prev", func() {
		if page > 1 {
			page--
			update()
		}
	}); err != nil {
		return err
	}
	return listen("next", func() {
		if page < 5 {
			page++
			update()
		}
	})
}
