// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_style_class.go; edit that file, not this copy.
// API tour: classes, inline style, computed style, stylesheets and animation state
// (examples/api_tours/style_class). On start the host calls every method once, checks the result
// against what the C ABI documents and logs one row per check into #tour-log. Animation state is
// only observable once frames run, so those checks finish from a 400 ms timer and #tour-result
// becomes "PASS n/n" then. The panel buttons repeat the calls by hand.
package main

import (
	"errors"
	"fmt"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

// ---- Self-check bookkeeping: count, log one <tr> per check, publish the verdict ----------
// check: a documented behaviour; a failure makes the tour FAIL. (No known engine bugs in this
// tour; see kui_api_dom_tree.go for the knownBug variant that logs a reported bug without failing.)

type tour struct {
	doc         *kui.Document
	total       int
	failed      int
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
	_ = result.SetTextContent(text)
	_ = result.SetAttribute("data-total", fmt.Sprint(t.total))
	_ = result.SetAttribute("data-failed", fmt.Sprint(t.failed))
	if t.failed == 0 {
		_ = result.AddClass("pass")
	} else {
		_ = result.AddClass("fail")
	}
}

// ---- Small readers -----------------------------------------------------------------------

func isNotFound(err error) bool {
	var kuiErr *kui.Error
	return errors.As(err, &kuiErr) && kuiErr.Status == kui.ErrNotFound
}

func classOf(e *kui.Element) string {
	if e == nil {
		return ""
	}
	value, _ := e.ClassName()
	return value
}

func has(e *kui.Element, token string) bool {
	if e == nil {
		return false
	}
	yes, _ := e.HasClass(token)
	return yes
}

func computed(e *kui.Element, property string) string {
	if e == nil {
		return ""
	}
	value, _ := e.ComputedStyleValue(property)
	return value
}

func computedByID(doc *kui.Document, id, property string) string {
	element, _ := doc.GetElementByID(id)
	if element == nil {
		return ""
	}
	defer element.Close()
	return computed(element, property)
}

func propertyNames(e *kui.Element) []string {
	if e == nil {
		return nil
	}
	names, _ := e.StylePropertyNames()
	return names
}

func styleText(e *kui.Element) string {
	if e == nil {
		return ""
	}
	text, _ := e.StyleText()
	return text
}

// describeInline lists "count: name=value[ !important], ..." for every inline property.
func describeInline(e *kui.Element) string {
	names := propertyNames(e)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		value, _ := e.StyleProperty(name)
		part := name + "=" + value
		if important, _ := e.StylePropertyImportant(name); important {
			part += " !important"
		}
		parts = append(parts, part)
	}
	return fmt.Sprintf("%d: %s", len(names), strings.Join(parts, ", "))
}

// pageCSS returns the page's own <style> text, read from <head>.
func pageCSS(doc *kui.Document) string {
	head, _ := doc.Head()
	if head == nil {
		return ""
	}
	defer head.Close()
	style, _ := head.QuerySelector("style")
	if style == nil {
		return ""
	}
	defer style.Close()
	text, _ := style.TextContent()
	return text
}

// ---- Class list: add / remove / toggle / contains, class name and id --------------------

func checkClassList(t *tour) {
	el, _ := t.doc.GetElementByID("sc-cls")
	if el == nil {
		t.check("class.page", false, "missing #sc-cls")
		return
	}
	defer el.Close()

	id, _ := el.ID()
	t.check("el.get_id", id == "sc-cls", fmt.Sprintf("id '%s'", id))
	t.check("el.get_class_name", classOf(el) == "a b", fmt.Sprintf("className '%s'", classOf(el)))

	_ = el.AddClass("c")
	t.check("el.class_list_add", classOf(el) == "a b c" && has(el, "c"), fmt.Sprintf("className '%s'", classOf(el)))
	_ = el.AddClass("a")
	t.check("el.class_list_add-duplicate", classOf(el) == "a b c", fmt.Sprintf("className '%s'", classOf(el)))

	_ = el.RemoveClass("b")
	t.check("el.class_list_remove", classOf(el) == "a c" && !has(el, "b"),
		fmt.Sprintf("className '%s', contains(b) %v", classOf(el), has(el, "b")))
	t.check("el.class_list_contains", has(el, "a") && !has(el, "zz"),
		fmt.Sprintf("contains(a) %v, contains(zz) %v", has(el, "a"), has(el, "zz")))

	_ = el.ToggleClass("d")
	on := has(el, "d")
	_ = el.ToggleClass("d")
	off := has(el, "d")
	t.check("el.class_list_toggle", on && !off, fmt.Sprintf("after toggles %v then %v", on, off))

	_ = el.SetClassName("tour-wide")
	t.check("el.set_class_name", classOf(el) == "tour-wide" && !has(el, "a"), fmt.Sprintf("className '%s'", classOf(el)))
	// The page stylesheet gives #sc-cls.tour-wide a width of 77px.
	t.check("el.class_drives_style", computed(el, "width") == "77px", fmt.Sprintf("computed width '%s'", computed(el, "width")))
}

