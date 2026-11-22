//go:build ignore

// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_dom_tree.go; edit that file, not this copy.
// API tour: host-side DOM tree operations (examples/api_tours/dom_tree). On start the host calls
// every lookup / traversal / tree-editing / HTML method once, checks the result against what the
// C ABI documents and logs one row per check into #tour-log; #tour-result ends as "PASS n/n".
// The panel buttons repeat the calls on the visible sandboxes.
package main

import (
	"errors"
	"fmt"
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

// ---- Small readers (nil-safe; every returned element must be closed) ---------------------

func closeAll(elements ...*kui.Element) {
	for _, element := range elements {
		if element != nil {
			element.Close()
		}
	}
}

func textOf(e *kui.Element) string {
	if e == nil {
		return ""
	}
	text, _ := e.TextContent()
	return text
}

func idOf(e *kui.Element) string {
	if e == nil {
		return ""
	}
	id, _ := e.ID()
	return id
}

func tagOf(e *kui.Element) string {
	if e == nil {
		return ""
	}
	tag, _ := e.TagName()
	return tag
}

func innerOf(e *kui.Element) string {
	if e == nil {
		return ""
	}
	html, _ := e.InnerHTML()
	return html
}

func childrenOf(e *kui.Element) int {
	if e == nil {
		return 0
	}
	count, _ := e.ChildCount()
	return count
}

// joinTexts joins the texts with ',' and closes the elements.
func joinTexts(items []*kui.Element) string {
	texts := make([]string, 0, len(items))
	for _, item := range items {
		texts = append(texts, textOf(item))
	}
	closeAll(items...)
	return strings.Join(texts, ",")
}

// ---- Document lookups: root / head / body, id, selectors, class and tag lists, hit test ----

func checkLookups(t *tour) {
	doc := t.doc
	root, _ := doc.Root()
	head, _ := doc.Head()
	body, _ := doc.Body()
	defer closeAll(root, head, body)

	t.check("doc.root", tagOf(root) == "html", fmt.Sprintf("root tag '%s'", tagOf(root)))

	var title *kui.Element
	if head != nil {
		title, _ = head.QuerySelector("title")
	}
	t.check("doc.head", tagOf(head) == "head" && textOf(title) == "KUI API — DOM tree",
		fmt.Sprintf("head tag '%s', <title> '%s'", tagOf(head), textOf(title)))
	closeAll(title)

	var parent *kui.Element
	if body != nil {
		parent, _ = body.Parent()
	}
	t.check("doc.body", body != nil && tagOf(parent) == "html",
		fmt.Sprintf("body parent tag '%s'", tagOf(parent)))
	closeAll(parent)

	box, _ := doc.GetElementByID("lookup-box")
	t.check("doc.get_element_by_id", idOf(box) == "lookup-box", fmt.Sprintf("id '%s'", idOf(box)))
	missing, err := doc.GetElementByID("no-such-id")
	t.check("doc.get_element_by_id-missing", missing == nil && err == nil, fmt.Sprintf("element %v, err %v", missing != nil, err))

	hot, _ := doc.QuerySelector("#lookup-box .hot")
	t.check("doc.query_selector", textOf(hot) == "beta", fmt.Sprintf("first .hot text '%s'", textOf(hot)))
	closeAll(hot)

	items, _ := doc.QuerySelectorAll(".chip")
	count, texts := len(items), joinTexts(items)
	t.check("doc.query_selector_all", count == 4 && texts == "alpha,beta,gamma,delta",
		fmt.Sprintf("count %d texts '%s'", count, texts))

	items, _ = doc.GetElementsByClassName("hot")
	count, texts = len(items), joinTexts(items)
	t.check("doc.get_elements_by_class_name", count == 2 && texts == "beta,delta",
		fmt.Sprintf("count %d texts '%s'", count, texts))
	// kui_c.h: several space-separated classes must all match.
	items, _ = doc.GetElementsByClassName("chip hot")
	count, texts = len(items), joinTexts(items)
	t.knownBug("doc.get_elements_by_class_name-multi", count == 2 && texts == "beta,delta",
		"a multi-class query matches nothing", fmt.Sprintf("count %d texts '%s'", count, texts))

	items, _ = doc.GetElementsByTagName("em")
	count, texts = len(items), joinTexts(items)
	t.check("doc.get_elements_by_tag_name", count == 1 && texts == "delta",
		fmt.Sprintf("count %d texts '%s'", count, texts))

	if box != nil {
		em, _ := box.QuerySelector("em")
		t.check("el.query_selector", textOf(em) == "delta", fmt.Sprintf("em text '%s'", textOf(em)))
		closeAll(em)
		items, _ = box.QuerySelectorAll("span")
		count, texts = len(items), joinTexts(items)
		t.check("el.query_selector_all", count == 3 && texts == "alpha,beta,gamma",
			fmt.Sprintf("count %d texts '%s'", count, texts))
	}
	closeAll(box)

	// Hit test the centre of #hit-target in viewport coordinates.
	target, _ := doc.GetElementByID("hit-target")
	var rect kui.Rect
	if target != nil {
		rect, _ = target.BoundingRect()
	}
	hit, _ := doc.ElementFromPoint(rect.X+rect.Width/2, rect.Y+rect.Height/2)
	t.check("doc.element_from_point", idOf(hit) == "hit-target",
		fmt.Sprintf("rect %.0f,%.0f %.0fx%.0f hit '%s'", rect.X, rect.Y, rect.Width, rect.Height, idOf(hit)))
	closeAll(hit, target)

	// An extra document reference (Clone = retain) is balanced by Close; the document stays usable.
	extra, err := doc.Clone()
	if err == nil {
		err = extra.Close()
	}
	again, _ := doc.GetElementByID("lookup-box")
	t.check("doc.retain", err == nil && again != nil, fmt.Sprintf("err %v", err))
	closeAll(again)
}

// ---- Traversal: parent, siblings, first / last child, closest and matches ---------------

func checkTraversal(t *tour) {
	doc := t.doc
	t2, _ := doc.GetElementByID("t2")
	t4, _ := doc.GetElementByID("t4")
	list, _ := doc.GetElementByID("trav-list")
	defer closeAll(t2, t4, list)
	if t2 == nil || t4 == nil || list == nil {
		t.check("traversal.page", false, "missing #t2 / #t4 / #trav-list")
		return
	}

	got, _ := t2.Parent()
	t.check("el.parent", idOf(got) == "trav-list", fmt.Sprintf("parent '%s'", idOf(got)))
	closeAll(got)

	got, _ = t2.NextSibling()
	t.check("el.next_sibling", idOf(got) == "t3", fmt.Sprintf("next '%s'", idOf(got)))
	closeAll(got)

	got, _ = t2.PreviousSibling()
	t.check("el.previous_sibling", idOf(got) == "t1", fmt.Sprintf("previous '%s'", idOf(got)))
	closeAll(got)

	got, err := t4.NextSibling()
	t.check("el.next_sibling-end", got == nil && err == nil, fmt.Sprintf("element %v, err %v", got != nil, err))
	closeAll(got)

	first, _ := list.FirstChild()
	last, _ := list.LastChild()
	t.check("el.first_last_child", idOf(first) == "t1" && idOf(last) == "t4" && childrenOf(list) == 4,
		fmt.Sprintf("first '%s' last '%s' count %d", idOf(first), idOf(last), childrenOf(list)))
	closeAll(first, last)

	got, _ = t2.Closest("ul")
	t.check("el.closest-ancestor", idOf(got) == "trav-list", fmt.Sprintf("closest(ul) '%s'", idOf(got)))
	closeAll(got)

	got, _ = t2.Closest("li")
	t.check("el.closest-inclusive", idOf(got) == "t2", fmt.Sprintf("closest(li) '%s'", idOf(got)))
	closeAll(got)

	got, err = t2.Closest("table")
	t.check("el.closest-none", got == nil && err == nil, fmt.Sprintf("element %v, err %v", got != nil, err))
	closeAll(got)

	yes, err := t2.Matches("li.mid")
	t.check("el.matches", yes && err == nil, fmt.Sprintf("matches(li.mid) = %v, err %v", yes, err))
	yes, err = t2.Matches("#t1")
	t.check("el.matches-false", !yes && err == nil, fmt.Sprintf("matches(#t1) = %v, err %v", yes, err))
}

// ---- Create & insert (in the hidden #scratch): create, append, insert_before, clone, replace --

func checkCreateInsert(t *tour) {
	doc := t.doc
	scratch, _ := doc.GetElementByID("scratch")
	defer closeAll(scratch)
	if scratch == nil {
		t.check("create.page", false, "missing #scratch")
		return
	}

	a, _ := doc.CreateElement("div")
	_ = a.SetAttribute("id", "sc-a")
	_ = a.SetAttribute("data-k", "v")
	_ = scratch.AppendChild(a)
	got, _ := doc.GetElementByID("sc-a")
	value := ""
	if got != nil {
		value, _ = got.Attribute("data-k")
	}
	t.check("el.create_append", childrenOf(scratch) == 1 && got != nil && value == "v",
		fmt.Sprintf("children %d, found %v, data-k '%s'", childrenOf(scratch), got != nil, value))
	closeAll(got)

	first, _ := doc.CreateElement("p")
	_ = first.SetAttribute("id", "sc-first")
	err := scratch.InsertBefore(first, a)
	got, _ = scratch.FirstChild()
	t.check("el.insert_before", err == nil && idOf(got) == "sc-first",
		fmt.Sprintf("err %v, first child '%s'", err, idOf(got)))
	closeAll(got, first)

	src, _ := doc.CreateElement("div")
	_ = src.SetAttribute("id", "sc-src")
	_ = src.SetInnerHTML("<b>bold</b> tail")
	_ = scratch.AppendChild(src)
	deep, _ := src.CloneNode(true)
	shallow, _ := src.CloneNode(false)
	closeAll(src)

	parent, parentErr := deep.Parent()
	bold, _ := deep.QuerySelector("b")
	t.check("el.clone-deep", textOf(deep) == "bold tail" && parent == nil && parentErr == nil && bold != nil,
		fmt.Sprintf("text '%s', parent %v, <b> %v", textOf(deep), parent != nil, bold != nil))
	closeAll(parent, bold)
	t.check("el.clone-shallow", shallow != nil && textOf(shallow) == "" && childrenOf(shallow) == 0 && tagOf(shallow) == "div",
		fmt.Sprintf("text '%s', children %d, tag '%s'", textOf(shallow), childrenOf(shallow), tagOf(shallow)))
	closeAll(shallow)

	// The deep copy (renamed, ids stay unique) takes #sc-a's place.
	_ = deep.SetAttribute("id", "sc-copy")
	err = scratch.ReplaceChild(deep, a)
	oldParent, _ := a.Parent()
	newParent, _ := deep.Parent()
	lookup, _ := doc.GetElementByID("sc-a")
	t.check("el.replace_child", err == nil && oldParent == nil && lookup == nil && idOf(newParent) == "scratch",
		fmt.Sprintf("err %v, old parent %v, lookup %v", err, oldParent != nil, lookup != nil))
	closeAll(oldParent, newParent, lookup, a, deep)
}

// ---- Remove, attributes and handle references (still in #scratch) ----------------------

func checkRemoveAttributes(t *tour) {
	doc := t.doc
	scratch, _ := doc.GetElementByID("scratch")
	first, _ := doc.GetElementByID("sc-first")
	src, _ := doc.GetElementByID("sc-src")
	copied, _ := doc.GetElementByID("sc-copy")
	defer closeAll(scratch, first, src, copied)
	if scratch == nil || first == nil || src == nil || copied == nil {
		t.check("remove.setup", false, "the create & insert checks left no #sc-first / #sc-src / #sc-copy")
		return
	}

	// A removed element leaves the document, but our handle keeps it alive and readable.
	err := first.Remove()
	got, _ := doc.GetElementByID("sc-first")
	t.check("el.remove", err == nil && got == nil && tagOf(first) == "p",
		fmt.Sprintf("err %v, still found %v, handle tag '%s'", err, got != nil, tagOf(first)))
	closeAll(got)

	before := childrenOf(scratch)
	err = scratch.RemoveChild(src)
	parent, _ := src.Parent()
	t.check("el.remove_child", err == nil && childrenOf(scratch)+1 == before && parent == nil,
		fmt.Sprintf("err %v, children %d -> %d", err, before, childrenOf(scratch)))
	closeAll(parent)

	_ = copied.SetAttribute("title", "hello")
	title, _ := copied.Attribute("title")
	t.check("el.set_get_attribute", title == "hello", fmt.Sprintf("title '%s'", title))
	has, _ := copied.HasAttribute("title")
	t.check("el.has_attribute", has, fmt.Sprintf("has_attribute = %v", has))
	err = copied.RemoveAttribute("title")
	has, _ = copied.HasAttribute("title")
	_, readErr := copied.Attribute("title")
	var kuiErr *kui.Error
	notFound := errors.As(readErr, &kuiErr) && kuiErr.Status == kui.ErrNotFound
	t.check("el.remove_attribute", err == nil && !has && notFound,
		fmt.Sprintf("err %v, has %v, get err %v", err, has, readErr))

	// Clone (= retain) adds a reference; closing it leaves our handle working.
	extra, err := copied.Clone()
	if err == nil {
		err = extra.Close()
	}
	t.check("el.retain", err == nil && idOf(copied) == "sc-copy", fmt.Sprintf("err %v, id '%s'", err, idOf(copied)))
}

// ---- HTML & text: inner / outer HTML, insert_adjacent_html, text nodes ------------------

// ---- Form control state: the current value / checked / selected index, not the content attribute --

func checkFormState(t *tour) {
	doc := t.doc
	scratch, _ := doc.GetElementByID("scratch")
	box, _ := doc.CreateElement("div")
	defer closeAll(scratch, box)
	if scratch == nil || box == nil {
		t.check("form.setup", false, "missing #scratch")
		return
	}
	_ = scratch.AppendChild(box)
	_ = box.SetInnerHTML(`<input id="sc-fv" type="text" value="initial"><select id="sc-fs"><option value="s">S</option><option value="m" selected>M</option><option value="l">L</option></select><input id="sc-fc" type="checkbox"><i id="sc-fi">x</i>`)
	input, _ := doc.GetElementByID("sc-fv")
	select_, _ := doc.GetElementByID("sc-fs")
	checkbox, _ := doc.GetElementByID("sc-fc")
	plain, _ := doc.GetElementByID("sc-fi")
	defer closeAll(input, select_, checkbox, plain)
	if input == nil || select_ == nil || checkbox == nil || plain == nil {
		t.check("form.setup", false, "form fixture missing")
		return
	}

	initial, _ := input.Value()
	_ = input.SetValue("Ada")
	current, _ := input.Value()
	attribute, _ := input.Attribute("value")
	t.check("el.set_get_value", initial == "initial" && current == "Ada" && attribute == "initial",
		fmt.Sprintf("initial '%s', current '%s', attribute '%s'", initial, current, attribute))

	before, _ := select_.SelectedIndex()
	_ = select_.SetSelectedIndex(2)
	after, _ := select_.SelectedIndex()
	selected, _ := select_.Value()
	t.check("el.set_get_selected_index", before == 1 && after == 2 && selected == "l",
		fmt.Sprintf("before %d, after %d, value '%s'", before, after, selected))

	unchecked, _ := checkbox.Checked()
	_ = checkbox.SetChecked(true)
	now, _ := checkbox.Checked()
	t.check("el.set_get_checked", !unchecked && now, fmt.Sprintf("unchecked first %v, now %v", !unchecked, now))

	// An element with no form state: Value / SelectedIndex are not supported.
	_, valueErr := plain.Value()
	_, indexErr := plain.SelectedIndex()
	t.check("el.form-not-supported", valueErr != nil && indexErr != nil,
		fmt.Sprintf("Value error %v, SelectedIndex error %v", valueErr, indexErr))
}

func checkHTMLText(t *tour) {
	doc := t.doc
	scratch, _ := doc.GetElementByID("scratch")
	box, _ := doc.CreateElement("div")
	defer closeAll(scratch, box)
	if scratch == nil || box == nil {
		t.check("html.setup", false, "missing #scratch")
		return
	}
	_ = scratch.AppendChild(box)

	_ = box.SetInnerHTML("<i>a</i><i>b</i>")
	t.check("el.set_inner_html", childrenOf(box) == 2, fmt.Sprintf("children %d", childrenOf(box)))
	t.check("el.get_inner_html", innerOf(box) == "<i>a</i><i>b</i>", fmt.Sprintf("innerHTML '%s'", innerOf(box)))

	_ = box.SetInnerHTML(`<p id="sc-o" class="x">hi</p>`)
	inner, _ := doc.GetElementByID("sc-o")
	outer := ""
	if inner != nil {
		outer, _ = inner.OuterHTML()
	}
	t.check("el.get_outer_html", strings.HasPrefix(outer, "<p") && strings.Contains(outer, `id="sc-o"`) &&
		strings.Contains(outer, `class="x"`) && strings.HasSuffix(outer, "hi</p>"), fmt.Sprintf("outerHTML '%s'", outer))
	closeAll(inner)

	_ = box.SetInnerHTML(`<span id="sc-t">T</span>`)
	inner, _ = doc.GetElementByID("sc-t")
	if inner != nil {
		_ = inner.InsertAdjacentHTML("beforebegin", "<i>1</i>")
		_ = inner.InsertAdjacentHTML("afterbegin", "<i>2</i>")
		_ = inner.InsertAdjacentHTML("beforeend", "<i>3</i>")
		_ = inner.InsertAdjacentHTML("afterend", "<i>4</i>")
	}
	t.check("el.insert_adjacent_html", textOf(box) == "12T34" && textOf(inner) == "2T3" && childrenOf(box) == 3,
		fmt.Sprintf("box text '%s', target text '%s', box children %d", textOf(box), textOf(inner), childrenOf(box)))
	closeAll(inner)

	span, _ := doc.CreateElement("span")
	defer closeAll(span)
	_ = span.SetTextContent("a")
	_ = span.AppendText("b")
	t.check("el.append_text", textOf(span) == "ab" && innerOf(span) == "ab" && childrenOf(span) == 0,
		fmt.Sprintf("text '%s', innerHTML '%s'", textOf(span), innerOf(span)))

	// set_text_content never parses markup: it becomes one text node.
	_ = span.SetTextContent("<b>x</b>")
	t.check("el.set_text_content", textOf(span) == "<b>x</b>" && childrenOf(span) == 0,
		fmt.Sprintf("text '%s', children %d", textOf(span), childrenOf(span)))
	// HTML serialization escapes '<' '>' '&' in text, so innerHTML parses back to the same text.
	t.check("el.get_inner_html-escapes-text", innerOf(span) == "&lt;b&gt;x&lt;/b&gt;",
		fmt.Sprintf("innerHTML '%s'", innerOf(span)))
}

// ---- Panel buttons -----------------------------------------------------------------------

func bindLookupButtons(ctx *runner.Context) error {
	doc := ctx.Document
	setOut := func(text string) { setText(doc, "lookup-out", text) }
	if err := on(ctx, "btn-lookup-count", func() {
		chips, _ := doc.QuerySelectorAll(".chip")
		hot, _ := doc.GetElementsByClassName("hot")
		em, _ := doc.GetElementsByTagName("em")
		setOut(fmt.Sprintf("chips=%d | hot=%d | em=%d", len(chips), len(hot), len(em)))
		closeAll(chips...)
		closeAll(hot...)
		closeAll(em...)
	}); err != nil {
		return err
	}
	if err := on(ctx, "btn-lookup-hit", func() {
		target, _ := doc.GetElementByID("hit-target")
		var rect kui.Rect
		if target != nil {
			rect, _ = target.BoundingRect()
		}
		hit, _ := doc.ElementFromPoint(rect.X+rect.Width/2, rect.Y+rect.Height/2)
		name := "(none)"
		if hit != nil {
			name = idOf(hit)
		}
		setOut("element_from_point -> #" + name)
		closeAll(hit, target)
	}); err != nil {
		return err
	}
	return on(ctx, "btn-lookup-root", func() {
		root, _ := doc.Root()
		head, _ := doc.Head()
		body, _ := doc.Body()
		setOut(fmt.Sprintf("root=%s | head=%s | body=%s", tagOf(root), tagOf(head), tagOf(body)))
		closeAll(root, head, body)
	})
}

func bindTraversalButtons(ctx *runner.Context) error {
	doc := ctx.Document
	// Moves the "cur" highlight to the next / previous item, wrapping at the ends.
	move := func(forward bool) {
		current, _ := doc.QuerySelector("#trav-list li.cur")
		list, _ := doc.GetElementByID("trav-list")
		defer closeAll(current, list)
		if current == nil || list == nil {
			return
		}
		var target *kui.Element
		if forward {
			if target, _ = current.NextSibling(); target == nil {
				target, _ = list.FirstChild()
			}
		} else {
			if target, _ = current.PreviousSibling(); target == nil {
				target, _ = list.LastChild()
			}
		}
		defer closeAll(target)
		_ = current.RemoveClass("cur")
		if target != nil {
			_ = target.AddClass("cur")
		}
		setText(doc, "trav-out", "#"+idOf(target))
	}
	if err := on(ctx, "btn-trav-next", func() { move(true) }); err != nil {
		return err
	}
	if err := on(ctx, "btn-trav-prev", func() { move(false) }); err != nil {
		return err
	}
	return on(ctx, "btn-trav-info", func() {
		current, _ := doc.QuerySelector("#trav-list li.cur")
		if current == nil {
			return
		}
		defer current.Close()
		parent, _ := current.Parent()
		panel, _ := current.Closest("section.panel")
		mid, _ := current.Matches(".mid")
		answer := "no"
		if mid {
			answer = "yes"
		}
		setText(doc, "trav-out", fmt.Sprintf("parent=#%s | closest(.panel)=#%s | matches(.mid)=%s", idOf(parent), idOf(panel), answer))
		closeAll(parent, panel)
	})
}

func bindBuildButtons(ctx *runner.Context) error {
	doc := ctx.Document
	serial := 0
	withBox := func(fn func(box *kui.Element)) func() {
		return func() {
			box, _ := doc.GetElementByID("build-box")
			if box == nil {
				return
			}
			defer box.Close()
			fn(box)
			setText(doc, "build-out", fmt.Sprintf("children=%d", childrenOf(box)))
		}
	}
	newItem := func(class, text string) *kui.Element {
		item, _ := doc.CreateElement("p")
		if item != nil {
			_ = item.SetClassName(class)
			_ = item.SetTextContent(text)
		}
		return item
	}
	if err := on(ctx, "btn-build-append", withBox(func(box *kui.Element) {
		serial++
		item := newItem("item", fmt.Sprintf("item %d", serial))
		_ = box.AppendChild(item)
		closeAll(item)
	})); err != nil {
		return err
	}
	if err := on(ctx, "btn-build-prepend", withBox(func(box *kui.Element) {
		serial++
		item := newItem("item", fmt.Sprintf("first %d", serial))
		first, _ := box.FirstChild()
		_ = box.InsertBefore(item, first) // nil first appends
		closeAll(item, first)
	})); err != nil {
		return err
	}
	if err := on(ctx, "btn-build-clone", withBox(func(box *kui.Element) {
		seed, _ := doc.GetElementByID("build-seed")
		if seed == nil {
			return
		}
		defer seed.Close()
		if copied, err := seed.CloneNode(true); err == nil && copied != nil {
			serial++
			_ = copied.SetAttribute("id", fmt.Sprintf("seed-copy-%d", serial))
			_ = box.AppendChild(copied)
			copied.Close()
		}
	})); err != nil {
		return err
	}
	return on(ctx, "btn-build-replace", withBox(func(box *kui.Element) {
		first, _ := box.FirstChild()
		if first == nil {
			return
		}
		serial++
		item := newItem("replaced", fmt.Sprintf("replaced %d", serial))
		_ = box.ReplaceChild(item, first)
		closeAll(item, first)
	}))
}

func bindRemoveButtons(ctx *runner.Context) error {
	doc := ctx.Document
	withBox := func(fn func(box *kui.Element)) func() {
		return func() {
			box, _ := doc.GetElementByID("rm-box")
			if box == nil {
				return
			}
			defer box.Close()
			fn(box)
			flag := "removed"
			if has, _ := box.HasAttribute("data-flag"); has {
				flag = "on"
			}
			setText(doc, "rm-out", fmt.Sprintf("left=%d | data-flag=%s", childrenOf(box), flag))
		}
	}
	if err := on(ctx, "btn-rm-last", withBox(func(box *kui.Element) {
		if last, _ := box.LastChild(); last != nil {
			_ = last.Remove()
			last.Close()
		}
	})); err != nil {
		return err
	}
	if err := on(ctx, "btn-rm-first", withBox(func(box *kui.Element) {
		if first, _ := box.FirstChild(); first != nil {
			_ = box.RemoveChild(first)
			first.Close()
		}
	})); err != nil {
		return err
	}
	if err := on(ctx, "btn-rm-attr", withBox(func(box *kui.Element) {
		if has, _ := box.HasAttribute("data-flag"); has {
			_ = box.RemoveAttribute("data-flag")
		} else {
			_ = box.SetAttribute("data-flag", "on")
		}
	})); err != nil {
		return err
	}
	return on(ctx, "btn-rm-retain", func() {
		box, _ := doc.GetElementByID("rm-box")
		if box == nil {
			return
		}
		defer box.Close()
		if extra, err := box.Clone(); err == nil {
			extra.Close()
		}
		setText(doc, "rm-out", "retained handle -> #"+idOf(box))
	})
}

func bindHTMLButtons(ctx *runner.Context) error {
	doc := ctx.Document
	withCore := func(fn func(core *kui.Element)) func() {
		return func() {
			core, _ := doc.GetElementByID("html-core")
			if core == nil {
				return
			}
			defer core.Close()
			fn(core)
		}
	}
	if err := on(ctx, "btn-html-reset", func() {
		box, _ := doc.GetElementByID("html-box")
		if box == nil {
			return
		}
		defer box.Close()
		_ = box.SetInnerHTML(`<span id="html-core">core</span>`)
		setText(doc, "html-out", innerOf(box))
	}); err != nil {
		return err
	}
	if err := on(ctx, "btn-html-adjacent", withCore(func(core *kui.Element) {
		_ = core.InsertAdjacentHTML("beforebegin", "<i>[</i>")
		_ = core.InsertAdjacentHTML("afterbegin", "<i>(</i>")
		_ = core.InsertAdjacentHTML("beforeend", "<i>)</i>")
		_ = core.InsertAdjacentHTML("afterend", "<i>]</i>")
		box, _ := doc.GetElementByID("html-box")
		setText(doc, "html-out", textOf(box))
		closeAll(box)
	})); err != nil {
		return err
	}
	if err := on(ctx, "btn-html-text", withCore(func(core *kui.Element) {
		_ = core.AppendText("+")
		setText(doc, "html-out", textOf(core))
	})); err != nil {
		return err
	}
	return on(ctx, "btn-html-outer", withCore(func(core *kui.Element) {
		outer, _ := core.OuterHTML()
		setText(doc, "html-out", outer)
	}))
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
	checkLookups(t)
	checkTraversal(t)
	checkCreateInsert(t)
	checkRemoveAttributes(t)
	checkHTMLText(t)
	checkFormState(t)
	t.finish()

	for _, bind := range []func(*runner.Context) error{
		bindLookupButtons, bindTraversalButtons, bindBuildButtons, bindRemoveButtons, bindHTMLButtons,
	} {
		if err := bind(ctx); err != nil {
			return err
		}
	}
	return nil
}
