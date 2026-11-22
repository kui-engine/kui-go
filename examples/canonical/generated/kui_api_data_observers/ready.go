// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_data_observers.go; edit that file, not this copy.
// API tour: observable values and bindings, Mutation / Resize / Intersection observers, the theme
// manager and the document router (examples/api_tours/data_observers). On start the host calls
// every method once, checks the results against what the C ABI documents and logs one row per
// check into #tour-log. Resize and intersection entries only arrive after rendered frames, so
// those checks run from a chain of timers and #tour-result becomes "PASS n/n" at the end (about
// 1 s). The panel buttons repeat the calls on the visible elements.
//
// Gaps in the Go binding (reported): no wrapper for kui_observable_retain, kui_observer_retain or
// kui_document_router_get_current_path, so those three are not called here (the C and Rust tours
// do); the current path comes from Router.Current().Path instead.
package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------
// check:    a documented behaviour; a failure makes the tour FAIL.
// knownBug: a documented behaviour the engine currently gets wrong (reported bug). A failure is
//           logged as "KNOWN BUG" and kept out of the verdict; once it passes it counts as ok.

type tour struct {
	doc         *kui.Document
	total       int
	failed      int
	known       int
	failedNames []string
}

func (t *tour) row(name, result, cell string) {
	log, err := t.doc.GetElementByID("tour-log")
	if err != nil || log == nil {
		return
	}
	defer log.Close()
	row, err := t.doc.CreateElement("tr")
	if err != nil || row == nil {
		return
	}
	defer row.Close()
	_ = row.SetAttribute("data-check", name)
	_ = row.SetAttribute("data-result", result)
	if result != "ok" {
		_ = row.AddClass("bad")
	}
	for _, text := range []string{name, cell} {
		if td, err := t.doc.CreateElement("td"); err == nil && td != nil {
			_ = td.SetTextContent(text)
			_ = row.AppendChild(td)
			td.Close()
		}
	}
	_ = log.AppendChild(row)
}

func (t *tour) check(name string, ok bool, detail string) {
	t.total++
	if ok {
		t.row(name, "ok", "ok")
		return
	}
	t.failed++
	t.failedNames = append(t.failedNames, name)
	fmt.Printf("tour check failed: %s: %s\n", name, detail)
	t.row(name, "fail", "FAIL: "+detail)
}

func (t *tour) knownBug(name string, ok bool, bug, detail string) {
	if ok {
		t.total++
		t.row(name, "ok", "ok")
		return
	}
	t.known++
	fmt.Printf("tour known bug: %s: %s\n", name, detail)
	t.row(name, "known", fmt.Sprintf("KNOWN BUG: %s (%s)", bug, detail))
}

func (t *tour) finish() {
	result, err := t.doc.GetElementByID("tour-result")
	if err != nil || result == nil {
		return
	}
	defer result.Close()
	text := fmt.Sprintf("PASS %d/%d", t.total, t.total)
	if t.failed > 0 {
		text = fmt.Sprintf("FAIL %d/%d: %s", t.failed, t.total, strings.Join(t.failedNames, ", "))
	}
	if t.known > 0 {
		text += fmt.Sprintf(" (known bugs: %d)", t.known)
	}
	_ = result.SetTextContent(text)
	_ = result.SetAttribute("data-total", fmt.Sprint(t.total))
	_ = result.SetAttribute("data-failed", fmt.Sprint(t.failed))
	_ = result.SetAttribute("data-known", fmt.Sprint(t.known))
	if t.failed == 0 {
		_ = result.AddClass("pass")
	} else {
		_ = result.AddClass("fail")
	}
}

// ---- Small readers -----------------------------------------------------------------------

func statusOf(err error) kui.Status {
	var kuiErr *kui.Error
	if errors.As(err, &kuiErr) {
		return kuiErr.Status
	}
	if err == nil {
		return kui.OK
	}
	return -1
}

func near(a, b, tolerance float32) bool { return math.Abs(float64(a-b)) <= float64(tolerance) }

func idOf(e *kui.Element) string {
	if e == nil {
		return ""
	}
	id, _ := e.ID()
	return id
}

func withID(doc *kui.Document, id string, fn func(*kui.Element)) {
	if element, _ := doc.GetElementByID(id); element != nil {
		defer element.Close()
		fn(element)
	}
}

func textOf(doc *kui.Document, id string) (text string) {
	withID(doc, id, func(e *kui.Element) { text, _ = e.TextContent() })
	return
}

func attrOf(doc *kui.Document, id, name string) (value string) {
	withID(doc, id, func(e *kui.Element) { value, _ = e.Attribute(name) })
	return
}

func computed(doc *kui.Document, id, property string) (value string) {
	withID(doc, id, func(e *kui.Element) { value, _ = e.ComputedStyleValue(property) })
	return
}

func rectOf(doc *kui.Document, id string) (rect kui.Rect) {
	withID(doc, id, func(e *kui.Element) { rect, _ = e.BoundingRect() })
	return
}

func setText(doc *kui.Document, id, text string) {
	withID(doc, id, func(e *kui.Element) { _ = e.SetTextContent(text) })
}

func setAttr(doc *kui.Document, id, name, value string) {
	withID(doc, id, func(e *kui.Element) { _ = e.SetAttribute(name, value) })
}

func setStyle(doc *kui.Document, id, property, value string) {
	withID(doc, id, func(e *kui.Element) { _ = e.SetStyleProperty(property, value, false) })
}

func setClass(doc *kui.Document, id, name string, on bool) {
	withID(doc, id, func(e *kui.Element) {
		if on {
			_ = e.AddClass(name)
		} else {
			_ = e.RemoveClass(name)
		}
	})
}

func scrollTo(doc *kui.Document, id string, top float32) {
	withID(doc, id, func(e *kui.Element) { _ = e.ScrollTo(0, top) })
}

// colourIs accepts a CSS colour as "#rrggbb" or "rgb(r, g, b)" / "rgba(r, g, b, a)".
func colourIs(value string, r, g, b int) bool {
	var got [3]int
	if strings.HasPrefix(value, "#") && len(value) == 7 {
		hex, err := strconv.ParseUint(value[1:], 16, 32)
		if err != nil {
			return false
		}
		got = [3]int{int(hex>>16) & 255, int(hex>>8) & 255, int(hex) & 255}
	} else {
		open, end := strings.IndexByte(value, '('), strings.IndexByte(value, ')')
		if !strings.HasPrefix(value, "rgb") || open < 0 || end < open {
			return false
		}
		parts := strings.Split(value[open+1:end], ",")
		if len(parts) < 3 {
			return false
		}
		for i := 0; i < 3; i++ {
			number, err := strconv.Atoi(strings.TrimSpace(parts[i]))
			if err != nil {
				return false
			}
			got[i] = number
		}
	}
	return got == [3]int{r, g, b}
}

