// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/kui_api_widgets.go; edit that file, not this copy.
// API tour: native widgets (examples/api_tours/widgets). The host mounts two ListViews (single
// and multi selection), a multi-select TreeView and a VirtualList into the page, calls every
// selection / expand / scroll / data method once and checks the state read back (selected set,
// focus, node and visible row counts, rendered range); #tour-result ends as "PASS n/n". The
// panel buttons repeat the calls and print the state into each panel's readout.
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

// ---- Small readers -------------------------------------------------------------------------

func indices(values []uint32) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = fmt.Sprint(value)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func (t *tour) text(id string) string {
	el, err := t.doc.GetElementByID(id)
	if err != nil || el == nil {
		return ""
	}
	defer el.Close()
	text, _ := el.TextContent()
	return text
}

func (t *tour) has(id, needle string) bool { return strings.Contains(t.text(id), needle) }

func (t *tour) set(id string, format string, args ...any) {
	el, err := t.doc.GetElementByID(id)
	if err != nil || el == nil {
		return
	}
	defer el.Close()
	_ = el.SetTextContent(fmt.Sprintf(format, args...))
}

func isNotFound(err error) bool {
	var kuiErr *kui.Error
	return errors.As(err, &kuiErr) && kuiErr.Status == kui.ErrNotFound
}

// ==========================================================================================
// ---- ListView: selection modes, range selection, item count, scrolling -------------------

const (
	lvCount    = 200
	lvRow      = 30
	lvOverscan = 2
)

type lists struct {
	t             *tour
	single, multi *kui.ListView
	generation    int // refresh: appended to every row label once > 0
	provided      int // ItemProvider calls
	count         uint32
}

func lvSelected(list *kui.ListView) string {
	values, _ := list.Selected()
	return indices(values)
}

func lvIs(list *kui.ListView, index uint32) bool {
	selected, _ := list.IsSelected(index)
	return selected
}

func lvFocus(list *kui.ListView) int32 {
	index, _ := list.FocusedIndex()
	return index
}

func (l *lists) show() {
	state, _ := l.multi.State()
	l.t.set("lv-single-sel", "%s", lvSelected(l.single))
	l.t.set("lv-multi-sel", "%s", lvSelected(l.multi))
	l.t.set("lv-focus", "%d", lvFocus(l.multi))
	l.t.set("lv-start", "%d", state.VisibleStart)
	l.t.set("lv-end", "%d", state.VisibleEnd)
	l.t.set("lv-total", "%d", state.TotalItems)
}

func (l *lists) create(id string, mode kui.SelectionMode, name string) (*kui.ListView, error) {
	container, err := l.t.doc.GetElementByID(id)
	if err != nil || container == nil {
		return nil, fmt.Errorf("没有 #%s: %v", id, err)
	}
	defer container.Close()
	return kui.NewListView(container, kui.ListViewConfig{
		ItemCount: lvCount, ItemHeight: lvRow, Overscan: lvOverscan, SelectionMode: mode,
		ItemBg: "#ffffff", ItemHoverBg: "#f1f5f9", ItemSelectedBg: "#2563eb",
		ItemTextColor: "#0f172a", ItemSubColor: "#64748b",
		ItemProvider: func(index uint32) kui.ListViewItem {
			l.provided++
			if l.generation > 0 {
				return kui.ListViewItem{Text: fmt.Sprintf("Row %d r%d", index, l.generation)}
			}
			return kui.ListViewItem{Text: fmt.Sprintf("Row %d", index)}
		},
		// Clicks and keyboard selection on either list.
		OnSelect: func(index uint32) {
			l.t.set("lv-event", "on_select(%d) %s", index, name)
			l.show()
		},
		OnActivate: func(index uint32) { l.t.set("lv-event", "on_activate(%d) %s", index, name) },
	})
}

