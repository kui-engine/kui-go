// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_01_tabs.go; edit that file, not this copy.
// Host logic of examples/native/component/01_tabs/main.cpp: clicking tab N marks tab N and
// panel N active and every other tab / panel inactive.
package main

import (
	"fmt"

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
	switchTab := func(index int) {
		for i := 0; i < 4; i++ {
			tab, panel := "tab", "panel"
			if i == index {
				tab, panel = "tab active", "panel active"
			}
			setClass(fmt.Sprintf("tab%d", i), tab)
			setClass(fmt.Sprintf("p%d", i), panel)
		}
	}
	for i := 0; i < 4; i++ {
		tab, err := doc.GetElementByID(fmt.Sprintf("tab%d", i))
		if err != nil || tab == nil {
			return fmt.Errorf("没有 #tab%d: %v", i, err)
		}
		ctx.Keep(tab)
		index := i
		if _, err := tab.AddEventListener("click", func(*kui.Event) { switchTab(index) }); err != nil {
			return err
		}
	}
	return nil
}
