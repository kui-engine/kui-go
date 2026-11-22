//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_script_example.go; edit that file, not this copy.
// Host logic of examples/script_example/main.cpp: a browser for HTML files with embedded Lua.
// It scans script_html/ next to the executable (the original's POST_BUILD copy of
// tests/script/html; also --dir / --file / a positional path / --list), wraps each file's body
// and first <style> in a page with a nav bar (prev / counter / next / title / file name / reload)
// and a status bar, runs the file's <script> blocks in a fresh script engine, and shows the
// #result / #detail text in the status bar.
//
// Differences from the original:
//   - The original adds its Left / Right / F5 keydown listener to doc->body() before any page is
//     loaded, when the document has no body, so keyboard navigation never works. This port adds
//     the listener to the body of every page it loads (what the original means to do).
//   - --list prints the list and quits right away, but the window already exists by then (the
//     example host code runs after the runner opened the window).
//
// Like the original it ticks the script timers at 0 and 100 ms right after the scripts
// (ScriptEngine.TickTimers), so short timers land before #result is read.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

type htmlEntry struct {
	name     string // display name
	filename string // e.g. "01_basic_eval.html"
	path     string // full path
}

func prettyName(filename string) string {
	s := strings.TrimSuffix(filename, filepath.Ext(filename))
	// Remove leading number prefix like "01_"
	if len(s) > 3 && s[2] == '_' && unicode.IsDigit(rune(s[0])) && unicode.IsDigit(rune(s[1])) {
		s = s[3:]
	}
	var b strings.Builder
	capNext := true
	for _, c := range s {
		if c == '_' {
			b.WriteByte(' ')
			capNext = true
		} else {
			if capNext {
				c = unicode.ToUpper(c)
			}
			b.WriteRune(c)
			capNext = false
		}
	}
	return b.String()
}

func asciiLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// scanDirectory lists the .html files of dir in name order.
func scanDirectory(dir string) []htmlEntry {
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, item := range items {
		if item.Type().IsRegular() && filepath.Ext(item.Name()) == ".html" {
			files = append(files, item.Name())
		}
	}
	sort.Strings(files)
	entries := make([]htmlEntry, 0, len(files))
	for _, f := range files {
		entries = append(entries, htmlEntry{name: prettyName(f), filename: f, path: filepath.Join(dir, f)})
	}
	return entries
}

const statusHTML = `
<div id="__status" style="
    display: flex; align-items: center; gap: 8px;
    padding: 4px 12px;
    background: #0f172a; border-top: 1px solid #334155;
    font-family: 'Segoe UI', sans-serif; font-size: 11px; color: #64748b;">
  <span>Result:</span>
  <span id="__result_display" style="font-weight: 700; color: #94a3b8;">—</span>
  <span id="__detail_display" style="flex: 1; color: #475569; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;"></span>
</div>
`

const pageHead = `<!DOCTYPE html>
<html>
<head>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body { height: 100vh; display: flex; flex-direction: column; background: #0f172a; }
#__content_area {
    flex: 1; overflow: auto; padding: 16px;
    background: #0f172a; color: #e2e8f0;
    font-family: 'Segoe UI', sans-serif;
}
</style>
</head>
<body>
`

type browser struct {
	ctx     *runner.Context
	entries []htmlEntry
	current int
	engine  *kui.ScriptEngine // fresh engine per loaded entry
}

func (b *browser) navHTML() string {
	var name, filename string
	if b.current >= 0 && b.current < len(b.entries) {
		name, filename = b.entries[b.current].name, b.entries[b.current].filename
	}
	return `
<div id="__nav" style="
    display: flex; align-items: center; gap: 8px;
    padding: 6px 12px;
    background: #1e293b; border-bottom: 1px solid #334155;
    font-family: 'Segoe UI', sans-serif; position: relative;">
  <span style="font-size: 13px; font-weight: 700; color: #38bdf8;">KUI</span>
  <span style="font-size: 11px; color: #64748b;">Script</span>
  <div id="__btn_prev" style="
      padding: 2px 10px; font-size: 14px; color: #94a3b8;
      background: #0f172a; border: 1px solid #475569; border-radius: 4px;
      cursor: pointer;">&#9664;</div>
  <span id="__counter" style="font-size: 11px; color: #94a3b8; min-width: 50px; text-align: center;">` +
		fmt.Sprintf("%d / %d", b.current+1, len(b.entries)) + `</span>
  <div id="__btn_next" style="
      padding: 2px 10px; font-size: 14px; color: #94a3b8;
      background: #0f172a; border: 1px solid #475569; border-radius: 4px;
      cursor: pointer;">&#9654;</div>
  <span id="__title" style="
      flex: 1; font-size: 13px; font-weight: 600; color: #e2e8f0;
      overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">` + name + `</span>
  <span id="__filename" style="font-size: 10px; color: #475569;">` + filename + `</span>
  <div id="__btn_reload" style="
      padding: 2px 10px; font-size: 13px; color: #94a3b8;
      background: #0f172a; border: 1px solid #475569; border-radius: 4px;
      cursor: pointer;">&#8635;</div>
</div>
`
}