func (l *lists) checks() {
	t := l.t
	state, _ := l.multi.State()
	t.check("list.create", state.TotalItems == lvCount && state.ItemHeight == lvRow &&
		state.TotalHeight == 6000 && state.VisibleStart == 0 && state.VisibleEnd > 0 &&
		state.VisibleEnd <= lvCount && l.provided > 0 && t.has("lv-multi", "Row 0"),
		fmt.Sprintf("state %+v provided %d", state, l.provided))

	// Single selection: a second select replaces the first.
	_ = l.single.Select(2)
	_ = l.single.Select(4)
	got := lvSelected(l.single)
	t.check("list.select-single", got == "[4]" && !lvIs(l.single, 2) && lvIs(l.single, 4) &&
		lvFocus(l.single) == 4, fmt.Sprintf("selected %s focus %d", got, lvFocus(l.single)))

	// Multi selection: selections accumulate.
	_ = l.multi.Select(2)
	_ = l.multi.Select(4)
	got = lvSelected(l.multi)
	t.check("list.select-multi", got == "[2,4]" && lvIs(l.multi, 2) && lvIs(l.multi, 4) &&
		lvFocus(l.multi) == 4, fmt.Sprintf("selected %s focus %d", got, lvFocus(l.multi)))

	// kui_c.h: select_range adds rows [start, end).
	_ = l.multi.SelectRange(6, 9)
	got = lvSelected(l.multi)
	t.check("list.select_range", got == "[2,4,6,7,8]" && !lvIs(l.multi, 9), "selected "+got)

	_ = l.multi.Deselect(4)
	got = lvSelected(l.multi)
	t.check("list.deselect", got == "[2,6,7,8]" && !lvIs(l.multi, 4), "selected "+got)

	values, err := l.multi.Selected()
	t.check("list.get_selected-count", err == nil && len(values) == 4, fmt.Sprintf("%v %v", values, err))

	_ = l.single.ClearSelection()
	_ = l.multi.ClearSelection()
	values, _ = l.multi.Selected()
	t.check("list.clear_selection", len(values) == 0 && !lvIs(l.multi, 2) && !lvIs(l.single, 4),
		fmt.Sprintf("selected %v", values))

	_ = l.multi.ScrollToIndex(100)
	state, _ = l.multi.State()
	t.check("list.scroll_to_index", state.VisibleStart <= 100 && state.VisibleEnd > 100 &&
		state.VisibleStart > 0, fmt.Sprintf("rendered %d-%d", state.VisibleStart, state.VisibleEnd))

	_ = l.multi.ScrollToBottom()
	state, _ = l.multi.State()
	t.check("list.scroll_to_bottom", state.VisibleEnd == lvCount && state.VisibleStart > 100,
		fmt.Sprintf("rendered %d-%d", state.VisibleStart, state.VisibleEnd))

	_ = l.multi.ScrollToTop()
	state, _ = l.multi.State()
	t.check("list.scroll_to_top", state.VisibleStart == 0,
		fmt.Sprintf("rendered %d-%d", state.VisibleStart, state.VisibleEnd))

	// Shrinking the list drops selected rows that no longer exist.
	_ = l.multi.Select(10)
	_ = l.multi.Select(150)
	_ = l.multi.SetItemCount(100)
	state, _ = l.multi.State()
	got = lvSelected(l.multi)
	t.check("list.set_item_count", state.TotalItems == 100 && state.TotalHeight == 3000 && got == "[10]",
		fmt.Sprintf("total %d height %.0f selected %s", state.TotalItems, state.TotalHeight, got))
	_ = l.multi.SetItemCount(lvCount)
	_ = l.multi.ClearSelection()

	// refresh re-asks the provider for the rows in view.
	before := l.provided
	l.generation = 1
	_ = l.multi.Refresh()
	_ = l.single.Refresh()
	t.check("list.refresh", l.provided > before && t.has("lv-multi", "Row 0 r1") &&
		t.has("lv-single", "Row 0 r1"), fmt.Sprintf("provider calls %d -> %d", before, l.provided))
}