func (t *tour) inputValue(id string) string {
	element, err := t.doc.GetElementByID(id)
	if err != nil || element == nil {
		return ""
	}
	defer element.Close()
	value, _ := element.Value()
	return value
}

func bindID(doc *kui.Document, obs *kui.Observable, id, property string, twoWay bool) (binding uint32, err error) {
	err = &kui.Error{Status: kui.ErrNotFound, Context: "bindID"}
	withID(doc, id, func(e *kui.Element) { binding, err = obs.Bind(e, property, twoWay) })
	return
}

// ---- Observable: values, subscriptions, one- and two-way bindings, flush ----------------

func checkObservable(t *tour) {
	doc := t.doc
	obs, err := kui.NewObservable("0")
	value := ""
	if obs != nil {
		value, _ = obs.Get()
	}
	t.check("ob.create_get", err == nil && obs != nil && value == "0", fmt.Sprintf("err %v, value '%s'", err, value))
	if obs == nil {
		return
	}
	defer obs.Close()

	if empty, err := kui.NewObservable(""); empty != nil {
		value, getErr := empty.Get()
		t.check("ob.create_empty", err == nil && getErr == nil && value == "", fmt.Sprintf("value '%s', err %v", value, getErr))
		_ = empty.Close()
	} else {
		t.check("ob.create_empty", false, fmt.Sprintf("err %v", err))
	}

	_ = obs.Set("1")
	value, _ = obs.Get()
	t.check("ob.set_get", value == "1", fmt.Sprintf("value '%s'", value))

	// Subscribers run synchronously, and only when the value actually changed.
	calls := 0
	sub, _ := obs.Subscribe(func() { calls++ })
	_ = obs.Set("2")
	t.check("ob.subscribe", calls == 1, fmt.Sprintf("calls after one change %d", calls))
	_ = obs.Set("2")
	t.check("ob.set_unchanged", calls == 1, fmt.Sprintf("calls after setting the same value %d", calls))
	err = obs.Unsubscribe(sub)
	_ = obs.Set("3")
	again := obs.Unsubscribe(sub)
	t.check("ob.unsubscribe", err == nil && calls == 1 && statusOf(again) == kui.ErrNotFound,
		fmt.Sprintf("err %v, calls %d, second unsubscribe %v", err, calls, again))

	// Binding writes the current value right away.
	textBinding, err := bindID(doc, obs, "s-text", "textContent", false)
	text := textOf(doc, "s-text")
	t.check("ob.bind_text", err == nil && textBinding != 0 && text == "3", fmt.Sprintf("err %v, id %d, text '%s'", err, textBinding, text))

	// Later changes are batched until the next frame or an explicit flush.
	_ = obs.Set("4")
	text = textOf(doc, "s-text")
	t.check("ob.batched", text == "3", fmt.Sprintf("text right after set '%s' (want the old '3')", text))
	err = doc.FlushBindings()
	text = textOf(doc, "s-text")
	t.check("ob.flush_bindings", err == nil && text == "4", fmt.Sprintf("err %v, text '%s'", err, text))

	// Any other property name is an attribute.
	attrBinding, err := bindID(doc, obs, "s-attr", "data-level", false)
	attr := attrOf(doc, "s-attr", "data-level")
	t.check("ob.bind_attr", err == nil && attr == "4", fmt.Sprintf("err %v, data-level '%s'", err, attr))

	// "style" is an attribute too: the value is the element's inline declaration.
	if style, _ := kui.NewObservable("background-color: #ff0000"); style != nil {
		_, err = bindID(doc, style, "s-style", "style", false)
		red := computed(doc, "s-style", "background-color")
		_ = style.Set("background-color: #00ff00")
		_ = doc.FlushBindings()
		green := computed(doc, "s-style", "background-color")
		t.check("ob.bind_style", err == nil && colourIs(red, 255, 0, 0) && colourIs(green, 0, 255, 0),
			fmt.Sprintf("err %v, computed '%s' then '%s'", err, red, green))
		defer style.Close()
	}

	// Two-way: "value" on an <input> (user input writes back; typed by the test case).
	if input, _ := kui.NewObservable("seed"); input != nil {
		_, err = bindID(doc, input, "s-input", "value", true)
		seeded := t.inputValue("s-input")
		_ = input.Set("next")
		_ = doc.FlushBindings()
		after := t.inputValue("s-input")
		t.check("ob.bind_two_way", err == nil && seeded == "seed" && after == "next",
			fmt.Sprintf("err %v, input value '%s' then '%s'", err, seeded, after))
		defer input.Close()
	}

	_, err = bindID(doc, obs, "s-text", "", false)
	t.check("ob.bind_invalid", statusOf(err) == kui.ErrInvalidArg, fmt.Sprintf("empty property -> %v", err))

	// Unbinding one binding leaves the others.
	err = obs.Unbind(textBinding)
	_ = obs.Set("5")
	_ = doc.FlushBindings()
	text, attr = textOf(doc, "s-text"), attrOf(doc, "s-attr", "data-level")
	again = obs.Unbind(textBinding)
	t.check("ob.unbind", err == nil && text == "4" && attr == "5" && statusOf(again) == kui.ErrNotFound,
		fmt.Sprintf("err %v, text '%s' (want 4), data-level '%s' (want 5)", err, text, attr))

	err = obs.UnbindAll()
	_ = obs.Set("6")
	_ = doc.FlushBindings()
	attr = attrOf(doc, "s-attr", "data-level")
	again = obs.Unbind(attrBinding)
	t.check("ob.unbind_all", err == nil && attr == "5" && statusOf(again) == kui.ErrNotFound,
		fmt.Sprintf("err %v, data-level '%s' (want 5)", err, attr))

	// Close drops the last reference (kui_observable_release); the handle is gone afterwards.
	err = obs.Close()
	t.check("ob.close", err == nil, fmt.Sprintf("Close: %v", err))
}

// ---- MutationObserver: child list, attributes, character data, filter, disconnect -------

type mutationLog struct {
	batches, records       int
	kind                   kui.MutationType
	target, added, removed string
	targetType, addedType  int
	attribute, oldValue    string
}

