//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"errors"
	"fmt"
	"testing"
)

// C ABI 第 18 节（3.61.8）：内置元素、主题、VirtualList、ListView、TreeView。
func TestComponents(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		if err := Init(); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := Shutdown(); err != nil {
				t.Errorf("Shutdown: %v", err)
			}
		}()
		window, err := NewWindow(WindowProps{Title: "kui-go-components", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()

		if profile, _ := document.StyleProfile(); profile != StyleNativeGUI {
			t.Fatalf("default style profile: %v", profile)
		}
		if err := document.SetStyleProfile(StyleBrowserCompat); err != nil {
			t.Fatal(err)
		}
		if profile, _ := document.StyleProfile(); profile != StyleBrowserCompat {
			t.Fatalf("StyleProfile: %v", profile)
		}
		_ = document.SetStyleProfile(StyleNativeGUI)
		if err := document.RegisterBuiltinComponents(); err != nil {
			t.Fatal(err)
		}
		if defined, _ := document.CustomElementDefined("kui-dialog"); !defined {
			t.Fatal("<kui-dialog> should be defined")
		}

		themes, err := NewThemeManager()
		if err != nil {
			t.Fatal(err)
		}
		defer themes.Close()
		if names, _ := themes.ThemeNames(); len(names) != 2 {
			t.Fatalf("built-in themes: %v", names)
		}
		if err := themes.RegisterTheme("brand", map[string]string{"--brand-bg": "#123456"}); err != nil {
			t.Fatal(err)
		}
		if vars, err := themes.Theme("brand"); err != nil || vars["--brand-bg"] != "#123456" {
			t.Fatalf("Theme: %v %v", vars, err)
		}
		var kuiErr *Error
		if _, err := themes.Theme("missing"); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotFound {
			t.Fatalf("unknown theme: %v", err)
		}
		if err := document.LoadHTML(`<html><body style="margin:0"><div id="vl" style="width:200px;height:100px"></div><div style="position:absolute;left:300px;top:0"><div id="lv" style="width:200px;height:120px"></div></div><div style="position:absolute;left:0;top:200px"><div id="tv" style="width:200px;height:120px"></div></div></body></html>`); err != nil {
			t.Fatal(err)
		}
		if err := themes.Apply("brand", document); err != nil {
			t.Fatal(err)
		}
		if active, _ := themes.ActiveTheme(); active != "brand" {
			t.Fatalf("ActiveTheme: %q", active)
		}
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.01)
		}

		byID := func(id string) *Element {
			element, err := document.GetElementByID(id)
			if err != nil || element == nil {
				t.Fatalf("%s: %v", id, err)
			}
			return element
		}

		// VirtualList
		rendered, focusChanges := 0, 0
		vlContainer := byID("vl")
		defer vlContainer.Close()
		vlist, err := NewVirtualList(vlContainer, VirtualListConfig{
			TotalItems: 1000, ItemHeight: 20,
			RenderItem: func(doc *Document, index uint32) *Element {
				row, err := doc.CreateElement("div")
				if err != nil {
					return nil
				}
				_ = row.SetTextContent(fmt.Sprintf("row %d", index))
				rendered++
				return row
			},
			OnFocusChange: func(int32, int32) { focusChanges++ },
		})
		if err != nil {
			t.Fatal(err)
		}
		state, _ := vlist.State()
		if rendered == 0 || state.VisibleEnd-state.VisibleStart >= 50 || state.TotalHeight != 20000 {
			t.Fatalf("virtual list: rendered %d state %+v", rendered, state)
		}
		_ = vlist.ScrollToIndex(500)
		if state, _ = vlist.State(); state.VisibleStart > 500 || 500 >= state.VisibleEnd {
			t.Fatalf("ScrollToIndex: %+v", state)
		}
		_ = vlist.SetFocusedIndex(3)
		_ = vlist.MoveFocus(2)
		if state, _ = vlist.State(); state.FocusedIndex != 5 || focusChanges != 2 {
			t.Fatalf("focus: %+v %d", state, focusChanges)
		}
		if err := vlist.RemoveItems(990, 20); err == nil {
			t.Fatal("removing past the end should fail")
		}
		_ = vlist.Close()

		// ListView
		selected := -1
		lvContainer := byID("lv")
		defer lvContainer.Close()
		list, err := NewListView(lvContainer, ListViewConfig{
			ItemCount: 100, ItemHeight: 30, SelectionMode: SelectionMulti, ItemSelectedBg: "#ff0000",
			ItemProvider: func(index uint32) ListViewItem {
				return ListViewItem{Text: fmt.Sprintf("Item %d", index), Subtitle: "sub"}
			},
			OnSelect: func(index uint32) { selected = int(index) },
		})
		if err != nil {
			t.Fatal(err)
		}
		defer list.Close()
		_ = list.Select(5)
		_ = list.SelectRange(10, 13)
		if values, _ := list.Selected(); len(values) != 4 || values[0] != 5 || values[3] != 12 {
			t.Fatalf("ListView selection: %v", values)
		}
		_ = list.ClearSelection()
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.01)
		}
		_ = document.DispatchMouse(MousePressed, 350, 70, MouseButtonLeft, 0, 1, 0, 0)
		_ = document.DispatchMouse(MouseReleased, 350, 70, MouseButtonLeft, 0, 1, 0, 0)
		if is, _ := list.IsSelected(2); selected != 2 || !is {
			t.Fatalf("click should select row 2: %d %v", selected, is)
		}

		// TreeView
		toggles, selectionSize := 0, 0
		tree, err := NewTreeView()
		if err != nil {
			t.Fatal(err)
		}
		defer tree.Close()
		a, _ := tree.AddNode("A", "+")
		a1, _ := tree.AddChild(a, "A1", "")
		_, _ = tree.AddChild(a, "A2", "")
		b, _ := tree.AddNode("B", "")
		if node, _ := tree.Node(a); len(node.Children) != 2 || node.Icon != "+" || node.Parent != -1 {
			t.Fatalf("Node: %+v", node)
		}
		_ = tree.SetNode(b, "B2", "", true)
		if node, _ := tree.Node(b); node.Text != "B2" || !node.Disabled {
			t.Fatalf("SetNode: %+v", node)
		}
		// UpdateNode：没设的字段保持原值。
		newText := "B3"
		if err := tree.UpdateNode(b, TreeNodeUpdate{Text: &newText}); err != nil {
			t.Fatal(err)
		}
		if node, _ := tree.Node(b); node.Text != "B3" || !node.Disabled {
			t.Fatalf("UpdateNode 只改文字，禁用状态应保持: %+v", node)
		}
		enabled := false
		if err := tree.UpdateNode(a, TreeNodeUpdate{Disabled: &enabled}); err != nil {
			t.Fatal(err)
		}
		if node, _ := tree.Node(a); node.Text != "A" || node.Icon != "+" || node.Disabled {
			t.Fatalf("UpdateNode 只改禁用，文字与图标应保持: %+v", node)
		}
		_ = tree.SetNode(b, "B2", "", true)
		tvContainer := byID("tv")
		defer tvContainer.Close()
		if err := tree.Setup(tvContainer, TreeViewConfig{
			MultiSelect:        true,
			OnToggle:           func(uint32, bool) { toggles++ },
			OnSelectionChanged: func(values []uint32) { selectionSize = len(values) },
		}); err != nil {
			t.Fatal(err)
		}
		if rows, _ := tree.VisibleRowCount(); rows != 2 {
			t.Fatalf("collapsed rows: %d", rows)
		}
		_ = tree.Expand(a)
		if rows, _ := tree.VisibleRowCount(); rows != 4 || toggles != 1 {
			t.Fatalf("Expand: %d %d", rows, toggles)
		}
		_ = tree.Select(a1)
		_ = tree.ToggleSelect(b)
		if values, _ := tree.Selected(); len(values) != 2 || values[0] != a1 || values[1] != b || selectionSize != 2 {
			t.Fatalf("multi-select: %v %d", values, selectionSize)
		}
		if err := tree.Expand(42); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotFound {
			t.Fatalf("unknown node: %v", err)
		}
	})
}