func (b *browser) text(id string) (string, bool) {
	el, err := b.ctx.Document.GetElementByID(id)
	if err != nil || el == nil {
		return "", false
	}
	defer el.Close()
	value, _ := el.TextContent()
	return value, true
}

// updateStatus copies #result / #detail into the status bar (the original's post-load block).
func (b *browser) updateStatus() {
	doc := b.ctx.Document
	display, err := doc.GetElementByID("__result_display")
	if err == nil && display != nil {
		if result, ok := b.text("result"); ok {
			_ = display.SetTextContent(result)
			color := "#ef4444"
			if result == "PASS" {
				color = "#22c55e"
			}
			_ = display.SetStyleProperty("color", color, false)
		} else {
			_ = display.SetTextContent("(no #result element)")
			_ = display.SetStyleProperty("color", "#f59e0b", false)
		}
		display.Close()
	}
	if detail, ok := b.text("detail"); ok {
		if el, err := doc.GetElementByID("__detail_display"); err == nil && el != nil {
			_ = el.SetTextContent(detail)
			el.Close()
		}
	}
}

func (b *browser) listen(el *kui.Element, event string, fn func(*kui.Event)) {
	if el == nil {
		return
	}
	b.ctx.Keep(el)
	_, _ = el.AddEventListener(event, fn)
}

// addKeyboardListener: Left / Right arrows to switch, F5 to reload.
func (b *browser) addKeyboardListener() {
	body, err := b.ctx.Document.Body()
	if err != nil || body == nil {
		return
	}
	b.listen(body, "keydown", func(e *kui.Event) {
		key, _ := e.Key()
		switch key {
		case "ArrowLeft", "Left":
			if b.current > 0 {
				b.load(b.current - 1)
			}
		case "ArrowRight", "Right":
			if b.current < len(b.entries)-1 {
				b.load(b.current + 1)
			}
		case "F5":
			b.load(b.current)
		}
	})
}

func (b *browser) wireNavEvents() {
	doc := b.ctx.Document
	if el, err := doc.GetElementByID("__btn_prev"); err == nil {
		b.listen(el, "click", func(*kui.Event) {
			if b.current > 0 {
				b.load(b.current - 1)
			}
		})
	}
	if el, err := doc.GetElementByID("__btn_next"); err == nil {
		b.listen(el, "click", func(*kui.Event) {
			if b.current < len(b.entries)-1 {
				b.load(b.current + 1)
			}
		})
	}
	if el, err := doc.GetElementByID("__btn_reload"); err == nil {
		b.listen(el, "click", func(*kui.Event) { b.load(b.current) })
	}
	b.addKeyboardListener()
}

