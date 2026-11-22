//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// C ABI 第 9 节（3.61.8）：DOM 编辑与选择器、指针捕获、时钟与显示器查找。
func TestCommonDOMAndPlatform(t *testing.T) {
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
		document, err := NewDocument()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		if err := document.LoadHTML(`<html><body><div id="list" class="box"><p id="a">A</p><p id="b">B</p></div><span id="free">f</span></body></html>`); err != nil {
			t.Fatal(err)
		}
		get := func(id string) *Element {
			element, err := document.GetElementByID(id)
			if err != nil || element == nil {
				t.Fatalf("#%s: %v", id, err)
			}
			return element
		}
		list, a, b, free := get("list"), get("a"), get("b"), get("free")
		ids := func() string {
			var parts []string
			for child, _ := list.FirstChild(); child != nil; child, _ = child.NextSibling() {
				id, _ := child.ID()
				parts = append(parts, id)
			}
			return strings.Join(parts, ",")
		}

		if err := list.InsertBefore(free, b); err != nil || ids() != "a,free,b" {
			t.Fatalf("InsertBefore: %v %q", err, ids())
		}
		if err := list.InsertBefore(a, nil); err != nil || ids() != "free,b,a" {
			t.Fatalf("InsertBefore(nil): %v %q", err, ids())
		}
		if err := a.InsertBefore(list, nil); err == nil {
			t.Fatal("插入祖先应被拒绝")
		}
		replacement, err := document.CreateElement("em")
		if err != nil {
			t.Fatal(err)
		}
		if err := list.ReplaceChild(replacement, b); err != nil {
			t.Fatal(err)
		}
		if copy, err := list.CloneNode(true); err != nil || copy == nil {
			t.Fatalf("CloneNode: %v", err)
		} else if count, _ := copy.ChildCount(); count != 3 {
			t.Fatalf("深复制的子元素数 %d", count)
		}
		if ok, err := list.Matches("div.box"); err != nil || !ok {
			t.Fatalf("Matches: %v %v", ok, err)
		}
		if up, err := a.Closest(".box"); err != nil || up == nil {
			t.Fatalf("Closest: %v", err)
		} else if id, _ := up.ID(); id != "list" {
			t.Fatalf("Closest 找到 %q", id)
		}
		if none, err := a.Closest(".missing"); err != nil || none != nil {
			t.Fatalf("Closest 没有匹配时应返回 nil：%v %v", none, err)
		}
		if html, err := a.OuterHTML(); err != nil || !strings.Contains(html, `id="a"`) {
			t.Fatalf("OuterHTML: %q %v", html, err)
		}
		if value, err := list.ComputedStyleValue("display"); err != nil || value != "block" {
			t.Fatalf("ComputedStyleValue(display): %q %v", value, err)
		}
		if value, err := list.ComputedStyleValue("no-such-property"); err != nil || value != "" {
			t.Fatalf("ComputedStyleValue(未知属性): %q %v", value, err)
		}
		if err := list.InsertAdjacentHTML("beforeend", `<i id="tail"></i>`); err != nil {
			t.Fatal(err)
		}
		if last, _ := list.LastChild(); last == nil {
			t.Fatal("InsertAdjacentHTML 没有插入")
		} else if id, _ := last.ID(); id != "tail" {
			t.Fatalf("InsertAdjacentHTML 插入 %q", id)
		}
		if err := a.AppendText(" more"); err != nil {
			t.Fatal(err)
		}
		if text, _ := a.TextContent(); text != "A more" {
			t.Fatalf("AppendText: %q", text)
		}
		if err := a.SetPointerCapture(1); err != nil {
			t.Fatal(err)
		}
		if captured, _ := a.HasPointerCapture(1); !captured {
			t.Fatal("SetPointerCapture 后应处于捕获状态")
		}
		if err := a.ReleasePointerCapture(1); err != nil {
			t.Fatal(err)
		}
		if focused, err := a.HasFocus(); err != nil || focused {
			t.Fatalf("HasFocus: %v %v", focused, err)
		}

		if items, err := document.GetElementsByTagName("p"); err != nil || len(items) != 1 {
			t.Fatalf("GetElementsByTagName: %d %v", len(items), err)
		}
		if items, err := document.GetElementsByClassName("box"); err != nil || len(items) != 1 {
			t.Fatalf("GetElementsByClassName: %d %v", len(items), err)
		}

		first, err := Now()
		if err != nil || first <= 0 {
			t.Fatalf("Now: %v %v", first, err)
		}
		primary, err := PrimaryMonitor()
		if err != nil {
			t.Fatal(err)
		}
		hit, err := MonitorFromPoint(primary.Bounds.X+1, primary.Bounds.Y+1)
		if err != nil || !hit.Primary {
			t.Fatalf("MonitorFromPoint: %+v %v", hit, err)
		}
		if _, err := MessageBox("t", "m", MessageBoxType(9), MessageBoxOK); err == nil {
			t.Fatal("非法的消息框类型应被拒绝")
		}
	})
}