// Panel buttons: the same calls on the visible lists.
func (l *lists) buttons() map[string]func() {
	both := func(fn func(*kui.ListView)) func() {
		return func() {
			fn(l.single)
			fn(l.multi)
			l.show()
		}
	}
	return map[string]func(){
		"btn-lv-sel2":   both(func(v *kui.ListView) { _ = v.Select(2) }),
		"btn-lv-sel4":   both(func(v *kui.ListView) { _ = v.Select(4) }),
		"btn-lv-range":  func() { _ = l.multi.SelectRange(6, 9); l.show() },
		"btn-lv-desel":  both(func(v *kui.ListView) { _ = v.Deselect(2) }),
		"btn-lv-clear":  both(func(v *kui.ListView) { _ = v.ClearSelection() }),
		"btn-lv-goto":   both(func(v *kui.ListView) { _ = v.ScrollToIndex(100) }),
		"btn-lv-top":    both(func(v *kui.ListView) { _ = v.ScrollToTop() }),
		"btn-lv-bottom": both(func(v *kui.ListView) { _ = v.ScrollToBottom() }),
		"btn-lv-count": func() {
			if l.count == lvCount {
				l.count = 50
			} else {
				l.count = lvCount
			}
			_ = l.single.SetItemCount(l.count)
			_ = l.multi.SetItemCount(l.count)
			l.show()
		},
		"btn-lv-refresh": func() {
			l.generation++
			_ = l.single.Refresh()
			_ = l.multi.Refresh()
			l.t.set("lv-event", "refresh r%d", l.generation)
		},
	}
}

// ==========================================================================================
// ---- TreeView: nodes, expand / collapse, multi selection, scrolling -----------------------

// Node indices of the visible tree (added in this order):
//
//	0 src { 1 core { 2 types.h, 3 string.h }, 4 ui { 5 button.c, 6 list.c }, 7 main.c }
//	8 docs { 9 guide.md, 10 api.md }   11 README
const (
	nSrc uint32 = iota
	nCore
	nTypes
	nString
	nUI
	nButton
	nList
	nMain
	nDocs
	nGuide
	nAPI
	nReadme
	nNodes
)

type trees struct {
	t             *tour
	tree, scratch *kui.TreeView
	selection     string // last OnSelectionChanged
	lastSelect    int64
	lastToggle    int64
	lastExpanded  int
	readmeLong    bool
}

func treeSelected(tree *kui.TreeView) string {
	values, _ := tree.Selected()
	return indices(values)
}

func treeIs(tree *kui.TreeView, index uint32) bool {
	selected, _ := tree.IsSelected(index)
	return selected
}

func treeIndex(tree *kui.TreeView) int32 {
	index, _ := tree.SelectedIndex()
	return index
}

func treeRows(tree *kui.TreeView) uint32 {
	rows, _ := tree.VisibleRowCount()
	return rows
}

func treeExpanded(tree *kui.TreeView, index uint32) bool {
	node, err := tree.Node(index)
	return err == nil && node.Expanded
}

func (r *trees) scrollTop() float32 {
	box, err := r.t.doc.GetElementByID("tree-box")
	if err != nil || box == nil {
		return -1
	}
	defer box.Close()
	metrics, _ := box.ScrollMetrics()
	return metrics.Y
}

func (r *trees) show() {
	r.t.set("tree-sel", "%s", treeSelected(r.tree))
	r.t.set("tree-idx", "%d", treeIndex(r.tree))
	nodes, _ := r.tree.NodeCount()
	r.t.set("tree-rows", "%d", treeRows(r.tree))
	r.t.set("tree-nodes", "%d", nodes)
	if node, err := r.tree.Node(nReadme); err == nil {
		r.t.set("tree-readme", "%s", node.Text)
	}
	r.t.set("tree-scroll", "%.0f", r.scrollTop())
}

func (r *trees) create(id string, multi bool) (*kui.TreeView, error) {
	container, err := r.t.doc.GetElementByID(id)
	if err != nil || container == nil {
		return nil, fmt.Errorf("没有 #%s: %v", id, err)
	}
	defer container.Close()
	tree, err := kui.NewTreeView()
	if err != nil {
		return nil, err
	}
	config := kui.TreeViewConfig{
		ItemHeight: 26, IndentWidth: 16, Overscan: 2,
		ItemBg: "#ffffff", ItemHoverBg: "#f1f5f9", ItemSelectedBg: "#16a34a", ItemTextColor: "#0f172a",
		MultiSelect: multi,
	}
	if multi {
		add := func(parent int, text string) {
			if parent < 0 {
				_, _ = tree.AddNode(text, "")
			} else {
				_, _ = tree.AddChild(uint32(parent), text, "")
			}
		}
		add(-1, "src")
		add(int(nSrc), "core")
		add(int(nCore), "types.h")
		add(int(nCore), "string.h")
		add(int(nSrc), "ui")
		add(int(nUI), "button.c")
		add(int(nUI), "list.c")
		add(int(nSrc), "main.c")
		add(-1, "docs")
		add(int(nDocs), "guide.md")
		add(int(nDocs), "api.md")
		add(-1, "README")
		config.OnSelect = func(index uint32) {
			r.lastSelect = int64(index)
			r.t.set("tree-event", "on_select(%d)", index)
			r.show()
		}
		config.OnToggle = func(index uint32, expanded bool) {
			r.lastToggle = int64(index)
			r.lastExpanded = 0
			state := "collapsed"
			if expanded {
				r.lastExpanded = 1
				state = "expanded"
			}
			r.t.set("tree-event", "on_toggle(%d, %s)", index, state)
			r.show()
		}
		config.OnSelectionChanged = func(selected []uint32) { r.selection = indices(selected) }
	} else {
		// Scratch tree: 0 A { 2 C }, 1 B.
		_, _ = tree.AddNode("A", "*")
		_, _ = tree.AddNode("B", "*")
		_, _ = tree.AddChild(0, "C", "*")
	}
	if err := tree.Setup(container, config); err != nil {
		tree.Close()
		return nil, err
	}
	return tree, nil
}

