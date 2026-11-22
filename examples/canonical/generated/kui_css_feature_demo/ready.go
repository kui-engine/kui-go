// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_css_feature_demo.go; edit that file, not this copy.
// Host logic of examples/css_feature_demo/main.cpp: builds the "C++ Dynamic DOM" section into
// #dynamic-section through the DOM API, updates #status-text, and binds the sidebar nav items
// (active class + scroll the section into view). The original loads css_feature_test.html itself;
// the runner loads that same file as this example's page.
// The nav items scroll their section into view with Element.ScrollIntoView.
package main

import (
	"fmt"
	"os"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

type style [2]string

// makeEl is the original's createElement + setTextContent + style()->setProperty.
func makeEl(doc *kui.Document, tag, text string, styles ...style) *kui.Element {
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

func heading(doc *kui.Document, section *kui.Element, text string) {
	appendEl(section, makeEl(doc, "div", text,
		style{"font-size", "12px"}, style{"color", "var(--c-dim)"},
		style{"text-transform", "uppercase"}, style{"letter-spacing", "1px"},
		style{"margin-bottom", "6px"}))
}

func buildDynamicSection(doc *kui.Document) {
	container, err := doc.GetElementByID("dynamic-section")
	if err != nil || container == nil {
		return
	}
	defer container.Close()

	section := makeEl(doc, "div", "",
		style{"background", "var(--c-surface)"}, style{"border", "1px solid var(--c-border)"},
		style{"border-radius", "var(--radius)"}, style{"padding", "16px"},
		style{"margin-bottom", "16px"})
	appendEl(section, makeEl(doc, "div", "C++ Dynamic DOM Manipulation",
		style{"font-size", "15px"}, style{"font-weight", "700"},
		style{"color", "var(--c-accent)"}, style{"margin-bottom", "10px"},
		style{"display", "flex"}, style{"align-items", "center"}, style{"gap", "8px"}))
	appendEl(section, makeEl(doc, "div", "The following elements are created entirely via C++ API.",
		style{"font-size", "11px"}, style{"color", "var(--c-dim)"}, style{"margin-bottom", "12px"}))

	// --- 1. Dynamic color palette ---
	heading(doc, section, "1. Dynamic Color Palette (C++ generated)")
	row := makeEl(doc, "div", "", style{"display", "flex"}, style{"gap", "4px"},
		style{"flex-wrap", "wrap"}, style{"margin-bottom", "12px"})
	for i := 0; i < 16; i++ {
		hue := i * 22
		appendEl(row, makeEl(doc, "div", fmt.Sprint(hue),
			style{"width", "36px"}, style{"height", "36px"},
			style{"background", fmt.Sprintf("hsl(%d, 70%%, 55%%)", hue)},
			style{"border-radius", "6px"}, style{"display", "flex"}, style{"align-items", "center"},
			style{"justify-content", "center"}, style{"font-size", "9px"}, style{"color", "white"},
			style{"font-weight", "700"}, style{"transition", "transform 0.2s ease, box-shadow 0.2s"},
			style{"cursor", "pointer"}))
	}
	appendEl(section, row)

	// --- 2. Dynamic flex layout (proportional bars) ---
	heading(doc, section, "2. Dynamic Flex Layout (proportional bars)")
	row = makeEl(doc, "div", "", style{"display", "flex"}, style{"gap", "3px"},
		style{"height", "30px"}, style{"margin-bottom", "12px"})
	labels := []string{"CSS 2.1", "Selectors", "Flexbox", "Grid", "Transforms", "Animations", "Colors"}
	values := []int{98, 85, 100, 95, 100, 100, 88}
	colors := []string{"#e17055", "#fdcb6e", "#00b894", "#0984e3", "#6c5ce7", "#a29bfe", "#fd79a8"}
	for i := range labels {
		appendEl(row, makeEl(doc, "div", labels[i],
			style{"flex-grow", fmt.Sprint(values[i])}, style{"background", colors[i]},
			style{"border-radius", "4px"}, style{"display", "flex"}, style{"align-items", "center"},
			style{"justify-content", "center"}, style{"font-size", "9px"}, style{"color", "white"},
			style{"font-weight", "700"}, style{"overflow", "hidden"}, style{"white-space", "nowrap"},
			style{"transition", "flex-grow 0.5s ease"}))
	}
	appendEl(section, row)

	// --- 3. Dynamic grid dashboard ---
	heading(doc, section, "3. Dynamic Grid Dashboard (C++ data binding)")
	grid := makeEl(doc, "div", "", style{"display", "grid"},
		style{"grid-template-columns", "repeat(4, 1fr)"}, style{"gap", "8px"},
		style{"margin-bottom", "12px"})
	stats := [][4]string{
		{"CSS Properties", "307", "#6c5ce7", "+12"},
		{"Selectors", "44/52", "#00b894", "85%"},
		{"Layout Tests", "100%", "#0984e3", "Flex+Grid"},
		{"Animations", "11/11", "#fd79a8", "100%"},
	}
	for _, st := range stats {
		card := makeEl(doc, "div", "", style{"background", "var(--c-card)"},
			style{"border-radius", "8px"}, style{"padding", "12px"},
			style{"border-left", "3px solid " + st[2]})
		appendEl(card, makeEl(doc, "div", st[0], style{"font-size", "10px"},
			style{"color", "var(--c-dim)"}, style{"margin-bottom", "4px"}))
		appendEl(card, makeEl(doc, "div", st[1], style{"font-size", "20px"},
			style{"font-weight", "800"}, style{"color", st[2]}))
		appendEl(card, makeEl(doc, "div", st[3], style{"font-size", "10px"},
			style{"color", "var(--c-dim)"}, style{"margin-top", "2px"}))
		appendEl(grid, card)
	}
	appendEl(section, grid)

	// --- 4. Programmatic style manipulation ---
	heading(doc, section, "4. Programmatic Style Manipulation")
	row = makeEl(doc, "div", "", style{"display", "flex"}, style{"gap", "6px"},
		style{"flex-wrap", "wrap"}, style{"margin-bottom", "12px"})
	demos := []struct {
		label  string
		styles []style
	}{
		{"border-radius: 50%", []style{{"width", "40px"}, {"height", "40px"},
			{"background", "var(--c-primary)"}, {"border-radius", "50%"}, {"display", "flex"},
			{"align-items", "center"}, {"justify-content", "center"}, {"color", "white"},
			{"font-size", "9px"}}},
		{"box-shadow", []style{{"padding", "8px"}, {"background", "var(--c-card)"},
			{"border-radius", "6px"}, {"box-shadow", "0 4px 12px rgba(108,92,231,0.4)"},
			{"font-size", "10px"}, {"color", "var(--c-dim)"}}},
		{"gradient bg", []style{{"padding", "8px"},
			{"background", "linear-gradient(135deg, #6c5ce7, #fd79a8)"}, {"border-radius", "6px"},
			{"color", "white"}, {"font-size", "10px"}}},
		{"transform", []style{{"padding", "8px"}, {"background", "var(--c-card)"},
			{"border-radius", "6px"}, {"transform", "rotate(5deg) scale(1.05)"},
			{"font-size", "10px"}, {"color", "var(--c-dim)"}}},
		{"opacity: 0.5", []style{{"padding", "8px"}, {"background", "var(--c-accent)"},
			{"border-radius", "6px"}, {"opacity", "0.5"}, {"color", "white"}, {"font-size", "10px"}}},
		{"filter: blur", []style{{"padding", "8px"}, {"background", "var(--c-success)"},
			{"border-radius", "6px"}, {"filter", "blur(1px)"}, {"color", "white"},
			{"font-size", "10px"}}},
		{"text-shadow", []style{{"padding", "8px"}, {"background", "var(--c-card)"},
			{"border-radius", "6px"}, {"text-shadow", "0 0 8px var(--c-primary)"},
			{"font-size", "10px"}, {"color", "var(--c-secondary)"}}},
		{"outline", []style{{"padding", "8px"}, {"background", "var(--c-card)"},
			{"border-radius", "6px"}, {"outline", "2px solid var(--c-accent)"},
			{"outline-offset", "3px"}, {"font-size", "10px"}, {"color", "var(--c-dim)"}}},
	}
	for _, d := range demos {
		appendEl(row, makeEl(doc, "div", d.label, d.styles...))
	}
	appendEl(section, row)

	// --- 5. Complex nested layout ---
	heading(doc, section, "5. Complex Nested Layout (sidebar + grid + flex)")
	layout := makeEl(doc, "div", "", style{"display", "flex"}, style{"gap", "8px"},
		style{"height", "100px"}, style{"margin-bottom", "12px"})
	miniSidebar := makeEl(doc, "div", "", style{"width", "60px"}, style{"background", "var(--c-card)"},
		style{"border-radius", "6px"}, style{"padding", "6px"}, style{"display", "flex"},
		style{"flex-direction", "column"}, style{"gap", "4px"})
	for i := 0; i < 4; i++ {
		bg, fg := "var(--c-border)", "var(--c-dim)"
		if i == 0 {
			bg, fg = "var(--c-primary)", "white"
		}
		appendEl(miniSidebar, makeEl(doc, "div", fmt.Sprint(i+1), style{"background", bg},
			style{"border-radius", "4px"}, style{"padding", "2px"}, style{"text-align", "center"},
			style{"font-size", "10px"}, style{"color", fg}))
	}
	appendEl(layout, miniSidebar)
	contentGrid := makeEl(doc, "div", "", style{"flex", "1"}, style{"display", "grid"},
		style{"grid-template-columns", "1fr 1fr"}, style{"grid-template-rows", "1fr 1fr"},
		style{"gap", "4px"})
	cellColors := []string{"var(--c-danger)", "var(--c-success)", "var(--c-primary)", "var(--c-warning)"}
	cellLabels := []string{"Errors: 0", "Tests: 730", "Coverage: 91%", "Build: OK"}
	for i := range cellLabels {
		appendEl(contentGrid, makeEl(doc, "div", cellLabels[i], style{"background", cellColors[i]},
			style{"border-radius", "6px"}, style{"display", "flex"}, style{"align-items", "center"},
			style{"justify-content", "center"}, style{"font-size", "10px"}, style{"color", "white"},
			style{"font-weight", "700"}))
	}
	appendEl(layout, contentGrid)
	appendEl(section, layout)

	appendEl(container, section)

	// Update status bar
	if status, err := doc.GetElementByID("status-text"); err == nil && status != nil {
		_ = status.SetTextContent("Loaded: 16 sections, 307+ CSS properties tested")
		status.Close()
	}
}

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	_ = doc.RegisterBuiltinComponents() // before loadHTML in the original

	// Build the dynamic C++ section after HTML is loaded
	buildDynamicSection(doc)

	// Bind sidebar nav-item click events: nav-id -> section-id
	navMappings := [][2]string{
		{"nav-1", "sec-1"}, {"nav-2", "sec-2"}, {"nav-3", "sec-3"}, {"nav-4", "sec-4"},
		{"nav-5", "sec-5"}, {"nav-7", "sec-7"}, {"nav-8", "sec-8"}, {"nav-9", "sec-9"},
		{"nav-10", "sec-10"}, {"nav-11", "sec-11"}, {"nav-13", "sec-13"}, {"nav-14", "sec-14"},
		{"nav-15", "sec-15"}, {"nav-16", "sec-16"}, {"nav-cpp", "dynamic-section"},
	}
	var allNavItems []*kui.Element
	for _, m := range navMappings {
		if nav, err := doc.GetElementByID(m[0]); err == nil && nav != nil {
			ctx.Keep(nav)
			allNavItems = append(allNavItems, nav)
		}
	}
	for _, m := range navMappings {
		nav, err := doc.GetElementByID(m[0])
		if err != nil || nav == nil {
			continue
		}
		sec, err := doc.GetElementByID(m[1])
		if err != nil || sec == nil {
			nav.Close()
			continue
		}
		sec.Close()
		ctx.Keep(nav)
		secID := m[1]
		self := nav
		if _, err := nav.AddEventListener("click", func(*kui.Event) {
			// Update active state on all nav items
			for _, n := range allNavItems {
				_ = n.SetAttribute("class", "nav-item")
			}
			_ = self.SetAttribute("class", "nav-item active")
			// Scroll the section into view
			if section, err := ctx.Document.GetElementByID(secID); err == nil && section != nil {
				_ = section.ScrollIntoView(true)
				section.Close()
			}
		}); err != nil {
			return err
		}
	}
	fmt.Fprintln(os.Stderr, "[CSS Demo] Running — 16 CSS categories, 332 CSS properties")
	return nil
}