// ---- Inline style: properties, !important, enumeration, whole-declaration text ----------

func checkInlineStyle(t *tour) {
	el, _ := t.doc.GetElementByID("sc-style")
	if el == nil {
		t.check("inline.page", false, "missing #sc-style")
		return
	}
	defer el.Close()

	_, err := el.StyleProperty("color")
	t.check("el.get_style_property-missing", isNotFound(err), fmt.Sprintf("err %v", err))

	// kui_c.h: named colours come back serialized as hex, lengths keep their unit.
	_ = el.SetStyleProperty("color", "red", false)
	color, err := el.StyleProperty("color")
	t.check("el.set_get_style_property", err == nil && color == "#ff0000", fmt.Sprintf("err %v, color '%s'", err, color))
	_ = el.SetStyleProperty("width", "100px", false)
	width, _ := el.StyleProperty("width")
	t.check("el.get_style_property-length", width == "100px", fmt.Sprintf("width '%s'", width))

	before, _ := el.StylePropertyImportant("color")
	_ = el.SetStyleProperty("background-color", "#00ff00", true)
	after, _ := el.StylePropertyImportant("background-color")
	t.check("el.style_property_is_important", !before && after, fmt.Sprintf("color %v, background-color %v", before, after))

	names := strings.Join(propertyNames(el), ",")
	t.check("el.get_style_property_count_name", len(propertyNames(el)) == 3 && strings.Contains(names, "color") &&
		strings.Contains(names, "width") && strings.Contains(names, "background-color"),
		fmt.Sprintf("count %d, names '%s'", len(propertyNames(el)), names))

	text := styleText(el)
	t.check("el.get_style_text", strings.Contains(text, "width") && strings.Contains(text, "100px") &&
		strings.Contains(text, "background-color") && strings.Contains(text, "!important"), fmt.Sprintf("cssText '%s'", text))

	// SetStyleText replaces the whole declaration.
	_ = el.SetStyleText("margin-left: 4px; padding-top: 2px")
	margin, _ := el.StyleProperty("margin-left")
	_, colorErr := el.StyleProperty("color")
	t.check("el.set_style_text", len(propertyNames(el)) == 2 && margin == "4px" && isNotFound(colorErr),
		fmt.Sprintf("count %d, margin-left '%s', color err %v", len(propertyNames(el)), margin, colorErr))

	_ = el.RemoveStyleProperty("margin-left")
	_, err = el.StyleProperty("margin-left")
	t.check("el.remove_style_property", isNotFound(err) && len(propertyNames(el)) == 1,
		fmt.Sprintf("err %v, count %d", err, len(propertyNames(el))))
}

// ---- Computed style: the cascade's resolved values --------------------------------------

func checkComputed(t *tour) {
	el, _ := t.doc.GetElementByID("cs-target")
	if el == nil {
		t.check("computed.page", false, "missing #cs-target")
		return
	}
	defer el.Close()
	display, width, font, padding := computed(el, "display"), computed(el, "width"), computed(el, "font-size"), computed(el, "padding-left")
	t.check("el.get_computed_style_value", display == "flex" && width == "240px" && font == "15px" && padding == "8px",
		fmt.Sprintf("display '%s', width '%s', font-size '%s', padding-left '%s'", display, width, font, padding))
	unknown, err := el.ComputedStyleValue("no-such-property")
	t.check("el.get_computed_style_value-unknown", err == nil && unknown == "", fmt.Sprintf("err %v, value '%s'", err, unknown))
	// Inline style feeds the cascade too.
	inline := computedByID(t.doc, "sc-style", "padding-top")
	t.check("el.get_computed_style_value-inline", inline == "2px", fmt.Sprintf("padding-top '%s'", inline))
}

// ---- Stylesheets & profile: add, clear every author sheet, restore the page's own --------