func (r *trees) checkNodes() {
	t, tree := r.t, r.tree
	count, _ := tree.NodeCount()
	t.check("tree.get_node_count", count == nNodes, fmt.Sprintf("count %d", count))

	root, _ := tree.Node(nSrc)
	core, _ := tree.Node(nCore)
	t.check("tree.get_node", root.Parent == -1 && len(root.Children) == 3 && core.Parent == int32(nSrc) &&
		!core.Expanded && !core.Disabled && core.Text == "core",
		fmt.Sprintf("root %+v core %+v", root, core))

	// Node reads the children through get_node_child.
	t.check("tree.get_node_child", len(core.Children) == 2 && core.Children[0] == nTypes &&
		core.Children[1] == nString && len(root.Children) == 3 && root.Children[2] == nMain,
		fmt.Sprintf("core %v src %v", core.Children, root.Children))

	_, err := tree.Node(nNodes)
	t.check("tree.get_node-missing", isNotFound(err), fmt.Sprintf("err %v", err))
}

func (r *trees) checkExpand() {
	t, tree := r.t, r.tree
	rows := treeRows(tree)
	t.check("tree.visible-collapsed", rows == 3, fmt.Sprintf("rows %d", rows))

	_ = tree.Expand(nSrc)
	rows = treeRows(tree)
	t.check("tree.expand", rows == 6 && treeExpanded(tree, nSrc) && r.lastToggle == int64(nSrc) &&
		r.lastExpanded == 1, fmt.Sprintf("rows %d on_toggle(%d,%d)", rows, r.lastToggle, r.lastExpanded))

	_ = tree.ExpandAll()
	rows = treeRows(tree)
	t.check("tree.expand_all", rows == nNodes && treeExpanded(tree, nUI), fmt.Sprintf("rows %d", rows))

	_ = tree.Collapse(nCore)
	rows = treeRows(tree)
	t.check("tree.collapse", rows == 10 && !treeExpanded(tree, nCore) && r.lastToggle == int64(nCore) &&
		r.lastExpanded == 0, fmt.Sprintf("rows %d on_toggle(%d,%d)", rows, r.lastToggle, r.lastExpanded))

	_ = tree.Toggle(nCore)
	rows = treeRows(tree)
	_ = tree.Toggle(nCore)
	t.check("tree.toggle", rows == nNodes && treeRows(tree) == 10 && !treeExpanded(tree, nCore),
		fmt.Sprintf("rows after toggles %d, %d", rows, treeRows(tree)))

	_ = tree.CollapseAll()
	rows = treeRows(tree)
	t.check("tree.collapse_all", rows == 3 && !treeExpanded(tree, nSrc) && !treeExpanded(tree, nUI) &&
		!treeExpanded(tree, nDocs), fmt.Sprintf("rows %d", rows))
}