func (log *mutationLog) remember(records []kui.MutationRecord) {
	if len(records) == 0 {
		return
	}
	last := records[len(records)-1]
	log.batches++
	log.records += len(records)
	log.kind = last.Type
	log.target, log.targetType = idOf(last.Target), last.TargetNodeType
	log.added, log.addedType = idOf(last.Added), last.AddedNodeType
	log.removed = idOf(last.Removed)
	log.attribute, log.oldValue = last.AttributeName, last.OldValue
}

func checkMutation(t *tour) {
	doc := t.doc
	root, _ := doc.GetElementByID("s-mut")
	list, _ := doc.GetElementByID("s-list")
	if root == nil || list == nil {
		t.check("mut.page", false, "missing #s-mut / #s-list")
		return
	}
	defer root.Close()
	defer list.Close()

	log := &mutationLog{}
	obs, err := kui.NewMutationObserver(log.remember)
	t.check("mut.create", err == nil && obs != nil, fmt.Sprintf("err %v", err))
	if obs == nil {
		return
	}

	err = obs.ObserveMutations(root, kui.MutationOptions{})
	t.check("mut.observe_needs_kind", statusOf(err) == kui.ErrInvalidArg,
		fmt.Sprintf("no child_list / attributes / character_data -> %v", err))
	all := kui.MutationOptions{ChildList: true, Attributes: true, CharacterData: true, Subtree: true,
		AttributeOldValue: true, CharacterDataOldValue: true}
	err = obs.ObserveMutations(root, all)
	t.check("mut.observe", err == nil && log.batches == 0, fmt.Sprintf("err %v, batches %d", err, log.batches))

	// A new child in the subtree: one childList record, delivered synchronously.
	item, _ := doc.CreateElement("li")
	if item != nil {
		defer item.Close()
		_ = item.SetTextContent("one")
		_ = item.SetAttribute("id", "s-li-1")
		_ = list.AppendChild(item)
	}
	t.check("mut.child_list_add", log.records == 1 && log.kind == kui.MutationChildList && log.target == "s-list" &&
		log.added == "s-li-1" && log.addedType == 1,
		fmt.Sprintf("records %d, type %d, target '%s', added '%s' (type %d)", log.records, log.kind, log.target, log.added, log.addedType))

	if item != nil {
		_ = item.Remove()
	}
	t.check("mut.child_list_remove", log.records == 2 && log.kind == kui.MutationChildList && log.removed == "s-li-1" &&
		log.target == "s-list", fmt.Sprintf("records %d, type %d, removed '%s', target '%s'", log.records, log.kind, log.removed, log.target))

	setAttr(doc, "s-mitem", "data-state", "on")
	t.check("mut.attributes", log.records == 3 && log.kind == kui.MutationAttributes && log.target == "s-mitem" &&
		log.attribute == "data-state" && log.oldValue == "off",
		fmt.Sprintf("type %d, target '%s', name '%s', old '%s'", log.kind, log.target, log.attribute, log.oldValue))

	// Rewriting the lone text child changes its data in place: a characterData record whose
	// target is the parent element, with the text node's type.
	setText(doc, "s-mtext", "world")
	t.check("mut.character_data", log.kind == kui.MutationCharacterData && log.target == "s-mtext" && log.targetType == 3 &&
		log.oldValue == "hello", fmt.Sprintf("type %d, target '%s' (node type %d), old '%s'", log.kind, log.target, log.targetType, log.oldValue))

	// Observing again replaces the options: attributes only, filtered to data-state.
	_ = obs.ObserveMutations(root, kui.MutationOptions{Attributes: true, Subtree: true, AttributeFilter: []string{"data-state"}})
	before := log.records
	setAttr(doc, "s-mitem", "title", "ignored")
	setText(doc, "s-mtext", "again")
	filtered := log.records == before
	setAttr(doc, "s-mitem", "data-state", "off")
	t.check("mut.attribute_filter", filtered && log.records == before+1 && log.attribute == "data-state" && log.oldValue == "",
		fmt.Sprintf("filtered %v, records %d -> %d, name '%s', old '%s' (not requested)", filtered, before, log.records, log.attribute, log.oldValue))

	observeErr, unobserveErr := obs.Observe(root), obs.Unobserve(root)
	t.check("mut.generic_observe", statusOf(observeErr) == kui.ErrInvalidArg && statusOf(unobserveErr) == kui.ErrNotSupported,
		fmt.Sprintf("Observe %v (want INVALID_ARG), Unobserve %v (want NOT_SUPPORTED)", observeErr, unobserveErr))

	err = obs.Disconnect()
	before = log.records
	setAttr(doc, "s-mitem", "data-state", "on")
	t.check("mut.disconnect", err == nil && log.records == before, fmt.Sprintf("err %v, records %d -> %d", err, before, log.records))

	err = obs.Close()
	t.check("mut.close", err == nil, fmt.Sprintf("Close: %v", err))
}

// ---- Theme manager: built-ins, register / replace, apply, enumerate ---------------------

func hasName(names []string, name string) bool {
	for _, item := range names {
		if item == name {
			return true
		}
	}
	return false
}

func htmlCustomProperty(doc *kui.Document, name string) string {
	html, _ := doc.QuerySelector("html")
	if html == nil {
		return ""
	}
	defer html.Close()
	value, _ := html.StyleProperty(name)
	return value
}

var lightVars = map[string]string{"--tour-bg": "#f0fdf4", "--tour-fg": "#14532d", "--tour-accent": "#22c55e"}
var darkFirst = map[string]string{"--tour-bg": "#0f172a", "--tour-fg": "#e2e8f0", "--tour-accent": "#38bdf8"}
var darkVars = map[string]string{"--tour-bg": "#0f172a", "--tour-fg": "#e2e8f0", "--tour-accent": "#f59e0b"}