func checkStylesheets(t *tour) {
	doc := t.doc
	profile, err := doc.StyleProfile()
	t.check("doc.get_style_profile", err == nil && profile == kui.StyleNativeGUI, fmt.Sprintf("err %v, profile %d", err, profile))

	err = doc.AddStyleSheet("#sc-sheet { width: 123px; }")
	width := computedByID(doc, "sc-sheet", "width")
	t.check("doc.add_stylesheet", err == nil && width == "123px", fmt.Sprintf("err %v, width '%s'", err, width))

	css := pageCSS(doc)
	t.check("head.page_css", strings.Contains(css, "#cs-target"), fmt.Sprintf("page <style> text has %d bytes", len(css)))

	// Clearing drops the page sheet as well as ours: #cs-target falls back to display:block.
	err = doc.ClearStyleSheets()
	display, width := computedByID(doc, "cs-target", "display"), computedByID(doc, "sc-sheet", "width")
	t.check("doc.clear_stylesheets", err == nil && display == "block" && width != "123px",
		fmt.Sprintf("err %v, display '%s', width '%s'", err, display, width))

	err = doc.AddStyleSheet(css)
	display, width = computedByID(doc, "cs-target", "display"), computedByID(doc, "sc-sheet", "width")
	t.check("doc.add_stylesheet-restore", err == nil && display == "flex" && width != "123px",
		fmt.Sprintf("err %v, display '%s', width '%s'", err, display, width))
}

// ---- Animation state: needs running frames, so it is checked from a timer ---------------

func checkAnimation(t *tour) {
	doc := t.doc
	state := func(id string) (bool, bool) {
		element, _ := doc.GetElementByID(id)
		if element == nil {
			return false, false
		}
		defer element.Close()
		animating, transitioning, _ := element.AnimationState()
		return animating, transitioning
	}
	animating, transitioning := state("anim-box")
	t.check("el.get_animation_state-animation", animating, fmt.Sprintf("animating %v, transitioning %v", animating, transitioning))
	animating, transitioning = state("sc-trans")
	t.check("el.get_animation_state-transition", transitioning, fmt.Sprintf("animating %v, transitioning %v", animating, transitioning))
	animating, transitioning = state("cls-target")
	t.check("el.get_animation_state-idle", !animating && !transitioning, fmt.Sprintf("animating %v, transitioning %v", animating, transitioning))
	animations, transitions, _ := doc.AnimationCounts()
	t.check("doc.get_animation_counts", animations >= 1 && transitions >= 1, fmt.Sprintf("animations %d, transitions %d", animations, transitions))
	t.finish()
}

// startAnimationChecks starts the #anim-box animation (class .spinning) and a 2 s width transition
// on #sc-trans (inline style beats the 40px sheet value), then checks the state 400 ms later. It
// runs after the stylesheet checks only so that those checks see a quiet page; clearing the sheet
// that holds a running animation's @keyframes is safe (the engine cancels that animation:
// test_keyframes_stylesheet_removal).
func startAnimationChecks(t *tour) {
	if box, _ := t.doc.GetElementByID("anim-box"); box != nil {
		_ = box.AddClass("spinning")
		box.Close()
	}
	_ = computedByID(t.doc, "sc-trans", "width") // settle the old value first
	if bar, _ := t.doc.GetElementByID("sc-trans"); bar != nil {
		_ = bar.SetStyleProperty("width", "200px", false)
		bar.Close()
	}
	if _, err := kui.SetTimeout(func() { checkAnimation(t) }, 400); err != nil {
		t.check("animation.timer", false, fmt.Sprintf("SetTimeout: %v", err))
		t.finish()
	}
}

// ---- Panel buttons -----------------------------------------------------------------------

func bindClassButtons(ctx *runner.Context) error {
	doc := ctx.Document
	withTarget := func(fn func(el *kui.Element)) func() {
		return func() {
			el, _ := doc.GetElementByID("cls-target")
			if el == nil {
				return
			}
			defer el.Close()
			fn(el)
			id, _ := el.ID()
			accent := "no"
			if has(el, "accent") {
				accent = "yes"
			}
			setText(doc, "cls-out", fmt.Sprintf("#%s class=\"%s\" | accent=%s", id, classOf(el), accent))
		}
	}
	buttons := map[string]func(el *kui.Element){
		"btn-cls-toggle": func(el *kui.Element) { _ = el.ToggleClass("accent") },
		"btn-cls-add":    func(el *kui.Element) { _ = el.AddClass("big") },
		"btn-cls-remove": func(el *kui.Element) { _ = el.RemoveClass("big") },
		"btn-cls-reset":  func(el *kui.Element) { _ = el.SetClassName("card") },
	}
	for id, fn := range buttons {
		if err := on(ctx, id, withTarget(fn)); err != nil {
			return err
		}
	}
	return nil
}