func (r *trees) checkSelection() {
	t, tree := r.t, r.tree
	_ = tree.ExpandAll()
	_ = tree.Select(nDocs)
	got := treeSelected(tree)
	t.check("tree.select", got == "[8]" && treeIndex(tree) == int32(nDocs) && treeIs(tree, nDocs) &&
		r.lastSelect == int64(nDocs) && r.selection == "[8]",
		fmt.Sprintf("selected %s index %d on_select %d on_selection_changed %s", got, treeIndex(tree),
			r.lastSelect, r.selection))

	_ = tree.ToggleSelect(nReadme)
	got = treeSelected(tree)
	t.check("tree.toggle_select-add", got == "[8,11]" && r.selection == "[8,11]", "selected "+got)
	_ = tree.ToggleSelect(nDocs)
	got = treeSelected(tree)
	t.check("tree.toggle_select-remove", got == "[11]" && !treeIs(tree, nDocs), "selected "+got)

	// select sets the anchor; select_range selects the visible rows from it to the target.
	_ = tree.Select(nTypes)
	_ = tree.SelectRange(nButton)
	got = treeSelected(tree)
	t.check("tree.select_range", got == "[2,3,4,5]" && treeIndex(tree) == int32(nButton),
		fmt.Sprintf("selected %s index %d", got, treeIndex(tree)))

	_ = tree.SelectAll()
	values, _ := tree.Selected()
	t.check("tree.select_all", len(values) == int(nNodes) && treeIs(tree, nAPI), fmt.Sprintf("count %d", len(values)))

	_ = tree.ClearSelection()
	values, _ = tree.Selected()
	t.check("tree.clear_selection", len(values) == 0 && treeIndex(tree) == -1 && !treeIs(tree, nTypes) &&
		r.selection == "[]", fmt.Sprintf("count %d index %d", len(values), treeIndex(tree)))
}

// The scratch tree (single selection): replacing selection, SetNode, nodes added after setup.
func (r *trees) checkScratch() {
	t, tree := r.t, r.scratch
	_ = tree.Select(0)
	_ = tree.Select(1)
	got := treeSelected(tree)
	t.check("tree.select-single", got == "[1]" && !treeIs(tree, 0) && treeIndex(tree) == 1,
		fmt.Sprintf("selected %s index %d", got, treeIndex(tree)))

	// The Go wrapper always passes text and icon (an empty string is a value, not "keep").
	_ = tree.SetNode(2, "C2", "*", true)
	_ = tree.SetNode(1, "B", "#", false)
	c, _ := tree.Node(2)
	b, _ := tree.Node(1)
	t.check("tree.set_node", c.Text == "C2" && c.Icon == "*" && c.Disabled && b.Text == "B" &&
		b.Icon == "#" && !b.Disabled, fmt.Sprintf("node 2 %+v node 1 %+v", c, b))

	// kui_c.h: nodes added after setup appear after refresh.
	_ = tree.Expand(0)
	before := treeRows(tree)
	_, _ = tree.AddChild(0, "D", "")
	count, _ := tree.NodeCount()
	added := treeRows(tree)
	_ = tree.Refresh()
	t.check("tree.refresh", before == 3 && added == 3 && count == 4 && treeRows(tree) == 4 &&
		t.has("scratch-tree", "C2") && t.has("scratch-tree", "D"),
		fmt.Sprintf("rows %d -> %d -> %d, nodes %d", before, added, treeRows(tree), count))
}

func (r *trees) buttons() map[string]func() {
	tree := r.tree
	do := func(fn func()) func() { return func() { fn(); r.show() } }
	return map[string]func(){
		"btn-tree-expand-all":   do(func() { _ = tree.ExpandAll() }),
		"btn-tree-collapse-all": do(func() { _ = tree.CollapseAll() }),
		"btn-tree-toggle":       do(func() { _ = tree.Toggle(nSrc) }),
		"btn-tree-collapse":     do(func() { _ = tree.Collapse(nCore) }),
		"btn-tree-select":       do(func() { _ = tree.Select(nCore) }),
		"btn-tree-toggle-sel":   do(func() { _ = tree.ToggleSelect(nDocs) }),
		"btn-tree-range":        do(func() { _ = tree.SelectRange(nMain) }),
		"btn-tree-all":          do(func() { _ = tree.SelectAll() }),
		"btn-tree-clear":        do(func() { _ = tree.ClearSelection() }),
		"btn-tree-scroll":       do(func() { _ = tree.ScrollToNode(nReadme) }),
		"btn-tree-rename": do(func() {
			r.readmeLong = !r.readmeLong
			name := "README"
			if r.readmeLong {
				name = "README.md"
			}
			_ = tree.SetNode(nReadme, name, "", false)
			_ = tree.Refresh()
		}),
	}
}