func checkTheme(t *tour) *kui.ThemeManager {
	doc := t.doc
	themes, err := kui.NewThemeManager()
	if themes == nil {
		t.check("th.builtin", false, fmt.Sprintf("NewThemeManager: %v", err))
		return nil
	}
	names, _ := themes.ThemeNames()
	t.check("th.builtin", err == nil && len(names) == 2 && hasName(names, "light") && hasName(names, "dark"),
		fmt.Sprintf("err %v, names %v", err, names))

	dark, err := themes.Theme("dark")
	allCustom := true
	for key := range dark {
		allCustom = allCustom && strings.HasPrefix(key, "--")
	}
	t.check("th.builtin_variables", err == nil && len(dark) > 0 && allCustom,
		fmt.Sprintf("dark has %d variables, all custom properties %v", len(dark), allCustom))

	active, _ := themes.ActiveTheme()
	t.check("th.active_empty", active == "", fmt.Sprintf("active before any apply '%s'", active))

	_ = themes.RegisterTheme("tour-light", lightVars)
	err = themes.RegisterTheme("tour-dark", darkFirst)
	names, _ = themes.ThemeNames()
	light, _ := themes.Theme("tour-light")
	t.check("th.register", err == nil && len(names) == 4 && hasName(names, "tour-dark") && len(light) == 3,
		fmt.Sprintf("err %v, names %v, tour-light variables %d", err, names, len(light)))
	t.check("th.get_variable", light["--tour-fg"] == "#14532d", fmt.Sprintf("tour-light --tour-fg '%s'", light["--tour-fg"]))

	// Registering an existing name replaces it.
	_ = themes.RegisterTheme("tour-dark", darkVars)
	names, _ = themes.ThemeNames()
	replaced, _ := themes.Theme("tour-dark")
	t.check("th.register_replace", len(names) == 4 && replaced["--tour-accent"] == "#f59e0b",
		fmt.Sprintf("count %d, --tour-accent '%s'", len(names), replaced["--tour-accent"]))

	_, err = themes.Theme("nope")
	t.check("th.not_found", statusOf(err) == kui.ErrNotFound, fmt.Sprintf("unknown theme -> %v", err))

	// Applying sets the variables inline on <html>; the card picks them up through var().
	err = themes.Apply("tour-dark", doc)
	active, _ = themes.ActiveTheme()
	bg := htmlCustomProperty(doc, "--tour-bg")
	t.check("th.apply", err == nil && active == "tour-dark" && bg == "#0f172a",
		fmt.Sprintf("err %v, active '%s', <html> --tour-bg '%s'", err, active, bg))
	css := computed(doc, "theme-card", "background-color")
	t.check("th.apply_computed", colourIs(css, 0x0f, 0x17, 0x2a), fmt.Sprintf("card background-color '%s'", css))

	err = themes.Apply("nope", doc)
	active, _ = themes.ActiveTheme()
	t.check("th.apply_unknown", statusOf(err) == kui.ErrNotFound && active == "tour-dark",
		fmt.Sprintf("err %v, active '%s'", err, active))

	_ = themes.Apply("tour-light", doc)
	css = computed(doc, "theme-card", "background-color")
	t.check("th.switch", colourIs(css, 0xf0, 0xfd, 0xf4), fmt.Sprintf("card background-color after tour-light '%s'", css))

	other, err := kui.NewThemeManager()
	closeErr := other.Close()
	t.check("th.release", err == nil && closeErr == nil, fmt.Sprintf("create %v, Close %v", err, closeErr))
	return themes
}

// ---- Router: routes, views, navigation, history, params, guards, callbacks --------------

type routerState struct {
	doc          *kui.Document
	navCalls     int
	navPath      string
	guardSeen    string
	blockLeave   bool
	adminLocked  bool
	navigatedID  uint32
	adminGuardID uint32
}

// makeView builds the view element for a route; the returned element is handed over to KUI.
func makeView(id, title string) kui.RouteViewFunc {
	return func(doc *kui.Document, route kui.RouteInfo) *kui.Element {
		view, err := doc.CreateElement("div")
		if err != nil || view == nil {
			return nil
		}
		_ = view.SetAttribute("id", id)
		_ = view.SetAttribute("class", "view")
		text := title
		if id == "view-user" {
			text = "User " + route.Params["id"]
		}
		_ = view.SetTextContent(text)
		return view
	}
}

func addRoutes(router *kui.Router) error {
	return errors.Join(
		router.AddRoute("/", "home", makeView("view-home", "Home")),
		router.AddRoute("/user/:id", "user", makeView("view-user", "User")),
		router.AddRoute("/about", "about", makeView("view-about", "About")),
		router.AddRoute("/admin", "admin", makeView("view-admin", "Admin area")))
}

func (r *routerState) onNavigated(route kui.RouteInfo) {
	r.navCalls++
	r.navPath = route.Path
	setText(r.doc, "route-nav", fmt.Sprintf("navigated: %s (pattern %s)", route.Path, route.Pattern))
}

func viewID(doc *kui.Document) (id string) {
	withID(doc, "route-view", func(view *kui.Element) {
		if child, _ := view.FirstChild(); child != nil {
			id = idOf(child)
			child.Close()
		}
	})
	return
}

func viewChildren(doc *kui.Document) (count int) {
	count = 99
	withID(doc, "route-view", func(view *kui.Element) { count, _ = view.ChildCount() })
	return
}

func currentPath(router *kui.Router) string {
	route, _ := router.Current()
	return route.Path
}

