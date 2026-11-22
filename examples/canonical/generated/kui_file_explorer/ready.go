// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_file_explorer.go; edit that file, not this copy.
// Host logic of examples/file_explorer/main.cpp: lists a real directory into #file-grid (folders
// first, then by name; icon text and class by extension), builds the #breadcrumb from the path
// and a "Quick Access" #sidebar. Clicking a folder, a breadcrumb segment or a quick link
// navigates there; clicking a file does nothing.
//
// The original starts in (and every quick link points at) the current directory; the
// environment variable KUI_EXAMPLE_START_DIR overrides that start directory (test hook).
// Instead of one listener per created element, each of the three containers listens once and
// the clicked item carries its target directory in a data-path attribute.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

type entry struct {
	name  string
	isDir bool
}

func hasExt(ext string, list ...string) bool {
	for _, item := range list {
		if ext == item {
			return true
		}
	}
	return false
}

// iconFor is the original's getFileIcon / getIconClass (their extension lists differ: .svg and
// .rs only get an icon text, not a class).
func iconFor(e entry) (text, class string) {
	if e.isDir {
		return "[D]", "file-icon folder"
	}
	ext := strings.ToLower(filepath.Ext(e.name))
	if ext == e.name { // fs::path: ".hidden" has no extension
		ext = ""
	}
	docs := []string{".txt", ".md", ".doc", ".pdf"}
	switch {
	case hasExt(ext, ".png", ".jpg", ".bmp", ".gif", ".svg"):
		text = "[I]"
	case hasExt(ext, ".cpp", ".h", ".c", ".py", ".js", ".rs"):
		text = "[<>]"
	case hasExt(ext, docs...):
		text = "[T]"
	default:
		text = "[F]"
	}
	switch {
	case hasExt(ext, ".png", ".jpg", ".bmp", ".gif"):
		class = "file-icon image"
	case hasExt(ext, ".cpp", ".h", ".c", ".py", ".js"):
		class = "file-icon code"
	case hasExt(ext, docs...):
		class = "file-icon doc"
	default:
		class = "file-icon file"
	}
	return text, class
}

// listDirectory is fs::directory_iterator: errors give an empty list; links count as what they
// point at (directory_entry::is_directory follows them).
func listDirectory(dir string) []entry {
	items, _ := os.ReadDir(dir)
	entries := make([]entry, 0, len(items))
	for _, item := range items {
		isDir := item.IsDir()
		if item.Type()&os.ModeSymlink != 0 {
			info, err := os.Stat(filepath.Join(dir, item.Name()))
			isDir = err == nil && info.IsDir()
		}
		entries = append(entries, entry{item.Name(), isDir})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].isDir != entries[j].isDir {
			return entries[i].isDir
		}
		return entries[i].name < entries[j].name
	})
	return entries
}

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	start := os.Getenv("KUI_EXAMPLE_START_DIR")
	if start == "" {
		start, _ = os.Getwd()
	}
	if abs, err := filepath.Abs(start); err == nil {
		start = abs
	}

	// newEl creates <tag class=... data-path=...>text</tag>; empty strings are left out.
	newEl := func(tag, class, text, target string) *kui.Element {
		el, err := doc.CreateElement(tag)
		if err != nil || el == nil {
			return nil
		}
		if class != "" {
			_ = el.SetAttribute("class", class)
		}
		if text != "" {
			_ = el.SetTextContent(text)
		}
		if target != "" {
			_ = el.SetAttribute("data-path", target)
		}
		return el
	}
	appendTo := func(parent, child *kui.Element) {
		if child != nil {
			_ = parent.AppendChild(child)
			child.Close()
		}
	}
	// withCleared runs fn on the element with that id after removing all its children.
	withCleared := func(id string, fn func(*kui.Element)) {
		el, err := doc.GetElementByID(id)
		if err != nil || el == nil {
			return
		}
		defer el.Close()
		for {
			child, err := el.FirstChild()
			if err != nil || child == nil {
				break
			}
			_ = el.RemoveChild(child)
			child.Close()
		}
		fn(el)
	}

	navigateTo := func(path string) {
		withCleared("breadcrumb", func(bc *kui.Element) {
			var segments []string
			p := path
			for parent := filepath.Dir(p); parent != p; p, parent = parent, filepath.Dir(parent) {
				segments = append([]string{p}, segments...)
			}
			segments = append([]string{p}, segments...) // the root, shown as itself ("C:\", "/")
			for i, segment := range segments {
				name := filepath.Base(segment)
				if i == 0 {
					name = segment
				} else {
					appendTo(bc, newEl("span", "sep", "/", ""))
				}
				appendTo(bc, newEl("span", "seg", name, segment))
			}
		})
		withCleared("file-grid", func(grid *kui.Element) {
			entries := listDirectory(path)
			if len(entries) == 0 {
				appendTo(grid, newEl("div", "empty", "Empty folder", ""))
				return
			}
			for _, e := range entries {
				target := ""
				if e.isDir {
					target = filepath.Join(path, e.name)
				}
				item := newEl("div", "file-item", "", target)
				if item == nil {
					continue
				}
				text, class := iconFor(e)
				appendTo(item, newEl("div", class, text, ""))
				appendTo(item, newEl("div", "file-name", e.name, ""))
				appendTo(grid, item)
			}
		})
	}

	withCleared("sidebar", func(sidebar *kui.Element) {
		appendTo(sidebar, newEl("div", "sidebar-section", "Quick Access", ""))
		for _, link := range [][2]string{{"~", "Home"}, {"D", "Desktop"}, {"[>", "Documents"}} {
			item := newEl("div", "sidebar-item", "", start)
			if item == nil {
				continue
			}
			appendTo(item, newEl("span", "sidebar-icon", link[0], ""))
			appendTo(item, newEl("span", "", link[1], ""))
			appendTo(sidebar, item)
		}
	})

	onClick := func(event *kui.Event) {
		target, err := event.Target()
		if err != nil || target == nil {
			return
		}
		defer target.Close()
		item, err := target.Closest("[data-path]")
		if err != nil || item == nil {
			return
		}
		defer item.Close()
		if path, err := item.Attribute("data-path"); err == nil && path != "" {
			navigateTo(path)
		}
	}
	for _, id := range []string{"breadcrumb", "sidebar", "file-grid"} {
		el, err := doc.GetElementByID(id)
		if err != nil || el == nil {
			return fmt.Errorf("没有 #%s: %v", id, err)
		}
		ctx.Keep(el)
		if _, err := el.AddEventListener("click", onClick); err != nil {
			return err
		}
	}
	navigateTo(start)
	return nil
}
