// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_advanced_01_dom_manipulation.go; edit that file, not this copy.
// Host logic of examples/native/advanced/01_dom_manipulation/main.cpp: Add appends an "Item #n"
// card to the container, Remove Last drops the last card, Clear All empties it; the counter
// always shows the container's element count.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	nextID := 1
	updateCounter := func(container *kui.Element) {
		count, _ := container.ChildCount()
		if counter, err := doc.GetElementByID("counter"); err == nil && counter != nil {
			_ = counter.SetTextContent(fmt.Sprintf("Elements: %d", count))
			counter.Close()
		}
	}
	withContainer := func(fn func(container *kui.Element)) func(*kui.Event) {
		return func(*kui.Event) {
			container, err := doc.GetElementByID("container")
			if err != nil || container == nil {
				return
			}
			defer container.Close()
			fn(container)
		}
	}
	handlers := map[string]func(*kui.Event){
		"addBtn": withContainer(func(container *kui.Element) {
			card, err := doc.CreateElement("div")
			if err != nil || card == nil {
				return
			}
			defer card.Close()
			_ = card.SetClassName("card")
			_ = card.SetTextContent(fmt.Sprintf("Item #%d", nextID))
			nextID++
			_ = container.AppendChild(card)
			updateCounter(container)
		}),
		"removeBtn": withContainer(func(container *kui.Element) {
			last, err := container.LastChild()
			if err != nil || last == nil {
				return
			}
			defer last.Close()
			_ = container.RemoveChild(last)
			updateCounter(container)
		}),
		"clearBtn": withContainer(func(container *kui.Element) {
			_ = container.SetInnerHTML("")
			updateCounter(container)
		}),
	}
	for id, handler := range handlers {
		button, err := doc.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(button)
		if _, err := button.AddEventListener("click", handler); err != nil {
			return err
		}
	}
	return nil
}
