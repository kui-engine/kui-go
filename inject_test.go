//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// 结构化注入、盒模型、DevToolsCall、定时器与 PDF（T-DEBT-SDK-GO-LOWLEVEL-SIGNATURES）。
func TestStructuredInjectionBoxModelTimersPDF(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-lowlevel", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		if err := document.LoadHTML(`<html><body style="margin:0"><div id="box" tabindex="0" style="position:absolute;left:20px;top:30px;width:100px;height:50px;padding:5px;border:2px solid red;margin:3px"></div></body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.01)
		}

		// 盒模型：四层盒的左上角依次向外扩出 margin / border / padding。
		box, err := document.ElementBoxModel("#box")
		if err != nil {
			t.Fatal(err)
		}
		if box.Width != 100 || box.Height != 50 {
			t.Fatalf("content size: %+v", box)
		}
		if box.Margin[0] != (Point{X: 20, Y: 30}) || box.Border[0] != (Point{X: 23, Y: 33}) ||
			box.Padding[0] != (Point{X: 25, Y: 35}) || box.Content[0] != (Point{X: 30, Y: 40}) {
			t.Fatalf("box corners: %+v", box)
		}
		if box.Content[2] != (Point{X: 130, Y: 90}) {
			t.Fatalf("content bottom-right: %+v", box.Content[2])
		}
		var kuiErr *Error
		if _, err := document.ElementBoxModel("#missing"); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotFound {
			t.Fatalf("missing selector: %v", err)
		}

		// 结构化的鼠标 / 键盘注入，事件标志与修饰键。
		element, err := document.GetElementByID("box")
		if err != nil || element == nil {
			t.Fatalf("box: %v", err)
		}
		defer element.Close()
		var clickFlags EventFlags
		clicks := 0
		if _, err := element.AddEventListener("click", func(e *Event) {
			clicks++
			clickFlags, _ = e.FlagValues()
		}); err != nil {
			t.Fatal(err)
		}
		var keyModifiers KeyModifiers
		if _, err := element.AddEventListener("keydown", func(e *Event) { keyModifiers, _ = e.ModifierKeys() }); err != nil {
			t.Fatal(err)
		}
		for _, kind := range []MouseEventType{MousePressed, MouseReleased} {
			if err := document.DispatchMouseEvent(MouseInjection{Kind: kind, X: 60, Y: 60, Button: MouseButtonLeft, ClickCount: 1}); err != nil {
				t.Fatal(err)
			}
		}
		if clicks != 1 || !clickFlags.Bubbles || !clickFlags.Cancelable || clickFlags.DefaultPrevented {
			t.Fatalf("click: %d %+v", clicks, clickFlags)
		}
		_ = element.Focus()
		if err := document.DispatchKeyEvent(KeyInjection{Kind: KeyDown, Key: "A", Code: "KeyA", Modifiers: InjectShift | InjectCtrl}); err != nil {
			t.Fatal(err)
		}
		if keyModifiers != (KeyModifiers{Ctrl: true, Shift: true}) {
			t.Fatalf("modifiers: %+v", keyModifiers)
		}

		// DevToolsCall：结果解进结构体；命令错误是 *DevToolsError。
		var nodes []map[string]any // KUI 的 DOM.getDocument 结果是节点数组
		if err := document.DevToolsCall("DOM.getDocument", nil, &nodes); err != nil || len(nodes) == 0 {
			t.Fatalf("DOM.getDocument: %v %v", nodes, err)
		}
		var devErr *DevToolsError
		if err := document.DevToolsCall("No.suchMethod", map[string]int{"x": 1}, nil); !errors.As(err, &devErr) {
			t.Fatalf("unknown method should be a DevToolsError: %v", err)
		}

		// 定时器：一次性的只触发一次，间隔的清除后停止。
		once, repeated := 0, 0
		if _, err := SetTimeout(func() { once++ }, 10); err != nil {
			t.Fatal(err)
		}
		interval, err := SetInterval(func() { repeated++ }, 10)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 100 && (once == 0 || repeated < 3); i++ {
			_ = WaitEvents(0.01)
		}
		if once != 1 || repeated < 3 {
			t.Fatalf("timers: once=%d repeated=%d", once, repeated)
		}
		if err := ClearTimer(interval); err != nil {
			t.Fatal(err)
		}
		stopped := repeated
		for i := 0; i < 10; i++ {
			_ = WaitEvents(0.01)
		}
		if repeated != stopped || once != 1 {
			t.Fatalf("after ClearTimer: once=%d repeated=%d (was %d)", once, repeated, stopped)
		}

		// PDF：写出的文件以 %PDF 开头。
		path := filepath.Join(t.TempDir(), "page.pdf")
		written, err := document.PrintToPDF(path, PDFOptions{Title: "Go test", PageWidthPT: 612, PageHeightPT: 792})
		if err != nil || !written {
			t.Fatalf("PrintToPDF: %v %v", written, err)
		}
		data, err := os.ReadFile(path)
		if err != nil || !bytes.HasPrefix(data, []byte("%PDF")) {
			t.Fatalf("PDF file: %d bytes %v", len(data), err)
		}
	})
}