func checkRouter(t *tour) *routerState {
	doc := t.doc
	router := doc.Router()
	state := &routerState{doc: doc, adminLocked: true}
	view, _ := doc.GetElementByID("route-view")
	if view == nil {
		t.check("rt.page", false, "missing #route-view")
		return state
	}
	defer view.Close()

	index, size, _ := router.History()
	t.check("rt.history_initial", index == -1 && size == 0, fmt.Sprintf("index %d, size %d (want -1, 0)", index, size))

	err := addRoutes(router)
	count, _ := router.RouteCount()
	has, _ := router.HasRoute("/user/:id")
	t.check("rt.add_route", err == nil && count == 4 && has, fmt.Sprintf("err %v, count %d, has /user/:id %v", err, count, has))
	err = router.AddRoute("/about", "again", makeView("view-about", "About"))
	has, _ = router.HasRoute("/nope")
	t.check("rt.add_route_duplicate", statusOf(err) == kui.ErrInvalidArg && !has, fmt.Sprintf("duplicate -> %v, has /nope %v", err, has))

	_ = router.AddRoute("/tmp", "tmp", makeView("view-home", "Home"))
	err = router.RemoveRoute("/tmp")
	count, _ = router.RouteCount()
	again := router.RemoveRoute("/tmp")
	t.check("rt.remove_route", err == nil && count == 4 && statusOf(again) == kui.ErrNotFound, fmt.Sprintf("err %v, count %d, again %v", err, count, again))

	err = router.RegisterView(view)
	state.navigatedID, _ = router.OnNavigated(state.onNavigated)
	t.check("rt.register_view", err == nil && viewChildren(doc) == 0, fmt.Sprintf("err %v, children %d", err, viewChildren(doc)))

	navigated, _ := router.Push("/")
	t.check("rt.push", navigated && currentPath(router) == "/" && viewID(doc) == "view-home",
		fmt.Sprintf("navigated %v, path '%s', view '%s'", navigated, currentPath(router), viewID(doc)))
	t.check("rt.navigated_callback", state.navCalls == 1 && state.navPath == "/", fmt.Sprintf("calls %d, last path '%s'", state.navCalls, state.navPath))

	navigated, _ = router.Push("/user/42?tab=2")
	param, ok, _ := router.Param("id")
	text := textOf(doc, "view-user")
	t.check("rt.push_param", navigated && ok && param == "42" && viewID(doc) == "view-user" && text == "User 42",
		fmt.Sprintf("navigated %v, param id '%s', view '%s' '%s'", navigated, param, viewID(doc), text))
	route, err := router.Current()
	t.check("rt.get_current", err == nil && route.Path == "/user/42?tab=2" && route.Pattern == "/user/:id" && route.Query == "tab=2" &&
		route.QueryParams["tab"] == "2" && len(route.Params) == 1,
		fmt.Sprintf("path '%s', pattern '%s', query '%s', tab '%s', params %d", route.Path, route.Pattern, route.Query, route.QueryParams["tab"], len(route.Params)))
	_, ok, err = router.Param("nope")
	t.check("rt.get_param_missing", !ok && err == nil, fmt.Sprintf("Param(nope) ok %v, err %v", ok, err))
	index, size, _ = router.History()
	t.check("rt.history", index == 1 && size == 2, fmt.Sprintf("index %d, size %d (want 1, 2)", index, size))

	navigated, _ = router.Go(-1)
	index, size, _ = router.History()
	t.check("rt.go_back", navigated && currentPath(router) == "/" && index == 0 && size == 2,
		fmt.Sprintf("navigated %v, path '%s', index %d, size %d", navigated, currentPath(router), index, size))
	navigated, _ = router.Go(1)
	t.check("rt.go_forward", navigated && currentPath(router) == "/user/42?tab=2", fmt.Sprintf("navigated %v, path '%s'", navigated, currentPath(router)))
	navigated, _ = router.Go(5)
	t.check("rt.go_out_of_range", !navigated, fmt.Sprintf("Go(5) navigated %v", navigated))

	navigated, _ = router.Replace("/about")
	index, size, _ = router.History()
	t.check("rt.replace", navigated && currentPath(router) == "/about" && index == 1 && size == 2 && viewID(doc) == "view-about",
		fmt.Sprintf("navigated %v, path '%s', index %d, size %d, view '%s'", navigated, currentPath(router), index, size, viewID(doc)))

	// Guards: before-enter refuses /admin; before-leave refuses /about -> /.
	enter, _ := router.AddGuard(kui.RouteBeforeEnter, func(from, to kui.RouteInfo) bool {
		state.guardSeen = from.Path + " -> " + to.Path
		return to.Path != "/admin"
	})
	calls := state.navCalls
	navigated, _ = router.Push("/admin")
	t.check("rt.guard_enter_blocks", !navigated && currentPath(router) == "/about" && state.navCalls == calls && state.guardSeen == "/about -> /admin",
		fmt.Sprintf("navigated %v, path '%s', callbacks %d -> %d, guard saw '%s'", navigated, currentPath(router), calls, state.navCalls, state.guardSeen))

	leave, _ := router.AddGuard(kui.RouteBeforeLeave, func(from, to kui.RouteInfo) bool {
		return !(state.blockLeave && from.Path == "/about" && to.Path == "/")
	})
	state.blockLeave = true
	navigated, _ = router.Push("/")
	t.check("rt.guard_leave_blocks", !navigated && currentPath(router) == "/about", fmt.Sprintf("navigated %v, path '%s'", navigated, currentPath(router)))
	_ = router.RemoveGuard(leave)
	navigated, _ = router.Push("/")
	t.check("rt.remove_guard", navigated && currentPath(router) == "/", fmt.Sprintf("navigated %v, path '%s'", navigated, currentPath(router)))
	_ = router.RemoveGuard(enter)
	navigated, _ = router.Push("/admin")
	t.check("rt.guard_removed_allows", navigated && viewID(doc) == "view-admin", fmt.Sprintf("navigated %v, view '%s'", navigated, viewID(doc)))

	// A path no route matches still navigates and empties the view.
	navigated, _ = router.Push("/missing")
	route, _ = router.Current()
	t.check("rt.unmatched", navigated && route.Pattern == "" && viewChildren(doc) == 0,
		fmt.Sprintf("navigated %v, pattern '%s', children %d", navigated, route.Pattern, viewChildren(doc)))

	calls = state.navCalls
	err = router.RemoveNavigated(state.navigatedID)
	_, _ = router.Push("/about")
	t.check("rt.remove_navigated_callback", err == nil && state.navCalls == calls, fmt.Sprintf("err %v, callbacks %d -> %d", err, calls, state.navCalls))

	err = router.UnregisterView(view)
	_, _ = router.Push("/")
	children := viewChildren(doc)
	again = router.UnregisterView(view)
	t.check("rt.unregister_view", err == nil && children == 1 && statusOf(again) == kui.ErrNotFound,
		fmt.Sprintf("err %v, children after navigating unregistered %d (want the old 1), again %v", err, children, again))
	// Registering fills the view right away with the current route.
	_ = router.RegisterView(view)
	t.check("rt.register_fills", viewID(doc) == "view-home", fmt.Sprintf("view after register '%s'", viewID(doc)))

	err = router.Clear()
	count, _ = router.RouteCount()
	has, _ = router.HasRoute("/")
	t.check("rt.clear_routes", err == nil && count == 0 && !has, fmt.Sprintf("err %v, count %d, has / %v", err, count, has))

	// Set up again for the panel: routes, the /admin lock and the navigated callback.
	_ = addRoutes(router)
	state.adminGuardID, _ = router.AddGuard(kui.RouteBeforeEnter, func(from, to kui.RouteInfo) bool {
		return !(state.adminLocked && to.Path == "/admin")
	})
	state.navigatedID, _ = router.OnNavigated(state.onNavigated)
	_, _ = router.Push("/")
	return state
}

// ---- Resize & intersection observers: entries after rendered frames (timer chain) -------