// ==========================================================================================
// ---- VirtualList: focus, insert / remove, total count, scrolling --------------------------

const (
	vlCount    = 1000
	vlRow      = 24
	vlOverscan = 3
	vlNewBase  = 1000000 // ids at or above are inserted rows ("New n")
)

type vlist struct {
	t          *tour
	list       *kui.VirtualList
	items      []uint32 // the data model: one id per row
	nextNew    uint32
	generation int
	rendered   int   // RenderItem calls
	focus      int32 // last OnFocusChange
	focusOld   int32
	total      uint32 // set_total_items button
}

func (v *vlist) reset() {
	v.items = make([]uint32, vlCount)
	for i := range v.items {
		v.items[i] = uint32(i)
	}
	v.nextNew = 1
}

func (v *vlist) label(index uint32) string {
	id := index
	if int(index) < len(v.items) {
		id = v.items[index]
	}
	text := fmt.Sprintf("Item %d", id)
	if id >= vlNewBase {
		text = fmt.Sprintf("New %d", id-vlNewBase)
	}
	if v.generation > 0 {
		text += fmt.Sprintf(" r%d", v.generation)
	}
	return text
}

func (v *vlist) modelInsert(index, count int) {
	added := make([]uint32, count)
	for i := range added {
		added[i] = vlNewBase + v.nextNew
		v.nextNew++
	}
	v.items = append(v.items[:index], append(added, v.items[index:]...)...)
}

func (v *vlist) modelRemove(index, count int) {
	v.items = append(v.items[:index], v.items[index+count:]...)
}

func (v *vlist) state() kui.VirtualListState {
	state, _ := v.list.State()
	return state
}

func (v *vlist) show() {
	state := v.state()
	v.t.set("vl-total", "%d", state.TotalItems)
	v.t.set("vl-focus", "%d", state.FocusedIndex)
	v.t.set("vl-start", "%d", state.VisibleStart)
	v.t.set("vl-end", "%d", state.VisibleEnd)
	v.t.set("vl-first", "%s", v.label(0))
}

func (v *vlist) create() error {
	container, err := v.t.doc.GetElementByID("vl-box")
	if err != nil || container == nil {
		return fmt.Errorf("没有 #vl-box: %v", err)
	}
	defer container.Close()
	v.reset()
	v.focus, v.focusOld, v.total = -1, -2, vlCount
	v.list, err = kui.NewVirtualList(container, kui.VirtualListConfig{
		TotalItems: vlCount, ItemHeight: vlRow, Overscan: vlOverscan, EnableKeyboard: true,
		RenderItem: func(document *kui.Document, index uint32) *kui.Element {
			v.rendered++
			row, err := document.CreateElement("div")
			if err != nil || row == nil {
				return nil
			}
			background := "#ffffff"
			if int32(index) == v.focus {
				background = "#f59e0b"
			}
			for _, pair := range [][2]string{{"display", "flex"}, {"align-items", "center"},
				{"padding", "0 10px"}, {"font-size", "12px"}, {"color", "#0f172a"},
				{"background", background}} {
				_ = row.SetStyleProperty(pair[0], pair[1], false)
			}
			_ = row.SetTextContent(v.label(index))
			return row // KUI takes it
		},
		OnFocusChange: func(oldIndex, newIndex int32) {
			v.focusOld, v.focus = oldIndex, newIndex
			v.t.set("vl-event", "focus %d -> %d", oldIndex, newIndex)
		},
		OnActivate: func(index uint32) { v.t.set("vl-event", "on_activate(%d)", index) },
	})
	return err
}

