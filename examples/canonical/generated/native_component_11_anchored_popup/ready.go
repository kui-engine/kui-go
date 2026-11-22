// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_component_11_anchored_popup.go; edit that file, not this copy.
// Host logic of examples/native/component/11_anchored_popup/main.cpp: each .anchor button shows
// #popup next to itself with its data-place placement (PositionRelativeTo, offset 6, C++ defaults
// otherwise); clicking the same button again, or anywhere outside the anchors and the popup,
// hides it.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	popup, err := doc.GetElementByID("popup")
	if err != nil || popup == nil {
		return fmt.Errorf("没有 #popup: %v", err)
	}
	ctx.Keep(popup)
	openAnchor := "" // anchor the popup is open for
	hide := func() {
		_ = popup.SetStyleProperty("display", "none", false)
		openAnchor = ""
	}

	for _, id := range []string{"a-bottom", "a-top", "a-right", "a-left", "a-corner"} {
		button, err := doc.GetElementByID(id)
		if err != nil || button == nil {
			continue // the original skips missing anchors
		}
		ctx.Keep(button)
		anchorID := id
		_, err = button.AddEventListener("click", func(*kui.Event) {
			if openAnchor == anchorID {
				hide()
				return
			}
			place, _ := button.Attribute("data-place")
			if label, err := doc.GetElementByID("pp-place"); err == nil && label != nil {
				_ = label.SetTextContent(place)
				label.Close()
			}
			// Show first: positioning is a no-op while the popup is display:none.
			_ = popup.SetStyleProperty("display", "block", false)
			_ = popup.PositionRelativeTo(button, place,
				kui.PositioningOptions{Flip: true, ClampToViewport: true, Offset: 6, ViewportMargin: 8})
			openAnchor = anchorID
		})
		if err != nil {
			return err
		}
	}

	root, err := doc.Root()
	if err != nil || root == nil {
		return fmt.Errorf("没有根元素: %v", err)
	}
	ctx.Keep(root)
	_, err = root.AddEventListener("click", func(event *kui.Event) {
		if openAnchor == "" {
			return
		}
		// A non-element target cannot be walked; leave the popup alone.
		node, err := event.Target()
		if err != nil || node == nil {
			return
		}
		for node != nil {
			id, _ := node.ID()
			isAnchor, _ := node.HasClass("anchor")
			if id == "popup" || isAnchor {
				node.Close()
				return
			}
			parent, _ := node.Parent()
			node.Close()
			node = parent
		}
		hide()
	})
	return err
}