func bindInlineButtons(ctx *runner.Context) error {
	doc := ctx.Document
	withTarget := func(fn func(el *kui.Element) string) func() {
		return func() {
			el, _ := doc.GetElementByID("st-target")
			if el == nil {
				return
			}
			defer el.Close()
			setText(doc, "st-out", fn(el))
		}
	}
	showText := func(el *kui.Element) string {
		if text := styleText(el); text != "" {
			return text
		}
		return "(no inline style)"
	}
	buttons := map[string]func(el *kui.Element) string{
		"btn-st-set": func(el *kui.Element) string {
			_ = el.SetStyleProperty("color", "#b91c1c", false)
			_ = el.SetStyleProperty("background-color", "#fee2e2", true)
			return showText(el)
		},
		"btn-st-text": func(el *kui.Element) string {
			_ = el.SetStyleText("background-color: #fef3c7; padding: 12px")
			return showText(el)
		},
		"btn-st-remove": func(el *kui.Element) string {
			_ = el.RemoveStyleProperty("background-color")
			return showText(el)
		},
		"btn-st-list": describeInline,
	}
	for id, fn := range buttons {
		if err := on(ctx, id, withTarget(fn)); err != nil {
			return err
		}
	}
	return nil
}

func bindComputedButtons(ctx *runner.Context) error {
	doc := ctx.Document
	read := func(toggleWide bool) func() {
		return func() {
			el, _ := doc.GetElementByID("cs-target")
			if el == nil {
				return
			}
			defer el.Close()
			if toggleWide {
				_ = el.ToggleClass("wide")
			}
			setText(doc, "cs-out", fmt.Sprintf("display=%s | width=%s | font-size=%s | color=%s",
				computed(el, "display"), computed(el, "width"), computed(el, "font-size"), computed(el, "color")))
		}
	}
	if err := on(ctx, "btn-cs-read", read(false)); err != nil {
		return err
	}
	return on(ctx, "btn-cs-wide", read(true))
}

func bindSheetButtons(ctx *runner.Context) error {
	doc := ctx.Document
	if err := on(ctx, "btn-sheet-add", func() {
		_ = doc.AddStyleSheet("#sheet-probe { background-color: #dcfce7; border-left-color: #16a34a; }")
		setText(doc, "sheet-out", "host sheet added")
	}); err != nil {
		return err
	}
	// Safe while #anim-box is animating: the animation whose @keyframes live in a cleared sheet is
	// cancelled, and the restored page sheet starts it again.
	if err := on(ctx, "btn-sheet-clear", func() {
		css := pageCSS(doc)
		_ = doc.ClearStyleSheets()
		if css == "" {
			setText(doc, "sheet-out", "cleared all sheets (page <style> not found)")
			return
		}
		_ = doc.AddStyleSheet(css)
		setText(doc, "sheet-out", "cleared all sheets, page sheet restored")
	}); err != nil {
		return err
	}
	return on(ctx, "btn-sheet-profile", func() {
		profile, _ := doc.StyleProfile()
		name := "native-gui"
		if profile == kui.StyleBrowserCompat {
			name = "browser-compat"
		}
		setText(doc, "sheet-out", "style profile = "+name)
	})
}

func bindAnimationButtons(ctx *runner.Context) error {
	doc := ctx.Document
	toggle := func(id, class string) func() {
		return func() {
			if el, _ := doc.GetElementByID(id); el != nil {
				_ = el.ToggleClass(class)
				el.Close()
			}
		}
	}
	if err := on(ctx, "btn-anim-toggle", toggle("anim-box", "spinning")); err != nil {
		return err
	}
	if err := on(ctx, "btn-trans-go", toggle("trans-box", "wide")); err != nil {
		return err
	}
	return on(ctx, "btn-anim-read", func() {
		yesNo := func(value bool) string {
			if value {
				return "yes"
			}
			return "no"
		}
		animating, transitioning := false, false
		if box, _ := doc.GetElementByID("anim-box"); box != nil {
			animating, _, _ = box.AnimationState()
			box.Close()
		}
		if bar, _ := doc.GetElementByID("trans-box"); bar != nil {
			_, transitioning, _ = bar.AnimationState()
			bar.Close()
		}
		animations, transitions, _ := doc.AnimationCounts()
		setText(doc, "anim-out", fmt.Sprintf("box: animating=%s | bar: transitioning=%s | doc: animations=%d transitions=%d",
			yesNo(animating), yesNo(transitioning), animations, transitions))
	})
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

func setText(doc *kui.Document, id, text string) {
	if element, err := doc.GetElementByID(id); err == nil && element != nil {
		_ = element.SetTextContent(text)
		element.Close()
	}
}

func ready(ctx *runner.Context) error {
	t := &tour{doc: ctx.Document}
	checkClassList(t)
	checkInlineStyle(t)
	checkComputed(t)
	checkStylesheets(t)
	startAnimationChecks(t) // finishes the verdict from its timer

	for _, bind := range []func(*runner.Context) error{
		bindClassButtons, bindInlineButtons, bindComputedButtons, bindSheetButtons, bindAnimationButtons,
	} {
		if err := bind(ctx); err != nil {
			return err
		}
	}
	return nil
}