func (v *vlist) checks() {
	t, list := v.t, v.list
	state := v.state()
	t.check("vlist.get_state", state.TotalItems == vlCount && state.ItemHeight == vlRow &&
		state.TotalHeight == 24000 && state.FocusedIndex == -1 && state.VisibleStart == 0 &&
		state.VisibleEnd > 0 && state.VisibleEnd <= vlCount && v.rendered > 0 && t.has("vl-box", "Item 0"),
		fmt.Sprintf("state %+v", state))

	_ = list.ScrollToIndex(500)
	state = v.state()
	t.check("vlist.scroll_to_index", state.VisibleStart <= 500 && state.VisibleEnd > 500 &&
		t.has("vl-box", "Item 500"), fmt.Sprintf("rendered %d-%d", state.VisibleStart, state.VisibleEnd))

	_ = list.ScrollToBottom()
	state = v.state()
	t.check("vlist.scroll_to_bottom", state.VisibleEnd == vlCount && t.has("vl-box", "Item 999"),
		fmt.Sprintf("rendered %d-%d", state.VisibleStart, state.VisibleEnd))

	_ = list.ScrollToTop()
	state = v.state()
	t.check("vlist.scroll_to_top", state.VisibleStart == 0 && t.has("vl-box", "Item 0"),
		fmt.Sprintf("rendered %d-%d", state.VisibleStart, state.VisibleEnd))

	// kui_c.h: set_focused_index calls on_focus_change when it changes.
	_ = list.SetFocusedIndex(10)
	state = v.state()
	t.check("vlist.set_focused_index", state.FocusedIndex == 10 && v.focusOld == -1 && v.focus == 10 &&
		state.VisibleStart <= 10 && state.VisibleEnd > 10,
		fmt.Sprintf("focus %d event %d->%d rendered %d-%d", state.FocusedIndex, v.focusOld, v.focus,
			state.VisibleStart, state.VisibleEnd))

	// kui_c.h: clamped to the rows; -1 clears.
	_ = list.SetFocusedIndex(5000)
	clamped, eventOld := v.state().FocusedIndex, v.focusOld
	_ = list.SetFocusedIndex(-1)
	state = v.state()
	t.check("vlist.set_focused_index-clamp", clamped == 999 && eventOld == 10 && state.FocusedIndex == -1 &&
		v.focusOld == 999 && v.focus == -1,
		fmt.Sprintf("5000 -> %d, -1 -> %d (event %d->%d)", clamped, state.FocusedIndex, v.focusOld, v.focus))

	_ = list.SetFocusedIndex(10)
	_ = list.MoveFocus(5)
	moved := v.state().FocusedIndex
	_ = list.MoveFocus(-100)
	state = v.state()
	t.check("vlist.move_focus", moved == 15 && state.FocusedIndex == 0 && v.focus == 0,
		fmt.Sprintf("10+5 -> %d, -100 -> %d", moved, state.FocusedIndex))
	_ = list.SetFocusedIndex(-1)

	// Insert three rows at the top, then remove them again.
	_ = list.ScrollToTop()
	v.modelInsert(0, 3)
	_ = list.InsertItems(0, 3)
	_ = list.ScrollToTop()
	state = v.state()
	t.check("vlist.insert_items", state.TotalItems == vlCount+3 && state.TotalHeight == 24072 &&
		t.has("vl-box", "New 1"), fmt.Sprintf("total %d height %.0f", state.TotalItems, state.TotalHeight))

	v.modelRemove(0, 3)
	_ = list.RemoveItems(0, 3)
	state = v.state()
	t.check("vlist.remove_items", state.TotalItems == vlCount && !t.has("vl-box", "New 1") &&
		t.has("vl-box", "Item 0"), fmt.Sprintf("total %d", state.TotalItems))

	_ = list.SetTotalItems(50)
	state = v.state()
	t.check("vlist.set_total_items", state.TotalItems == 50 && state.TotalHeight == 1200 && state.VisibleEnd <= 50,
		fmt.Sprintf("total %d height %.0f rendered %d-%d", state.TotalItems, state.TotalHeight,
			state.VisibleStart, state.VisibleEnd))
	_ = list.SetTotalItems(vlCount)

	before := v.rendered
	v.generation = 1
	_ = list.Refresh()
	t.check("vlist.refresh", v.rendered > before && t.has("vl-box", "Item 0 r1"),
		fmt.Sprintf("render calls %d -> %d", before, v.rendered))

	// Removing the rows around the focused row: the focus must stay within the rows.
	_ = list.SetFocusedIndex(999)
	v.modelRemove(990, 10)
	_ = list.RemoveItems(990, 10)
	state = v.state()
	t.knownBug("vlist.remove_items-focus", state.TotalItems == 990 && state.FocusedIndex < 990,
		"remove_items leaves focused_index past the last row",
		fmt.Sprintf("total %d focus %d", state.TotalItems, state.FocusedIndex))

	// Back to the starting state.
	v.reset()
	_ = list.SetTotalItems(vlCount)
	_ = list.SetFocusedIndex(-1)
	_ = list.ScrollToTop()
	_ = list.Refresh()
}