// load builds the wrapper page for entry index and runs its scripts (the original's loadEntry).
func (b *browser) load(index int) {
	if index < 0 || index >= len(b.entries) {
		return
	}
	b.current = index
	entry := b.entries[index]
	raw, err := os.ReadFile(entry.path)
	if err != nil || len(raw) == 0 {
		fmt.Fprintf(os.Stderr, "[Script Example] Failed to read: %s\n", entry.path)
		return
	}
	content := string(raw)
	lower := asciiLower(content) // same byte offsets as content, like the original's ::tolower

	var page strings.Builder
	page.WriteString(pageHead)
	page.WriteString(b.navHTML())
	page.WriteString("<div id=\"__content_area\">\n")
	// Also extract the first <style> block
	styleStart, styleEnd := strings.Index(lower, "<style>"), strings.Index(lower, "</style>")
	if styleStart >= 0 && styleEnd >= 0 {
		page.WriteString(content[styleStart:styleEnd+8] + "\n")
	}
	// Inner body content (or the whole file when it has no body tags)
	bodyOpen, bodyClose := strings.Index(lower, "<body"), strings.Index(lower, "</body>")
	if bodyOpen >= 0 && bodyClose >= 0 {
		if closeTag := strings.IndexByte(content[bodyOpen:], '>'); closeTag >= 0 && bodyOpen+closeTag < bodyClose {
			page.WriteString(content[bodyOpen+closeTag+1 : bodyClose])
		}
	} else {
		page.WriteString(content)
	}
	page.WriteString("\n</div>\n")
	page.WriteString(statusHTML)
	page.WriteString("\n</body>\n</html>")

	doc := b.ctx.Document
	_ = doc.LoadHTML(page.String())
	_ = b.ctx.Window.SetTitle("KUI Script Example — " + entry.name)

	// Create / reset the script engine and run the <script> blocks of the original file
	if engine, err := kui.NewScriptEngine(); err == nil {
		if b.engine != nil {
			_ = b.engine.Close()
		}
		b.engine = engine
		_ = engine.BindDocument(doc)
		search, scriptIndex := 0, 0
		for {
			open := strings.Index(lower[search:], "<script>")
			if open < 0 {
				break
			}
			open += search
			end := strings.Index(lower[open:], "</script>")
			if end < 0 {
				break
			}
			end += open
			_, _ = engine.Eval(content[open+8:end], fmt.Sprintf("%s#%d", entry.filename, scriptIndex))
			scriptIndex++
			search = end + 9
		}
		// Tick timers: the original ticks 0 and 100 ms right after the scripts.
		_ = engine.TickTimers(0)
		_ = engine.TickTimers(100)
	}

	// Update status bar with the script result
	b.updateStatus()

	b.wireNavEvents()
	fmt.Fprintf(os.Stderr, "[Script Example] Loaded: %s\n", entry.filename)
}

func ready(ctx *runner.Context) error {
	_ = ctx.Document.RegisterBuiltinComponents()
	b := &browser{ctx: ctx}
	ctx.OnClose(func() {
		if b.engine != nil {
			_ = b.engine.Close()
		}
	})

	// Parse arguments (a leading runtime library path belongs to the runner)
	args := os.Args[1:]
	if len(args) != 0 && strings.EqualFold(filepath.Ext(args[0]), filepath.Ext(kui.RuntimeLibraryName)) {
		args = args[1:]
	}
	var fileToLoad, dirToScan string
	listMode := false
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--list":
			listMode = true
		case args[i] == "--file" && i+1 < len(args):
			i++
			fileToLoad = args[i]
		case args[i] == "--dir" && i+1 < len(args):
			i++
			dirToScan = args[i]
		case !strings.HasPrefix(args[i], "-"):
			// Positional: either a file or a directory
			if isFile(args[i]) {
				fileToLoad = args[i]
			} else {
				dirToScan = args[i]
			}
		}
	}

	// Default: scan script_html / tests/script/html relative to the exe, then the cwd
	if dirToScan == "" && fileToLoad == "" {
		defaultDirs := []string{"script_html", "assets/script_html", "tests/script/html", "../tests/script/html",
			"../../tests/script/html", "../../../tests/script/html", "../../../../tests/script/html"}
		if exe, err := os.Executable(); err == nil {
			for _, d := range defaultDirs {
				if candidate := filepath.Join(filepath.Dir(exe), d); isDir(candidate) {
					dirToScan = candidate
					break
				}
			}
		}
		for _, d := range defaultDirs {
			if dirToScan != "" {
				break
			}
			if isDir(d) {
				dirToScan = d
			}
		}
	}
	if dirToScan != "" {
		b.entries = scanDirectory(dirToScan)
	}

	if listMode {
		fmt.Printf("Available HTML files (%d):\n", len(b.entries))
		for i, e := range b.entries {
			fmt.Printf("  [%2d] %-30s %s\n", i, e.name, e.filename)
		}
		return kui.Quit(0)
	}

	switch {
	case fileToLoad != "":
		// loadSingleFile: a one-entry list
		name := filepath.Base(fileToLoad)
		b.entries = []htmlEntry{{name: prettyName(name), filename: name, path: fileToLoad}}
		b.load(0)
	case len(b.entries) > 0:
		b.load(0)
	default:
		// The packaged page is the original's "No HTML files found" page; keep it.
		b.addKeyboardListener()
	}
	return nil
}
