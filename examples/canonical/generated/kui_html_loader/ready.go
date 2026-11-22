// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_html_loader.go; edit that file, not this copy.
// Host logic of examples/html_loader/main.cpp (kui_html_loader, the baseline variant): scans the
// sample folder, loads the first sample with the navigation bar injected after <body>, titles the
// window after it, and wires the bar: the sample <select> switches pages, "Open File..." loads any
// HTML file, "Inspect" toggles its own highlight. Each load gets a fresh script engine and the
// virtual_components sample gets its VirtualList / ListView / TreeView and scroll-unit toggle.
// The four kui_html_loader* ports are the same file except `variant`.
// prepare sets the variant's renderer presets (KUI_ENABLE_RETAINED_SURFACE /
// KUI_ENABLE_LAYER_PROMOTION, only when unset: the user's environment wins) before the runtime is
// loaded, handles --help / --list and parses the command line as the original (sample name
// substring, --file, --dev, --devtools [port], --devtools-port, --size, --frameless, --shadow,
// --radius, --sizeable, --show-after-load, --dpr).
//
// Differences from the original, all forced by the C ABI or the runner:
//   - Window shape: --radius, --sizeable and --show-after-load go into WindowProps (FrameRadius, Sizeable,
//     Hidden), so they take effect at creation like the original's WindowConfig; --show-after-load then
//     renders the first frame (ProcessFrame) and shows the window once the page is loaded.
//   - --dev and the script engine's console: SetConsole*; --no-token: SetDevToolsTokenAuth; --dpr:
//     SetRenderDevicePixelRatio. --devtools uses the window's DevTools server (StartDevTools) and is
//     skipped when it already runs.
//   - The runner loads the packaged "No samples found" page first; the port replaces it.
//   - Base path: Document.SetBasePath.
//   - Sample switching listens to the DOM "change" event and reads selectedIndex through Lua; the
//     page is replaced from a zero-delay timer, because a listener must not destroy the tree that
//     is dispatching it (the original's onChange runs after dispatch).
//   - "Inspect" only toggles the button style: there is no ABI for EventDispatcher's inspector
//     mode or inspect callback (DevTools on demand, launchBrowser, pushInspectElement).
//   - Open dialog filters carry no labels. The stack trace writer and the crash reporter are installed in
//     prepare, as the original does before Platform::init.
//   - Samples are also searched in <exe>/assets/native_samples, where the Go package puts them.
package main