// C ABI 第 9 节的窗口部分：会弹出一个 320×240 的真实窗口。
func TestCommonWindowOperations(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-common", Width: 320, Height: 240})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		if err := window.Focus(); err != nil {
			t.Fatal(err)
		}
		if err := window.SetCursor(CursorPointer); err != nil {
			t.Fatal(err)
		}
		if err := window.SetFullscreen(true); err != nil {
			t.Fatal(err)
		}
		if on, _ := window.IsFullscreen(); !on {
			t.Fatal("SetFullscreen(true) 之后应处于全屏")
		}
		if err := window.SetFullscreen(false); err != nil {
			t.Fatal(err)
		}
		if err := window.SetFrameless(true); err != nil {
			t.Fatal(err)
		}
		// Linux / macOS 的内核还没有无边框窗口（setter 是空函数，T-DEBT-POSIX-FRAMELESS-RUNTIME-NOOP）：
		// 在那里断言「仍为 false」当绊线，内核实现后这里会失败，届时同一提交改回与 Windows 相同的断言。
		if state, err := window.FrameState(); err != nil || state.Frameless != (runtime.GOOS == "windows") {
			t.Fatalf("FrameState: %+v %v", state, err)
		}
		want := FrameRadii{TopLeft: 4, TopRight: 8, BottomRight: 12, BottomLeft: 16}
		if err := window.SetFrameRadii(want); err != nil {
			t.Fatal(err)
		}
		if got, ok, err := window.FrameRadii(); err != nil || !ok || got != want {
			t.Fatalf("FrameRadii: %+v %v %v", got, ok, err)
		}
		if err := window.ClearFrameRadius(); err != nil {
			t.Fatal(err)
		}
		if _, ok, err := window.FrameRadii(); err != nil || ok {
			t.Fatalf("清除后不应再有圆角：%v %v", ok, err)
		}
		if err := window.SetIMECandidatePos(10, 20); err != nil {
			t.Fatal(err)
		}
		if monitor, err := MonitorFromWindow(window); err != nil || monitor.Bounds.Width <= 0 {
			t.Fatalf("MonitorFromWindow: %+v %v", monitor, err)
		}
	})
}

// C ABI 第 10 节：菜单栏、托盘、快捷键、Toast、截图（会弹出一个真实窗口）。
func TestDesktopIntegration(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-desktop", Width: 320, Height: 240})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()

		menus := []Menu{{Text: "&File", Items: []MenuItem{
			{Text: "&Open", Shortcut: "Ctrl+O", ID: 1},
			{Separator: true},
			{Text: "Recent", Children: []MenuItem{{Text: "a.txt", ID: 31}}},
			{Text: "Wrap", ID: 2, Checkable: true},
		}}}
		if err := window.SetMenuBar(menus, func(uint32) {}); err != nil {
			t.Fatal(err)
		}
		if err := window.SetMenuItemChecked(2, true); err != nil {
			t.Fatal(err)
		}
		if err := window.SetMenuItemEnabled(31, false); err != nil {
			t.Fatal(err)
		}
		if err := window.SetMenuItemChecked(999, true); err == nil {
			t.Fatal("不存在的菜单项 ID 应报错")
		}
		if err := window.ClearMenuBar(); err != nil {
			t.Fatal(err)
		}

		if err := window.CreateTray("", "kui go tray", true); err == nil {
			var got []TrayEvent
			if err := window.SetTrayEventHandler(func(e TrayEvent) { got = append(got, e) }); err != nil {
				t.Fatal(err)
			}
			_ = window.SimulateTrayEvent(TrayClick)
			_ = window.SimulateTrayEvent(TrayBalloonClick)
			if len(got) != 2 || got[0] != TrayClick || got[1] != TrayBalloonClick {
				t.Fatalf("托盘事件：%v", got)
			}
			if err := window.SetTrayContextMenu([]MenuItem{{Text: "Quit", ID: 9}}, func(uint32) {}); err != nil {
				t.Fatal(err)
			}
			if err := window.DestroyTray(); err != nil {
				t.Fatal(err)
			}
		}

		first, err := window.BindShortcut("Ctrl+S", func() {})
		if err != nil || first == 0 {
			t.Fatalf("BindShortcut: %d %v", first, err)
		}
		if again, _ := window.BindShortcut("Ctrl+S", func() {}); again != first {
			t.Fatalf("同一组合再绑应返回同一 ID：%d %d", again, first)
		}
		if _, err := window.BindShortcut("", func() {}); err == nil {
			t.Fatal("空快捷键应被拒绝")
		}
		if err := window.UnbindAllShortcuts(); err != nil {
			t.Fatal(err)
		}

		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		if err := document.LoadHTML(`<html><body><p>desktop</p></body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			_ = PollEvents()
		}
		id, err := document.ShowToast(ToastOptions{Message: "Saved", Level: ToastSuccess, Closable: true})
		if err != nil || id == 0 {
			t.Fatalf("ShowToast: %d %v", id, err)
		}
		if count, _ := document.ToastCount(); count != 1 {
			t.Fatalf("ToastCount: %d", count)
		}
		_ = document.DismissToast(id)
		shot := filepath.Join(t.TempDir(), "shot.png")
		if saved, err := window.SaveScreenshot(shot); err != nil || !saved {
			t.Fatalf("SaveScreenshot: %v %v", saved, err)
		}
	})
}
