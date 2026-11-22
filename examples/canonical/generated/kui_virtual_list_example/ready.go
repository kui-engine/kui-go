// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_virtual_list_example.go; edit that file, not this copy.
// Host logic of examples/virtual_list_example/main.cpp: a 100,000-row VirtualList in
// #vlist-container (32px rows, overscan 5) with insert / remove / scroll buttons, a 50,000-item
// single-selection ListView in #listview-container, and the performance stats in #stats-content.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

type style [2]string

// makeEl is the original's createElement + setTextContent + style()->setProperty.
func makeEl(doc *kui.Document, tag, text string, styles []style) *kui.Element {
	el, err := doc.CreateElement(tag)
	if err != nil || el == nil {
		return nil
	}
	if text != "" {
		_ = el.SetTextContent(text)
	}
	for _, s := range styles {
		_ = el.SetStyleProperty(s[0], s[1], false)
	}
	return el
}

func appendEl(parent, child *kui.Element) {
	if child == nil {
		return
	}
	_ = parent.AppendChild(child)
	child.Close()
}

func renderRow(doc *kui.Document, index uint32) *kui.Element {
	color, background := "#cbd5e1", "rgba(255,255,255,0.02)"
	if index%2 != 0 {
		color, background = "#94a3b8", "transparent"
	}
	row := makeEl(doc, "div", "", []style{
		{"display", "flex"}, {"align-items", "center"}, {"padding", "0 12px"},
		{"font-size", "13px"}, {"color", color}, {"background", background},
		{"font-family", "'Consolas', monospace"},
	})
	if row == nil {
		return nil
	}
	appendEl(row, makeEl(doc, "span", fmt.Sprint(index), []style{
		{"color", "#475569"}, {"width", "70px"}, {"font-size", "11px"}, {"flex-shrink", "0"},
	}))
	text := fmt.Sprintf("Item #%d", index)
	if index%1000 == 0 {
		text += "  ** milestone **"
	}
	if index%10000 == 0 {
		text += "  >>> MAJOR <<<"
	}
	appendEl(row, makeEl(doc, "span", text, []style{{"flex", "1"}}))
	badge := "odd"
	if index%2 == 0 {
		badge = "even"
	}
	appendEl(row, makeEl(doc, "span", badge, []style{
		{"font-size", "10px"}, {"color", "#475569"}, {"flex-shrink", "0"},
	}))
	return row
}

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	var vlist *kui.VirtualList
	var setupMs, insertMs float64
	var insertOps int
	var insertCount uint32
	elapsedMs := func(t0 time.Time) float64 { return float64(time.Since(t0).Nanoseconds()) / 1e6 }
	totalItems := func() uint32 {
		state, _ := vlist.State()
		return state.TotalItems
	}

	// --- Stats display ---
	updateStatsDisplay := func() {
		box, err := doc.GetElementByID("stats-content")
		if err != nil || box == nil {
			return
		}
		defer box.Close()
		_ = box.SetTextContent("") // clear existing content
		addLine := func(text, color string) {
			appendEl(box, makeEl(doc, "div", text, []style{
				{"color", color}, {"font-family", "'Consolas', monospace"},
				{"font-size", "12px"}, {"margin-bottom", "2px"},
			}))
		}
		pick := func(good bool) string {
			if good {
				return "#4ade80"
			}
			return "#f87171"
		}
		state, _ := vlist.State()
		addLine(fmt.Sprintf("Total items: %d", state.TotalItems), "#38bdf8")
		addLine(fmt.Sprintf("Visible range: [%d, %d) (%d rendered)", state.VisibleStart,
			state.VisibleEnd, state.VisibleEnd-state.VisibleStart), "#94a3b8")
		addLine(fmt.Sprintf("Recycle pool: %d elements", state.RecyclePoolSize), "#94a3b8")
		addLine(fmt.Sprintf("Total height: %.0f px", state.TotalHeight), "#94a3b8")
		addLine(fmt.Sprintf("Setup time: %.2f ms", setupMs), pick(setupMs < 16))
		if insertOps > 0 {
			addLine(fmt.Sprintf("Last insert: %.2f ms (%d ops, %d items added)",
				insertMs, insertOps, insertCount), pick(insertMs < 50))
		}
	}

	// --- Button handlers ---
	timedInsert := func(index uint32, where string) {
		t0 := time.Now()
		_ = vlist.InsertItems(index, 1000)
		insertCount += 1000
		insertMs = elapsedMs(t0)
		insertOps++
		updateStatsDisplay()
		fmt.Fprintf(os.Stderr, "[VList] Inserted 1000 items at %s, took %.2f ms\n", where, insertMs)
	}
	handlers := []struct {
		id string
		fn func()
	}{
		{"btn-insert-mid", func() { mid := totalItems() / 2; timedInsert(mid, fmt.Sprint(mid)) }},
		{"btn-insert-begin", func() { timedInsert(0, "0") }},
		{"btn-insert-end", func() { timedInsert(totalItems(), "end") }},
		{"btn-remove-mid", func() {
			if totalItems() < 1000 {
				return
			}
			t0 := time.Now()
			mid := totalItems() / 2
			_ = vlist.RemoveItems(mid, 1000)
			insertMs = elapsedMs(t0)
			insertOps++
			updateStatsDisplay()
			fmt.Fprintf(os.Stderr, "[VList] Removed 1000 items at %d, took %.2f ms\n", mid, insertMs)
		}},
		{"btn-scroll-random", func() {
			total := totalItems()
			if total == 0 {
				return
			}
			idx := uint32(rand.Int63n(int64(total)))
			_ = vlist.ScrollToIndex(idx)
			updateStatsDisplay()
			fmt.Fprintf(os.Stderr, "[VList] Scrolled to index %d\n", idx)
		}},
		{"btn-refresh-stats", updateStatsDisplay},
	}
	for _, h := range handlers {
		button, err := doc.GetElementByID(h.id)
		if err != nil || button == nil {
			return fmt.Errorf("没有 #%s: %v", h.id, err)
		}
		ctx.Keep(button)
		fn := h.fn
		if _, err := button.AddEventListener("click", func(*kui.Event) { fn() }); err != nil {
			return err
		}
	}

	// --- VirtualList: 100,000 rows ---
	container, err := doc.GetElementByID("vlist-container")
	if err != nil || container == nil {
		return fmt.Errorf("没有 #vlist-container: %v", err)
	}
	t0 := time.Now()
	vlist, err = kui.NewVirtualList(container, kui.VirtualListConfig{
		TotalItems: 100000, ItemHeight: 32, Overscan: 5, RenderItem: renderRow,
	})
	setupMs = elapsedMs(t0)
	container.Close()
	if err != nil {
		return err
	}
	ctx.OnClose(func() { _ = vlist.Close() })
	fmt.Fprintf(os.Stderr, "[VList] Setup 100K items in %.2f ms\n", setupMs)

	// --- ListView: 50,000 items with selection ---
	icons := []string{"\U0001F4C4", "\U0001F4C1", "\U0001F527", "\U0001F4E6", "\U0001F310", "⚡"}
	lvContainer, err := doc.GetElementByID("listview-container")
	if err != nil || lvContainer == nil {
		return fmt.Errorf("没有 #listview-container: %v", err)
	}
	listView, err := kui.NewListView(lvContainer, kui.ListViewConfig{
		ItemCount: 50000, ItemHeight: 36, Overscan: 5, SelectionMode: kui.SelectionSingle,
		ItemProvider: func(index uint32) kui.ListViewItem {
			return kui.ListViewItem{
				Icon:     icons[index%6],
				Text:     fmt.Sprintf("List item %d", index),
				Subtitle: fmt.Sprintf("Description for item %d", index),
				Disabled: index%50 == 49,
			}
		},
		OnSelect: func(index uint32) { fmt.Fprintf(os.Stderr, "[ListView] Selected: %d\n", index) },
	})
	lvContainer.Close()
	if err != nil {
		return err
	}
	ctx.OnClose(func() { _ = listView.Close() })

	updateStatsDisplay() // initial stats display
	return nil
}
