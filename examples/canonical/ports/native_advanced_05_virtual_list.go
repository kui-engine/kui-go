//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_advanced_05_virtual_list.go; edit that file, not this copy.
// Host logic of examples/native/advanced/05_virtual_list/main.cpp: a 10,000-row VirtualList in
// #listContainer (42px rows, overscan 5), Top / Item 5,000 / Bottom buttons, and a 250 ms
// interval that reports the live DOM range in #stats.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	container, err := doc.GetElementByID("listContainer")
	if err != nil || container == nil {
		return fmt.Errorf("没有 #listContainer: %v", err)
	}
	defer container.Close()
	list, err := kui.NewVirtualList(container, kui.VirtualListConfig{
		TotalItems: 10000,
		ItemHeight: 42,
		Overscan:   5,
		RenderItem: func(itemDoc *kui.Document, index uint32) *kui.Element {
			row, err := itemDoc.CreateElement("div")
			if err != nil || row == nil {
				return nil
			}
			_ = row.SetClassName("row")
			_ = row.SetStyleProperty("height", "42px", false)
			_ = row.SetInnerHTML(fmt.Sprintf(`<span class="idx">#%d</span>Virtualized item %d`, index+1, index+1))
			return row
		},
	})
	if err != nil {
		return err
	}
	ctx.OnClose(func() { _ = list.Close() }) // released on exit, like the original's stack object
	updateStats := func() {
		state, err := list.State()
		if err != nil {
			return
		}
		if stats, err := doc.GetElementByID("stats"); err == nil && stats != nil {
			_ = stats.SetTextContent(fmt.Sprintf("10,000 logical items · live DOM range [%d, %d) · %d rows rendered",
				state.VisibleStart, state.VisibleEnd, state.VisibleEnd-state.VisibleStart))
			stats.Close()
		}
	}
	actions := map[string]func() error{
		"topBtn":    list.ScrollToTop,
		"middleBtn": func() error { return list.ScrollToIndex(4999) },
		"bottomBtn": list.ScrollToBottom,
	}
	for id, action := range actions {
		button, err := doc.GetElementByID(id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(button)
		action := action
		if _, err := button.AddEventListener("click", func(*kui.Event) { _ = action(); updateStats() }); err != nil {
			return err
		}
	}
	_, err = kui.SetInterval(updateStats, 250)
	return err
}