/*
#include <stdlib.h>
#ifdef _WIN32
#include <windows.h>
// _putenv_s of the CRT kui.dll uses (ucrtbase.dll), whatever CRT this file is compiled against.
static void kuiex_crt_setenv(const char* key, const char* value) {
	typedef int(__cdecl * putenv_s_fn)(const char*, const char*);
	HMODULE crt = LoadLibraryA("ucrtbase.dll");
	putenv_s_fn putenv_s = crt ? (putenv_s_fn)GetProcAddress(crt, "_putenv_s") : NULL;
	if (putenv_s) putenv_s(key, value);
}
#else
static void kuiex_crt_setenv(const char* key, const char* value) { setenv(key, value, 0); }
#endif
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

const variant = 0 // 1 = retained surface, 2 = layer promotion, 3 = both

const (
	inspectOn  = "padding:4px 12px;font-size:12px;cursor:pointer;background:#0c4a6e;color:#38bdf8;border:1px solid #38bdf8;border-radius:4px;flex-shrink:0;"
	inspectOff = "padding:4px 12px;font-size:12px;cursor:pointer;background:#0f172a;color:#94a3b8;border:1px solid #475569;border-radius:4px;flex-shrink:0;"
)

type sample struct{ name, filename, path string }

// Command line (main() of the original), filled by prepare.
var (
	samples       []sample
	current       int
	fileToLoad    string
	devtools      bool
	devtoolsPort  = 9229
	sizeable      bool
	showAfterLoad bool
	radius        float32
	hasRadius     bool
	devMode       bool
	noToken       bool
	backendDPR    float32
)

type loader struct {
	ctx      *runner.Context
	samples  []sample
	current  int
	engine   *kui.ScriptEngine // recreated on every load, like setupScriptEngine
	elements []*kui.Element    // elements of the current page that hold listeners / components
	vlist    *kui.VirtualList
	list     *kui.ListView
	tree     *kui.TreeView
	inspect  bool
}

// applyR17EnvPreset: the user's value wins, so only unset variables are written.
// The variables must reach the C runtime KUI reads them from: on Windows kui.dll calls getenv of
// ucrtbase.dll, which keeps its own copy of the environment and never sees os.Setenv.
func applyVariantPreset() {
	setIfUnset := func(key string) {
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, "1")
			name, value := C.CString(key), C.CString("1")
			C.kuiex_crt_setenv(name, value)
			C.free(unsafe.Pointer(name))
			C.free(unsafe.Pointer(value))
		}
	}
	if variant&1 != 0 {
		setIfUnset("KUI_ENABLE_RETAINED_SURFACE")
	}
	if variant&2 != 0 {
		setIfUnset("KUI_ENABLE_LAYER_PROMOTION")
	}
}

const usage = `Usage: kui_html_loader [options] [sample]
  --file <path.html>   load an arbitrary HTML file
  --list               print available samples and exit
  --dev                enable dev console (errors/warnings to stderr)
  --devtools [port]    start DevTools server (default 9229)
  --devtools-port N    explicit DevTools server port
  --no-token           disable DevTools WebSocket token auth (testing)
Window shape (defaults: 1024x768 decorated window):
  --size WxH           initial client size, e.g. 1280x720
  --frameless          frameless window (no title bar / borders)
  --shadow             window shadow (frameless only)
  --radius <px>        rounded frameless window, radius in DIP
  --sizeable           edge-resizable frameless window
  --show-after-load    create hidden, load the page, render one frame, then show()
                       (the order real apps such as mini-erp-burn use)
  --dpr <x>            simulated backend device pixel ratio (measurement only)
`

func prepare(args []string, props *kui.WindowProps) (bool, int) {
	applyVariantPreset()
	exe, _ := os.Executable()
	base := filepath.Dir(exe)
	// <exe>/samples, ./samples, examples/html_loader/samples, then the package layout.
	for _, dir := range []string{filepath.Join(base, "samples"), "samples", "examples/html_loader/samples",
		filepath.Join(base, "assets", "native_samples")} {
		if samples = scan(dir); len(samples) != 0 {
			break
		}
	}
	if len(args) != 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Print(usage)
		return true, 0
	}
	if len(args) != 0 && args[0] == "--list" {
		fmt.Printf("Available samples (%d):\n", len(samples))
		for i, s := range samples {
			fmt.Printf("  [%d] %-30s %s\n", i, s.name, s.filename)
		}
		return true, 0
	}
	sampleToFind := ""
	for i := 0; i < len(args); i++ {
		arg, hasNext := args[i], i+1 < len(args)
		switch {
		case arg == "--size" && hasNext:
			i++
			var w, h uint
			if n, _ := fmt.Sscanf(args[i], "%dx%d", &w, &h); n == 2 && w >= 160 && h >= 120 && w <= 16384 && h <= 16384 {
				props.Width, props.Height = int(w), int(h)
			} else {
				fmt.Fprintf(os.Stderr, "[HTML Loader] Ignoring invalid --size '%s' (expected WxH)\n", args[i])
			}
		case arg == "--frameless":
			props.Frameless, props.NoDecorations = true, true
		case arg == "--shadow":
			props.FramelessShadow = true
		case arg == "--sizeable":
			sizeable, props.Sizeable = true, true
		case arg == "--show-after-load":
			showAfterLoad, props.Hidden = true, true
		case arg == "--radius" && hasNext:
			i++
			var r float64
			if _, err := fmt.Sscanf(args[i], "%g", &r); err == nil && r >= 0 && r <= 1024 {
				radius, hasRadius = float32(r), true
				props.FrameRadius = &radius
			}
		case arg == "--dpr" && hasNext:
			i++
			var dpr float64
			if _, err := fmt.Sscanf(args[i], "%g", &dpr); err == nil {
				backendDPR = float32(dpr)
			}
		case arg == "--dev":
			devMode = true
		case arg == "--devtools":
			devtools = true
			if hasNext && !strings.HasPrefix(args[i+1], "-") {
				if p, err := strconv.Atoi(args[i+1]); err == nil && p > 0 && p < 65536 {
					devtoolsPort = p
					i++
				}
			}
		case arg == "--devtools-port" && hasNext:
			i++
			if p, err := strconv.Atoi(args[i]); err == nil && p > 0 && p < 65536 {
				devtoolsPort, devtools = p, true
			}
		case arg == "--no-token":
			noToken = true
		case arg == "--file" && hasNext:
			i++
			fileToLoad = args[i]
		case !strings.HasPrefix(arg, "-"):
			sampleToFind = arg
		}
	}
	if devMode {
		_ = kui.SetConsoleEnabled(true)
		_ = kui.SetConsoleStderrOutput(true)
		_ = kui.SetConsoleMinLevel(kui.LogDebug)
		fmt.Fprintln(os.Stderr, "[DEV] Developer console enabled")
	}
	installCrashDiagnostics()
	if sampleToFind != "" {
		selectSample(sampleToFind)
	}
	return false, 0
}

// The stack trace writer and the crash reporter, as the original installs them before Platform::init.
func installCrashDiagnostics() {
	if installed, err := kui.InstallStackTrace(kui.StackTraceConfig{
		OutputMode: kui.StackTraceFile | kui.StackTraceStderr | kui.StackTraceMessageBox,
		AppName:    "kui_html_loader", WriteMiniDump: true, MaxFrames: 96,
	}); err != nil || !installed {
		fmt.Fprintln(os.Stderr, "[StackTrace] Warning: in-process crash handler install failed.")
	}
	version, _ := kui.RuntimeVersion()
	if installed, err := kui.InstallCrashReporter(kui.CrashReporterConfig{
		AppName:    "kui_html_loader",
		AppVersion: fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch),
	}); err != nil || !installed {
		fmt.Fprintln(os.Stderr, "[CrashReporter] Warning: crash reporting unavailable.")
	}
}

// Case-insensitive substring match on the file name; the first match wins.
func selectSample(name string) {
	needle := strings.ToLower(name)
	for i, s := range samples {
		if strings.Contains(strings.ToLower(s.filename), needle) {
			current = i
			return
		}
	}
	fmt.Fprintf(os.Stderr, "[HTML Loader] No sample matching '%s'. Use --list to see available.\n", name)
}

// The window shape (radius, sizeable, hidden) was set at creation through WindowProps.
func beforeLoad(ctx *runner.Context) error {
	return nil
}

func ready(ctx *runner.Context) error {
	l := &loader{ctx: ctx, samples: samples, current: current}
	if backendDPR > 0 {
		_ = ctx.Document.SetRenderDevicePixelRatio(backendDPR)
	}
	_ = kui.SetF12DevToolsEnabled(true)
	ctx.OnClose(l.close)
	if devtools {
		running, _, _, _ := ctx.Window.DevToolsState()
		started := false
		if !running {
			started, _ = ctx.Window.StartDevTools(uint16(devtoolsPort))
		}
		if running || started {
			if noToken {
				_ = ctx.Window.SetDevToolsTokenAuth(false)
			}
			if url, err := ctx.Window.DevToolsURL(); err == nil {
				fmt.Fprintf(os.Stderr, "[DevTools] Open %s in your browser to inspect\n", url)
			}
		}
	}
	// No samples and no --file: the runner's page is the original's "No samples found" page.
	if fileToLoad != "" {
		l.loadExternalFile(fileToLoad)
	} else if len(l.samples) != 0 {
		l.loadSample(l.current)
	}
	if showAfterLoad {
		_ = ctx.Window.ProcessFrame()
		_ = ctx.Window.Show()
	}
	return nil
}

func scan(dir string) []sample {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if entry.Type().IsRegular() && filepath.Ext(entry.Name()) == ".html" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	var found []sample
	for _, name := range names {
		found = append(found, sample{prettyName(name), name, filepath.Join(dir, name)})
	}
	return found
}

// "01_box_model.html" -> "Box Model"
func prettyName(filename string) string {
	s := strings.TrimSuffix(filename, filepath.Ext(filename))
	if len(s) > 3 && s[2] == '_' && s[0] >= '0' && s[0] <= '9' && s[1] >= '0' && s[1] <= '9' {
		s = s[3:]
	}
	var out strings.Builder
	capNext := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			out.WriteByte(' ')
			capNext = true
			continue
		}
		if capNext && c >= 'a' && c <= 'z' {
			c -= 'a' - 'A'
		}
		out.WriteByte(c)
		capNext = false
	}
	return out.String()
}

func (l *loader) navBar() string {
	var nav strings.Builder
	nav.WriteString(`
<div id="__nav" style="
    display: flex; align-items: center; gap: 12px;
    padding: 8px 16px; margin-bottom: 16px;
    background: #1e293b; border: 1px solid #334155;
    border-radius: 8px; font-family: 'Segoe UI', sans-serif;">
  <span style="font-size: 13px; font-weight: 700; color: #38bdf8;">KUI</span>
  <span style="font-size: 12px; color: #64748b;">HTML Loader</span>
  <select id="__sample_select" style="
      flex: 1; padding: 4px 8px; font-size: 13px;
      background: #0f172a; color: #e2e8f0;
      border: 1px solid #475569; border-radius: 4px;">
`)
	for i, s := range l.samples {
		selected := ""
		if i == l.current {
			selected = " selected"
		}
		fmt.Fprintf(&nav, "    <option value=\"%d\"%s>%s</option>\n", i, selected, s.filename)
	}
	nav.WriteString(`  </select>
  <button id="__btn_open" style="padding:4px 12px;font-size:12px;cursor:pointer;background:#0f172a;color:#38bdf8;border:1px solid #38bdf8;border-radius:4px;flex-shrink:0;">Open File...</button>
  <button id="__btn_inspect" style="padding:4px 12px;font-size:12px;cursor:pointer;background:#0f172a;color:#94a3b8;border:1px solid #475569;border-radius:4px;flex-shrink:0;">Inspect</button>
  <span id="__filename" style="font-size:11px;color:#475569;max-width:200px;overflow:hidden;">`)
	if l.current >= 0 && l.current < len(l.samples) {
		nav.WriteString(l.samples[l.current].filename)
	}
	nav.WriteString("</span>\n</div>\n")
	return nav.String()
}

// Inserts the bar after the <body ...> tag, or wraps the content when there is none.
func (l *loader) injectNav(content string) string {
	nav := l.navBar()
	body := strings.Index(strings.ToLower(content), "<body")
	if body < 0 {
		return "<html><body>\n" + nav + content + "\n</body></html>"
	}
	if end := strings.IndexByte(content[body:], '>'); end >= 0 {
		at := body + end + 1
		return content[:at] + "\n" + nav + content[at:]
	}
	return content
}

func (l *loader) currentEngine() *kui.ScriptEngine {
	if l.engine != nil {
		return l.engine
	}
	return l.ctx.Script
}

// Shared part of loadSample / loadExternalFile; html already carries the bar.
func (l *loader) loadPage(html, sourcePath, titleName string) {
	_ = l.ctx.Document.SetBasePath(filepath.Dir(sourcePath) + "/")
	l.releasePage()
	_ = l.ctx.Document.LoadHTML(html)
	_ = l.ctx.Window.SetTitle("KUI HTML Loader — " + titleName)
	l.wireNav()
	l.setupScriptEngine()
}

func (l *loader) loadSample(index int) {
	if index < 0 || index >= len(l.samples) {
		return
	}
	l.current = index
	s := l.samples[index]
	content, err := os.ReadFile(s.path)
	if err != nil || len(content) == 0 {
		fmt.Fprintf(os.Stderr, "[HTML Loader] Failed to read: %s\n", s.path)
		return
	}
	l.loadPage(l.injectNav(string(content)), s.path, s.name)
	fmt.Fprintf(os.Stderr, "[HTML Loader] Loaded: %s (%s)\n", s.name, s.filename)
	if strings.Contains(s.filename, "virtual_components") {
		l.setupVirtualComponents()
	}
}

func (l *loader) loadExternalFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil || len(content) == 0 {
		fmt.Fprintf(os.Stderr, "[HTML Loader] Failed to read: %s\n", path)
		return
	}
	l.current = -1 // the bar then shows no sample selected
	name := filepath.Base(path)
	l.loadPage(l.injectNav(string(content)), path, name)
	fmt.Fprintf(os.Stderr, "[HTML Loader] Loaded external file: %s\n", path)
	if span, _ := l.ctx.Document.GetElementByID("__filename"); span != nil {
		_ = span.SetTextContent(name)
		_ = span.Close()
	}
}

// setupScriptEngine: drop the previous engine, bind a new one and run the page's scripts.
// Loads only happen from timers or ready, never inside a running script.
func (l *loader) setupScriptEngine() {
	if l.engine != nil {
		_ = l.engine.UnbindDocument()
		_ = l.engine.Close()
		l.engine = nil
	} else {
		_ = l.ctx.Script.UnbindDocument()
	}
	// print() / console.log() output reaches DevTools and stderr (useful during development).
	_ = kui.SetConsoleEnabled(true)
	_ = kui.SetConsoleStderrOutput(true)
	_ = kui.SetConsoleMinLevel(kui.LogDebug)
	engine, err := kui.NewScriptEngine()
	if err != nil {
		return
	}
	l.engine = engine
	_ = engine.SetErrorHandler(func(message, source string, line int) {
		fmt.Fprintf(os.Stderr, "[Lua] Error in %s:%d — %s\n", source, line, message)
	})
	_ = engine.BindDocument(l.ctx.Document)
	_ = engine.ExecuteScripts()
	_ = l.ctx.Window.SetScriptEngine(engine) // re-attaches DevTools too
}

func (l *loader) listen(id, event string, fn func(*kui.Event)) {
	element, _ := l.ctx.Document.GetElementByID(id)
	if element == nil {
		return
	}
	l.elements = append(l.elements, element) // Close would remove the listener
	_, _ = element.AddEventListener(event, fn)
}

func (l *loader) wireNav() {
	l.listen("__sample_select", "change", func(*kui.Event) {
		result, err := l.currentEngine().Eval("return tostring(document:getElementById('__sample_select').selectedIndex)", "example")
		index := -1
		if err == nil && result.Success {
			fmt.Sscanf(result.Value.String, "%d", &index)
		}
		if index >= 0 && index < len(l.samples) && index != l.current {
			_, _ = kui.SetTimeout(func() { l.loadSample(index) }, 0)
		}
	})
	l.listen("__btn_open", "click", func(*kui.Event) {
		path, ok, err := kui.OpenFileDialog(kui.FileDialogOptions{
			Title: "Open HTML File", Filters: []string{"*.html;*.htm", "*.*"}})
		if err == nil && ok && path != "" {
			_, _ = kui.SetTimeout(func() { l.loadExternalFile(path) }, 0)
		}
	})
	l.listen("__btn_inspect", "click", func(*kui.Event) {
		l.inspect = !l.inspect
		button, _ := l.ctx.Document.GetElementByID("__btn_inspect")
		if button == nil {
			return
		}
		defer button.Close()
		style := inspectOff
		if l.inspect {
			style = inspectOn
		}
		_ = button.SetAttribute("style", style)
	})
}

// releasePage closes the components and element handles that belong to the page being replaced.
func (l *loader) releasePage() {
	if l.vlist != nil {
		_ = l.vlist.Close()
	}
	if l.list != nil {
		_ = l.list.Close()
	}
	if l.tree != nil {
		_ = l.tree.Close()
	}
	l.vlist, l.list, l.tree = nil, nil, nil
	for _, element := range l.elements {
		_ = element.Close()
	}
	l.elements = nil
}

func (l *loader) close() {
	l.releasePage()
	if l.engine != nil {
		_ = l.ctx.Window.SetScriptEngine(nil)
		_ = l.engine.UnbindDocument()
		_ = l.engine.Close()
		l.engine = nil
	}
}

func setStyles(element *kui.Element, properties ...string) {
	for i := 0; i+1 < len(properties); i += 2 {
		_ = element.SetStyleProperty(properties[i], properties[i+1], false)
	}
}

func choose(cond bool, yes, no string) string {
	if cond {
		return yes
	}
	return no
}

// setupVirtualComponents of the original.
func (l *loader) setupVirtualComponents() {
	const folder, file = "📁", "📄"
	doc := l.ctx.Document
	if container, _ := doc.GetElementByID("vlist-container"); container != nil {
		l.elements = append(l.elements, container)
		l.vlist, _ = kui.NewVirtualList(container, kui.VirtualListConfig{
			TotalItems: 100000, ItemHeight: 32, Overscan: 5,
			RenderItem: func(d *kui.Document, index uint32) *kui.Element {
				row, err := d.CreateElement("div")
				if err != nil {
					return nil
				}
				odd := index%2 == 1
				setStyles(row, "display", "flex", "align-items", "center", "padding", "0 12px", "font-size", "13px",
					"color", choose(odd, "#94a3b8", "#cbd5e1"), "background", choose(odd, "transparent", "rgba(255,255,255,0.02)"),
					"font-family", "'Segoe UI', sans-serif")
				if number, err := d.CreateElement("span"); err == nil {
					setStyles(number, "color", "#475569", "width", "60px", "font-size", "11px")
					_ = number.SetTextContent(fmt.Sprint(index))
					_ = row.AppendChild(number)
					_ = number.Close()
				}
				if text, err := d.CreateElement("span"); err == nil {
					_ = text.SetTextContent(fmt.Sprintf("Item #%d%s", index, choose(index%1000 == 0, "  ** milestone **", "")))
					_ = row.AppendChild(text)
					_ = text.Close()
				}
				return row
			},
		})
	}
	if container, _ := doc.GetElementByID("listview-container"); container != nil {
		l.elements = append(l.elements, container)
		icons := []string{"📄", "📁", "🔧", "📦", "🌐", "⚡"}
		l.list, _ = kui.NewListView(container, kui.ListViewConfig{
			ItemCount: 50000, ItemHeight: 36, Overscan: 5, SelectionMode: kui.SelectionSingle,
			ItemProvider: func(index uint32) kui.ListViewItem {
				return kui.ListViewItem{Icon: icons[index%6], Text: fmt.Sprintf("List item %d", index),
					Subtitle: fmt.Sprintf("Description for item %d", index), Disabled: index%50 == 49}
			},
			OnSelect: func(index uint32) { fmt.Fprintf(os.Stderr, "[ListView] Selected: %d\n", index) },
		})
	}
	if container, _ := doc.GetElementByID("treeview-container"); container != nil {
		l.elements = append(l.elements, container)
		if tree, err := kui.NewTreeView(); err == nil {
			l.tree = tree
			child := func(parent uint32, text, icon string) uint32 {
				index, _ := tree.AddChild(parent, text, icon)
				return index
			}
			src, _ := tree.AddNode("src", folder)
			comp := child(src, "components", folder)
			for _, name := range []string{"Button.tsx", "Input.tsx", "Modal.tsx", "Table.tsx", "Dropdown.tsx"} {
				child(comp, name, file)
			}
			hooks := child(src, "hooks", folder)
			for _, name := range []string{"useAuth.ts", "useQuery.ts", "useTheme.ts"} {
				child(hooks, name, file)
			}
			pages := child(src, "pages", folder)
			home := child(pages, "Home", folder)
			child(home, "index.tsx", file)
			child(home, "styles.css", file)
			dash := child(pages, "Dashboard", folder)
			for _, name := range []string{"index.tsx", "Chart.tsx", "Stats.tsx"} {
				child(dash, name, file)
			}
			settings := child(pages, "Settings", folder)
			child(settings, "Profile.tsx", file)
			child(settings, "Security.tsx", file)
			utils := child(src, "utils", folder)
			for _, name := range []string{"api.ts", "format.ts", "validate.ts"} {
				child(utils, name, file)
			}
			child(src, "App.tsx", file)
			child(src, "index.tsx", file)
			pub, _ := tree.AddNode("public", folder)
			child(pub, "index.html", file)
			child(pub, "favicon.ico", file)
			for _, name := range []string{"package.json", "tsconfig.json", "README.md"} {
				_, _ = tree.AddNode(name, file)
			}
			_ = tree.Expand(src)
			_ = tree.Setup(container, kui.TreeViewConfig{
				ItemHeight: 28, IndentWidth: 18, Overscan: 5,
				OnSelect: func(index uint32) { fmt.Fprintf(os.Stderr, "[TreeView] Selected node: %d\n", index) },
				OnToggle: func(index uint32, expanded bool) {
					fmt.Fprintf(os.Stderr, "[TreeView] Node %d %s\n", index, choose(expanded, "expanded", "collapsed"))
				},
			})
		}
	}
	applyUnit := func(unit kui.ScrollUnit) {
		if l.vlist != nil {
			_ = l.vlist.SetScrollUnit(unit)
		}
		if l.list != nil {
			_ = l.list.SetScrollUnit(unit)
		}
		if l.tree != nil {
			_ = l.tree.SetScrollUnit(unit)
		}
		l.toggleStyles(unit == kui.ScrollLine)
		fmt.Fprintf(os.Stderr, "[VirtualComponents] scrollUnit -> %s\n", choose(unit == kui.ScrollLine, "line", "pixel"))
	}
	l.listen("vc-scroll-px", "click", func(*kui.Event) { applyUnit(kui.ScrollPixel) })
	l.listen("vc-scroll-line", "click", func(*kui.Event) { applyUnit(kui.ScrollLine) })
	l.toggleStyles(false) // default visual state: Pixel mode
}

func (l *loader) toggleStyles(line bool) {
	for i, id := range []string{"vc-scroll-px", "vc-scroll-line"} {
		button, _ := l.ctx.Document.GetElementByID(id)
		if button == nil {
			continue
		}
		active := (i == 1) == line
		setStyles(button, "background", choose(active, "#0ea5e9", "#1e293b"), "color", choose(active, "#ffffff", "#94a3b8"),
			"border-color", choose(active, "#0ea5e9", "#334155"))
		_ = button.Close()
	}
}