type sizeLog struct {
	reports            int
	contentW, contentH float32
	borderW, borderH   float32
	contentX, contentY float32
}

type seenLog struct {
	reports      int
	intersecting bool
	ratio        float32
	root         kui.Rect
}

type frames struct {
	t                   *tour
	resize, inter, port *kui.Observer
	box, fluid          sizeLog
	a, b, panel         seenLog
	fluidFrozen         int
	fluidBefore         float32
}

func (f *frames) showResize() {
	setText(f.t.doc, "rs-out", fmt.Sprintf("box %.0fx%.0f (border %.0fx%.0f) | fluid %.0f",
		f.box.contentW, f.box.contentH, f.box.borderW, f.box.borderH, f.fluid.contentW))
}

func (f *frames) onResize(entries []kui.ResizeEntry) {
	for _, entry := range entries {
		var log *sizeLog
		switch idOf(entry.Target) {
		case "rs-box":
			log = &f.box
		case "rs-fluid":
			log = &f.fluid
		default:
			continue
		}
		log.reports++
		log.contentW, log.contentH = entry.ContentWidth, entry.ContentHeight
		log.borderW, log.borderH = entry.BorderWidth, entry.BorderHeight
		log.contentX, log.contentY = entry.ContentX, entry.ContentY
	}
	f.showResize()
}

func inOut(log seenLog) string {
	if log.intersecting {
		return "in"
	}
	return "out"
}

func (f *frames) onIntersection(entries []kui.IntersectionEntry) {
	for _, entry := range entries {
		var log *seenLog
		switch idOf(entry.Target) {
		case "io-a":
			log = &f.a
		case "io-b":
			log = &f.b
		default:
			continue
		}
		log.reports++
		log.intersecting, log.ratio, log.root = entry.Intersecting, entry.Ratio, entry.RootBounds
	}
	setText(f.t.doc, "io-out", fmt.Sprintf("A %s %.2f | B %s %.2f", inOut(f.a), f.a.ratio, inOut(f.b), f.b.ratio))
}

func (f *frames) onViewport(entries []kui.IntersectionEntry) {
	for _, entry := range entries {
		f.panel.reports++
		f.panel.intersecting = entry.Intersecting
	}
}

func observeID(doc *kui.Document, obs *kui.Observer, id string) {
	withID(doc, id, func(e *kui.Element) { _ = obs.Observe(e) })
}

func (f *frames) next(delay uint32, step func()) {
	if _, err := kui.SetTimeout(step, delay); err != nil {
		f.t.check("frames.timer", false, fmt.Sprintf("SetTimeout: %v", err))
		f.t.finish()
	}
}

func startFrameChecks(t *tour) *frames {
	doc := t.doc
	f := &frames{t: t}
	var resizeErr, interErr, portErr error
	f.resize, resizeErr = kui.NewResizeObserver(f.onResize)
	withID(doc, "io-scroller", func(root *kui.Element) {
		f.inter, interErr = kui.NewIntersectionObserver(&kui.IntersectionOptions{Root: root, Thresholds: []float32{0, 0.5, 1}}, f.onIntersection)
	})
	f.port, portErr = kui.NewIntersectionObserver(nil, f.onViewport)
	ok := resizeErr == nil && interErr == nil && portErr == nil && f.inter != nil
	t.check("frames.create", ok, fmt.Sprintf("resize %v, intersection %v, viewport intersection %v", resizeErr, interErr, portErr))
	if !ok {
		t.finish()
		return f
	}
	observeID(doc, f.resize, "rs-box")
	observeID(doc, f.resize, "rs-fluid")
	observeID(doc, f.inter, "io-a")
	observeID(doc, f.inter, "io-b")
	observeID(doc, f.port, "p-io")
	f.next(350, f.phase1)
	return f
}

func (f *frames) phase1() {
	t, doc := f.t, f.t.doc
	stage := rectOf(doc, "rs-stage")
	var client kui.Rect
	withID(doc, "io-scroller", func(e *kui.Element) { client, _ = e.ClientMetrics() })

	// content-box 160x40, padding 8, border 2: border box 180x60, content at 10,10.
	t.check("rs.first_report", f.box.reports >= 1 && near(f.box.contentW, 160, 0.5) && near(f.box.contentH, 40, 0.5) &&
		near(f.box.borderW, 180, 0.5) && near(f.box.borderH, 60, 0.5),
		fmt.Sprintf("reports %d, content %.1fx%.1f, border %.1fx%.1f", f.box.reports, f.box.contentW, f.box.contentH, f.box.borderW, f.box.borderH))
	t.knownBug("rs.content_offset", near(f.box.contentX, 10, 0.5) && near(f.box.contentY, 10, 0.5),
		"content_x / content_y give the padding offset (contentRect.x), not the documented offset inside the border box (padding + border)",
		fmt.Sprintf("offset %.1f,%.1f (want 10,10)", f.box.contentX, f.box.contentY))
	// 50% of the stage's content box (border 1 + padding 4 on each side).
	t.check("rs.percentage", f.fluid.reports >= 1 && near(f.fluid.contentW, (stage.Width-10)/2, 0.6),
		fmt.Sprintf("reports %d, content %.1f, stage %.1f", f.fluid.reports, f.fluid.contentW, stage.Width))
	t.check("io.first_report", f.a.reports >= 1 && f.b.reports >= 1, fmt.Sprintf("reports A %d, B %d", f.a.reports, f.b.reports))
	t.check("io.visible", f.a.intersecting && near(f.a.ratio, 1, 0.01), fmt.Sprintf("A intersecting %v, ratio %.2f", f.a.intersecting, f.a.ratio))
	t.check("io.hidden", !f.b.intersecting && near(f.b.ratio, 0, 0.01), fmt.Sprintf("B intersecting %v, ratio %.2f", f.b.intersecting, f.b.ratio))
	// W3C IntersectionObserver: an element root that clips its content intersects through its
	// padding area, i.e. the client rect.
	t.knownBug("io.root_bounds", near(f.a.root.Height, client.Height, 1) && near(f.a.root.Width, client.Width, 1),
		"root_bounds is the root's border box (border and scrollbar included), not its padding area",
		fmt.Sprintf("root %.1fx%.1f, scroller client %.1fx%.1f", f.a.root.Width, f.a.root.Height, client.Width, client.Height))
	t.check("io.viewport_root", f.panel.reports >= 1 && f.panel.intersecting,
		fmt.Sprintf("panel against the viewport: reports %d, intersecting %v", f.panel.reports, f.panel.intersecting))

	f.fluidBefore = f.fluid.contentW
	setStyle(doc, "rs-box", "width", "200px")
	setClass(doc, "rs-stage", "narrow", true)
	scrollTo(doc, "io-scroller", 10000)
	f.next(350, f.phase2)
}