func (v *vlist) buttons() map[string]func() {
	list := v.list
	afterFocus := func(fn func()) func() {
		return func() {
			fn()
			_ = list.Refresh() // repaint the amber focus row
			v.show()
		}
	}
	do := func(fn func()) func() { return func() { fn(); v.show() } }
	return map[string]func(){
		"btn-vl-focus": afterFocus(func() { _ = list.SetFocusedIndex(10) }),
		"btn-vl-prev":  afterFocus(func() { _ = list.MoveFocus(-1) }),
		"btn-vl-next":  afterFocus(func() { _ = list.MoveFocus(1) }),
		"btn-vl-insert": do(func() {
			v.modelInsert(0, 3)
			_ = list.InsertItems(0, 3)
		}),
		"btn-vl-remove": do(func() {
			if len(v.items) < 3 {
				return
			}
			v.modelRemove(0, 3)
			_ = list.RemoveItems(0, 3)
		}),
		"btn-vl-goto":   do(func() { _ = list.ScrollToIndex(500) }),
		"btn-vl-top":    do(func() { _ = list.ScrollToTop() }),
		"btn-vl-bottom": do(func() { _ = list.ScrollToBottom() }),
		"btn-vl-total": do(func() {
			if v.total == 50 {
				v.total = uint32(len(v.items))
			} else {
				v.total = 50
			}
			_ = list.SetTotalItems(v.total)
		}),
		"btn-vl-refresh": do(func() {
			v.generation++
			_ = list.Refresh()
			v.t.set("vl-event", "refresh r%d", v.generation)
		}),
	}
}

// ==========================================================================================

func ready(ctx *runner.Context) error {
	t := &tour{doc: ctx.Document}
	l := &lists{t: t, count: lvCount}
	r := &trees{t: t, lastSelect: -1, lastToggle: -1, lastExpanded: -1}
	v := &vlist{t: t}
	var err error
	if l.single, err = l.create("lv-single", kui.SelectionSingle, "single"); err != nil {
		return err
	}
	ctx.OnClose(func() { l.single.Close() })
	if l.multi, err = l.create("lv-multi", kui.SelectionMulti, "multi"); err != nil {
		return err
	}
	ctx.OnClose(func() { l.multi.Close() })
	if r.tree, err = r.create("tree-box", true); err != nil {
		return err
	}
	ctx.OnClose(func() { r.tree.Close() })
	if r.scratch, err = r.create("scratch-tree", false); err != nil {
		return err
	}
	ctx.OnClose(func() { r.scratch.Close() })
	if err = v.create(); err != nil {
		return err
	}
	ctx.OnClose(func() { v.list.Close() })

	l.checks()
	r.checkNodes()
	r.checkExpand()
	r.checkSelection()
	r.checkScratch()
	v.checks()

	// Deferred: scrolling the tree needs a laid-out frame, then the verdict.
	_, err = kui.SetTimeout(func() {
		_ = r.tree.ScrollToNode(nReadme) // the tree is fully expanded here
		_, _ = kui.SetTimeout(func() {
			scrolled := r.scrollTop()
			_ = r.tree.ScrollToNode(nSrc)
			back := r.scrollTop()
			t.check("tree.scroll_to_node", scrolled > 0 && back == 0,
				fmt.Sprintf("README -> %.0f, src -> %.0f", scrolled, back))
			// Back to the starting state.
			_ = r.tree.CollapseAll()
			_ = r.tree.ClearSelection()
			r.show()
			l.show()
			v.show()
			t.set("lv-event", "-")
			t.set("tree-event", "-")
			t.set("vl-event", "-")
			t.finish()
		}, 150)
	}, 150)
	if err != nil {
		return err
	}

	for _, group := range []map[string]func(){l.buttons(), r.buttons(), v.buttons()} {
		for id, fn := range group {
			element, err := ctx.Document.GetElementByID(id)
			if err != nil || element == nil {
				return fmt.Errorf("没有 #%s: %v", id, err)
			}
			ctx.Keep(element)
			handler := fn
			if _, err := element.AddEventListener("click", func(*kui.Event) { handler() }); err != nil {
				return err
			}
		}
	}
	return nil
}