func (f *frames) phase2() {
	t, doc := f.t, f.t.doc
	fluid := rectOf(doc, "rs-fluid")
	t.check("rs.resize_box", f.box.reports >= 2 && near(f.box.contentW, 200, 0.5) && near(f.box.borderW, 220, 0.5),
		fmt.Sprintf("reports %d, content %.1f, border %.1f (want 200, 220)", f.box.reports, f.box.contentW, f.box.borderW))
	t.check("rs.resize_layout", f.fluid.reports >= 2 && near(f.fluid.contentW, fluid.Width, 0.6) && f.fluid.contentW < f.fluidBefore-10,
		fmt.Sprintf("reports %d, content %.1f (laid out %.1f, before %.1f)", f.fluid.reports, f.fluid.contentW, fluid.Width, f.fluidBefore))
	t.check("io.scroll_in", f.b.intersecting && near(f.b.ratio, 1, 0.01), fmt.Sprintf("B intersecting %v, ratio %.2f", f.b.intersecting, f.b.ratio))
	t.check("io.scroll_out", !f.a.intersecting && near(f.a.ratio, 0, 0.01), fmt.Sprintf("A intersecting %v, ratio %.2f", f.a.intersecting, f.a.ratio))

	// Stop watching the fluid box, then change everything back.
	withID(doc, "rs-fluid", func(e *kui.Element) { _ = f.resize.Unobserve(e) })
	f.fluidFrozen = f.fluid.reports
	setClass(doc, "rs-stage", "narrow", false)
	setStyle(doc, "rs-box", "width", "160px")
	scrollTo(doc, "io-scroller", 0)
	f.next(350, f.phase3)
}

func (f *frames) phase3() {
	t := f.t
	t.check("rs.unobserve", f.fluid.reports == f.fluidFrozen, fmt.Sprintf("unobserved fluid reports %d -> %d", f.fluidFrozen, f.fluid.reports))
	t.check("rs.still_observing", near(f.box.contentW, 160, 0.5), fmt.Sprintf("box content width after the reset %.1f (want 160)", f.box.contentW))
	t.check("io.scroll_back", f.a.intersecting && near(f.a.ratio, 1, 0.01) && !f.b.intersecting,
		fmt.Sprintf("A %v %.2f, B %v", f.a.intersecting, f.a.ratio, f.b.intersecting))
	disconnectErr := f.port.Disconnect()
	closeErr := f.port.Close()
	t.check("io.disconnect_close", disconnectErr == nil && closeErr == nil, fmt.Sprintf("Disconnect %v, Close %v", disconnectErr, closeErr))
	observeID(t.doc, f.resize, "rs-fluid") // back on for the panel buttons
	t.finish()
}

// ---- Panel buttons -----------------------------------------------------------------------

type panel struct {
	doc                       *kui.Document
	count, level, badge, name *kui.Observable
	textBinding               uint32
	mut                       *kui.Observer
	mutTotal, mutNext         int
	rsWidth                   int
	themes                    *kui.ThemeManager
	router                    *routerState
}

func obsText(obs *kui.Observable) string {
	if obs == nil {
		return ""
	}
	value, _ := obs.Get()
	return value
}

func (p *panel) showObservable() {
	bound := "no"
	if p.textBinding != 0 {
		bound = "yes"
	}
	setText(p.doc, "ob-out", fmt.Sprintf("count=%s text=%s level=%s bound=%s", obsText(p.count), textOf(p.doc, "ob-text"), obsText(p.level), bound))
}

func (p *panel) setupObservable() {
	p.count, _ = kui.NewObservable("0")
	p.level, _ = kui.NewObservable("cool")
	p.badge, _ = kui.NewObservable("background-color: #dbeafe")
	p.name, _ = kui.NewObservable("KUI")
	p.textBinding, _ = bindID(p.doc, p.count, "ob-text", "textContent", false)
	_, _ = bindID(p.doc, p.level, "ob-badge", "data-level", false)
	_, _ = bindID(p.doc, p.badge, "ob-badge", "style", false)
	_, _ = bindID(p.doc, p.name, "ob-input", "value", true)
	echo := func() { setText(p.doc, "ob-echo", "echo: "+obsText(p.name)) }
	_, _ = p.name.Subscribe(echo)
	echo()
	p.showObservable()
}

func (p *panel) observeMut() {
	withID(p.doc, "mut-box", func(box *kui.Element) {
		_ = p.mut.ObserveMutations(box, kui.MutationOptions{ChildList: true, Attributes: true, CharacterData: true,
			Subtree: true, AttributeOldValue: true, CharacterDataOldValue: true})
	})
}

func (p *panel) onMutations(records []kui.MutationRecord) {
	types := map[kui.MutationType]string{kui.MutationChildList: "childList", kui.MutationAttributes: "attributes", kui.MutationCharacterData: "characterData"}
	last := records[len(records)-1]
	p.mutTotal += len(records)
	out := fmt.Sprintf("records=%d | %s #%s", p.mutTotal, types[last.Type], idOf(last.Target))
	if id := idOf(last.Added); id != "" {
		out += " +#" + id
	}
	if id := idOf(last.Removed); id != "" {
		out += " -#" + id
	}
	if last.Type == kui.MutationAttributes {
		out += " " + last.AttributeName
	}
	if last.Type != kui.MutationChildList {
		out += " old=" + last.OldValue
	}
	setText(p.doc, "mut-out", out)
}

func (p *panel) showTheme() {
	active, _ := p.themes.ActiveTheme()
	setText(p.doc, "theme-out", fmt.Sprintf("active=%s | --tour-bg=%s", active, htmlCustomProperty(p.doc, "--tour-bg")))
}

func (p *panel) showRoute(note string) {
	router := p.doc.Router()
	route, _ := router.Current()
	index, size, _ := router.History()
	id, ok, _ := router.Param("id")
	if !ok {
		id = "-"
	}
	locked := "no"
	if p.router.adminLocked {
		locked = "yes"
	}
	setText(p.doc, "route-out", fmt.Sprintf("%spath=%s pattern=%s | %d/%d | id=%s | locked=%s", note, route.Path, route.Pattern, index, size, id, locked))
}

func (p *panel) navigate(path string, replace bool) {
	router := p.doc.Router()
	var navigated bool
	if replace {
		navigated, _ = router.Replace(path)
	} else {
		navigated, _ = router.Push(path)
	}
	note := ""
	if !navigated {
		note = "blocked " + path + " | "
	}
	p.showRoute(note)
}

func (p *panel) goBy(delta int, failNote string) {
	navigated, _ := p.doc.Router().Go(delta)
	if navigated {
		failNote = ""
	}
	p.showRoute(failNote)
}

func (p *panel) buttons() map[string]func() {
	doc := p.doc
	return map[string]func(){
		"btn-ob-inc": func() {
			value, _ := strconv.Atoi(obsText(p.count))
			_ = p.count.Set(strconv.Itoa(value + 1))
			_ = doc.FlushBindings()
			p.showObservable()
		},
		"btn-ob-color": func() {
			hot := obsText(p.level) != "hot"
			if hot {
				_ = p.level.Set("hot")
				_ = p.badge.Set("background-color: #fee2e2")
			} else {
				_ = p.level.Set("cool")
				_ = p.badge.Set("background-color: #dbeafe")
			}
			_ = doc.FlushBindings()
			p.showObservable()
		},
		"btn-ob-unbind": func() {
			if p.textBinding != 0 {
				_ = p.count.Unbind(p.textBinding)
			}
			p.textBinding = 0
			p.showObservable()
		},
		"btn-ob-rebind": func() {
			if p.textBinding == 0 {
				p.textBinding, _ = bindID(doc, p.count, "ob-text", "textContent", false)
			}
			p.showObservable()
		},
		"btn-mut-add": func() {
			item, _ := doc.CreateElement("li")
			if item == nil {
				return
			}
			defer item.Close()
			_ = item.SetTextContent(fmt.Sprintf("item %d", p.mutNext))
			_ = item.SetAttribute("id", fmt.Sprintf("mut-li-%d", p.mutNext))
			p.mutNext++
			withID(doc, "mut-list", func(list *kui.Element) { _ = list.AppendChild(item) })
		},
		"btn-mut-remove": func() {
			withID(doc, "mut-list", func(list *kui.Element) {
				if last, _ := list.LastChild(); last != nil {
					_ = list.RemoveChild(last)
					last.Close()
				}
			})
		},
		"btn-mut-attr": func() {
			next := "on"
			if attrOf(doc, "mut-item", "data-state") == "on" {
				next = "off"
			}
			setAttr(doc, "mut-item", "data-state", next)
		},
		"btn-mut-text": func() {
			next := "world"
			if textOf(doc, "mut-text") != "hello" {
				next = "hello"
			}
			setText(doc, "mut-text", next)
		},
		"btn-mut-stop": func() {
			_ = p.mut.Disconnect()
			setText(doc, "mut-out", fmt.Sprintf("disconnected (records=%d)", p.mutTotal))
		},
		"btn-mut-start": func() {
			p.observeMut()
			setText(doc, "mut-out", fmt.Sprintf("observing (records=%d)", p.mutTotal))
		},
		"btn-rs-grow": func() {
			p.rsWidth += 40
			setStyle(doc, "rs-box", "width", fmt.Sprintf("%dpx", p.rsWidth))
		},
		"btn-rs-reset": func() {
			p.rsWidth = 160
			setStyle(doc, "rs-box", "width", "160px")
			setClass(doc, "rs-stage", "narrow", false)
		},
		"btn-rs-stage": func() {
			withID(doc, "rs-stage", func(stage *kui.Element) {
				if narrow, _ := stage.HasClass("narrow"); narrow {
					_ = stage.RemoveClass("narrow")
				} else {
					_ = stage.AddClass("narrow")
				}
			})
		},
		"btn-io-bottom": func() { scrollTo(doc, "io-scroller", 10000) },
		"btn-io-top":    func() { scrollTo(doc, "io-scroller", 0) },
		"btn-th-light": func() {
			_ = p.themes.Apply("tour-light", doc)
			p.showTheme()
		},
		"btn-th-dark": func() {
			_ = p.themes.Apply("tour-dark", doc)
			p.showTheme()
		},
		"btn-th-list": func() {
			names, _ := p.themes.ThemeNames()
			out := fmt.Sprintf("%d themes:", len(names))
			for _, name := range names {
				vars, _ := p.themes.Theme(name)
				out += fmt.Sprintf(" %s(%d)", name, len(vars))
			}
			setText(doc, "theme-out", out)
		},
		"btn-rt-home":  func() { p.navigate("/", false) },
		"btn-rt-user":  func() { p.navigate("/user/7", false) },
		"btn-rt-about": func() { p.navigate("/about", true) },
		"btn-rt-admin": func() { p.navigate("/admin", false) },
		"btn-rt-lock": func() {
			p.router.adminLocked = !p.router.adminLocked
			p.showRoute("")
		},
		"btn-rt-back": func() { p.goBy(-1, "no back | ") },
		"btn-rt-fwd":  func() { p.goBy(1, "no forward | ") },
	}
}

// on registers a click handler; the runner keeps the element so the listener survives.
func on(ctx *runner.Context, id string, fn func()) error {
	element, err := ctx.Document.GetElementByID(id)
	if err != nil || element == nil {
		return fmt.Errorf("没有 #%s: %v", id, err)
	}
	ctx.Keep(element)
	_, err = element.AddEventListener("click", func(*kui.Event) { fn() })
	return err
}

func ready(ctx *runner.Context) error {
	t := &tour{doc: ctx.Document}
	checkObservable(t)
	checkMutation(t)
	themes := checkTheme(t)
	router := checkRouter(t)
	f := startFrameChecks(t) // finishes the verdict from its timers

	p := &panel{doc: ctx.Document, rsWidth: 160, mutNext: 1, themes: themes, router: router}
	p.setupObservable()
	p.mut, _ = kui.NewMutationObserver(p.onMutations)
	if p.mut != nil {
		p.observeMut()
	}
	p.showTheme()
	p.showRoute("")
	ctx.OnClose(func() {
		for _, obs := range []*kui.Observable{p.count, p.level, p.badge, p.name} {
			_ = obs.Close()
		}
		for _, obs := range []*kui.Observer{p.mut, f.resize, f.inter} {
			_ = obs.Close()
		}
		_ = themes.Close()
	})

	for id, fn := range p.buttons() {
		if err := on(ctx, id, fn); err != nil {
			return err
		}
	}
	return nil
}
