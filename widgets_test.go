//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// C ABI 第 12 节（3.61.8）：弹出层、<input> 选择器、游戏音频。
func TestWidgetsPopupsPickerAudio(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-widgets", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		if err := document.LoadHTML(`<html><body><input id="d" type="date" value="2026-11-16"><input id="t" type="text"><audio id="a"></audio><lottie-player id="l"></lottie-player></body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.01)
		}

		popups := document.Popups()
		if id, err := popups.ShowTooltip("hello", Rect{X: 10, Y: 10, Width: 40, Height: 20}, PopupBottom); err != nil || id == 0 {
			t.Fatalf("ShowTooltip: %d %v", id, err)
		}
		if id, err := popups.ShowContextMenu(20, 30, []MenuItem{{Text: "Copy", ID: 1}, {Separator: true}}, func(uint32) {}); err != nil || id == 0 {
			t.Fatalf("ShowContextMenu: %d %v", id, err)
		}
		if has, _ := popups.Has(PopupContextMenu); !has {
			t.Fatal("右键菜单应在显示")
		}
		_ = popups.DismissAll()
		if any, _ := popups.Any(); any {
			t.Fatal("DismissAll 之后不应还有弹出层")
		}
		if err := popups.SetTooltipTheme(TooltipTheme{Background: 0x11223344, Foreground: 0xffffffff, Border: 0, Radius: 4}); err != nil {
			t.Fatal(err)
		}
		if theme, _ := popups.TooltipTheme(); theme.Background != 0x11223344 || theme.Radius != 4 {
			t.Fatalf("TooltipTheme: %+v", theme)
		}
		if on, err := popups.DefaultContextMenuEnabled(); err != nil || on {
			t.Fatalf("内置右键菜单应默认关闭: %v %v", on, err)
		}
		if err := popups.SetDefaultContextMenuEnabled(true); err != nil {
			t.Fatal(err)
		}
		if on, _ := popups.DefaultContextMenuEnabled(); !on {
			t.Fatal("开启后 DefaultContextMenuEnabled 应为 true")
		}
		_ = popups.SetDefaultContextMenuEnabled(false)

		date, _ := document.GetElementByID("d")
		defer date.Close()
		text, _ := document.GetElementByID("t")
		defer text.Close()
		if err := date.OpenInputPicker(PickerNone); err != nil {
			t.Fatal(err)
		}
		picker := document.InputPicker()
		state, err := picker.State()
		if err != nil || !state.Open || state.Mode != PickerDate || state.VisibleYear != 2026 || state.VisibleMonth != 11 {
			t.Fatalf("InputPicker.State: %+v %v", state, err)
		}
		target, err := picker.Target()
		if err != nil || target == nil {
			t.Fatalf("InputPicker.Target: %v", err)
		}
		if id, _ := target.ID(); id != "d" {
			t.Fatalf("Target id: %q", id)
		}
		target.Close()
		_ = picker.Close()
		if target, _ := picker.Target(); target != nil {
			t.Fatal("关闭后不应有 Target")
		}
		var kuiErr *Error
		if err := text.OpenInputPicker(PickerNone); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotSupported {
			t.Fatalf("文本框没有面板: %v", err)
		}

		audio := document.Audio()
		available, err := audio.Available()
		if err != nil {
			t.Fatal(err)
		}
		if !available {
			if _, err := audio.Play("missing.wav", nil); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotSupported {
				t.Fatalf("无音频设备时 Play 应返回 ErrNotSupported: %v", err)
			}
		}
		if err := audio.SetGroup(AudioMusic, 0.5, true); err != nil {
			t.Fatal(err)
		}
		if volume, muted, _ := audio.Group(AudioMusic); volume != 0.5 || !muted {
			t.Fatalf("Group: %v %v", volume, muted)
		}
		if _, ok, _ := audio.Info(12345); ok {
			t.Fatal("未知句柄")
		}
		if active, _, _ := audio.Counts(AudioAllGroups); active != 0 {
			t.Fatalf("Counts: %d", active)
		}

		// 第 13 节：媒体与 Lottie 播放控制。
		media, _ := document.GetElementByID("a")
		defer media.Close()
		_ = media.SetMediaLoop(true)
		_ = media.SetMediaAutoplay(true)
		if loop, autoplay, ended, err := media.MediaPlaybackFlags(); err != nil || !loop || !autoplay || ended {
			t.Fatalf("MediaPlaybackFlags: %v %v %v %v", loop, autoplay, ended, err)
		}
		if err := media.MediaLoad(); err != nil {
			t.Fatal(err)
		}
		lottie, _ := document.GetElementByID("l")
		defer lottie.Close()
		_ = lottie.SetLottieSpeed(2)
		_ = lottie.SetLottieLoop(true)
		_ = lottie.LottieSeek(0.25)
		if state, err := lottie.LottieState(); err != nil || state.Speed != 2 || !state.Looping || state.Position != 0.25 || state.Load != LottieIdle {
			t.Fatalf("LottieState: %+v %v", state, err)
		}
		_ = lottie.LottieStop()
		if state, _ := lottie.LottieState(); state.Position != 0 {
			t.Fatalf("stop 应回到开头: %+v", state)
		}
		if err := media.LottieStop(); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotSupported {
			t.Fatalf("<audio> 不是 Lottie: %v", err)
		}

		// 第 14 节：扩展与 DevTools 生命周期。
		if _, err := document.InitExtensions(t.TempDir()); err != nil {
			t.Fatal(err)
		}
		if loaded, err := document.LoadExtension(filepath.Join(".", "no-such-extension"+filepath.Ext(RuntimeLibraryName))); err != nil || loaded {
			t.Fatalf("LoadExtension 不存在的 DLL: %v %v", loaded, err)
		}
		if err := document.UnloadExtension("kui-no-such"); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotFound {
			t.Fatalf("UnloadExtension 未知标签: %v", err)
		}
		_ = document.WatchExtensions(t.TempDir())
		// Linux / macOS 的内核还没有目录监视（T-DEBT-POSIX-EXTENSION-HOT-RELOAD-NOT-IMPLEMENTED），如实报告未监视：
		// 断言它当绊线，内核实现后这里会失败，届时同一提交改回与 Windows 相同的断言。
		if _, watching, _ := document.ExtensionState(); watching != (runtime.GOOS == "windows") {
			t.Fatalf("ExtensionState watching=%v on %s", watching, runtime.GOOS)
		}
		_ = document.StopWatchingExtensions()
		if count, watching, _ := document.ExtensionState(); watching || count != 0 {
			t.Fatalf("ExtensionState: %d %v", count, watching)
		}
		if started, err := window.StartDevTools(0); err == nil && started {
			if running, port, _, _ := window.DevToolsState(); !running || port == 0 {
				t.Fatalf("DevToolsState: %v %d", running, port)
			}
			_ = window.StopDevTools()
			if running, _, _, _ := window.DevToolsState(); running {
				t.Fatal("StopDevTools 之后不应在运行")
			}
		}
	})
}

// C ABI 第 15 节（3.61.8）：页内拖放数据，用注入的鼠标输入走一遍完整拖放。
func TestDragDropData(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-drag", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		if err := document.LoadHTML(`<html><body style="margin:0"><div id="src" draggable="true" style="position:absolute;left:10px;top:10px;width:80px;height:80px">s</div><div id="dst" style="position:absolute;left:200px;top:10px;width:80px;height:80px">d</div></body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.01)
		}
		source, err := document.GetElementByID("src")
		if err != nil || source == nil {
			t.Fatalf("src: %v", err)
		}
		defer source.Close()
		target, err := document.GetElementByID("dst")
		if err != nil || target == nil {
			t.Fatalf("dst: %v", err)
		}
		defer target.Close()

		started := 0
		if _, err := source.AddEventListener("dragstart", func(e *Event) {
			started++
			_ = e.SetDragData("text/plain", "hello drag")
			_ = e.SetDragData("application/x-kui", "42")
			_ = e.SetEffectAllowed(EffectCopy)
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := target.AddEventListener("dragover", func(e *Event) {
			_ = e.PreventDefault()
			_ = e.SetDropEffect(DropCopy)
		}); err != nil {
			t.Fatal(err)
		}
		type dropResult struct {
			data                string
			found, missingFound bool
			types               []string
			allowed             EffectAllowed
			effect              DropEffect
			afterOne, afterAll  []string
		}
		var got *dropResult
		if _, err := target.AddEventListener("drop", func(e *Event) {
			r := &dropResult{}
			r.data, r.found, _ = e.DragData("text/plain")
			_, r.missingFound, _ = e.DragData("missing/type")
			r.types, _ = e.DragTypes()
			r.allowed, _ = e.EffectAllowed()
			r.effect, _ = e.DropEffect()
			_ = e.ClearDragData("application/x-kui")
			r.afterOne, _ = e.DragTypes()
			_ = e.ClearAllDragData()
			r.afterAll, _ = e.DragTypes()
			got = r
		}); err != nil {
			t.Fatal(err)
		}
		var plainErr error
		if _, err := target.AddEventListener("mousedown", func(e *Event) { _, plainErr = e.DragTypes() }); err != nil {
			t.Fatal(err)
		}

		mouse := func(kind MouseEventType, x float32, clicks int) {
			if err := document.DispatchMouse(kind, x, 50, MouseButtonLeft, 0, clicks, 0, 0); err != nil {
				t.Fatal(err)
			}
		}
		mouse(MousePressed, 50, 1)
		for x := float32(60); x <= 240; x += 30 {
			mouse(MouseMoved, x, 0)
		}
		mouse(MouseReleased, 240, 1)
		if started != 1 || got == nil {
			t.Fatalf("拖放没有完成：dragstart=%d drop=%v", started, got != nil)
		}
		if got.data != "hello drag" || !got.found || got.missingFound {
			t.Fatalf("DragData: %+v", got)
		}
		if len(got.types) != 2 || got.types[0] != "text/plain" || got.types[1] != "application/x-kui" {
			t.Fatalf("DragTypes: %v", got.types)
		}
		if got.allowed != EffectCopy || got.effect != DropCopy {
			t.Fatalf("EffectAllowed / DropEffect: %v %v", got.allowed, got.effect)
		}
		if len(got.afterOne) != 1 || got.afterOne[0] != "text/plain" || len(got.afterAll) != 0 {
			t.Fatalf("ClearDragData / ClearAllDragData: %v %v", got.afterOne, got.afterAll)
		}

		mouse(MousePressed, 240, 1)
		mouse(MouseReleased, 240, 1)
		var kuiErr *Error
		if !errors.As(plainErr, &kuiErr) || kuiErr.Status != ErrNotSupported {
			t.Fatalf("普通鼠标事件应返回 ErrNotSupported：%v", plainErr)
		}
	})
}

// <kui-context-menu>：注入右键打开声明式菜单，menu-open 的 RelatedTarget 是被右键的元素，按在外面关闭。
func TestDeclarativeContextMenuRelatedTarget(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-context-menu", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		if err := document.LoadHTML(`<html><body style="margin:0"><div id="row" style="width:200px;height:60px">row</div>` +
			`<kui-context-menu id="menu" for="row"><kui-menu-item text="Open"></kui-menu-item></kui-context-menu></body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.01)
		}
		menu, err := document.GetElementByID("menu")
		if err != nil || menu == nil {
			t.Fatalf("menu: %v", err)
		}
		defer menu.Close()
		related := ""
		if _, err := menu.AddEventListener("menu-open", func(e *Event) {
			if element, err := e.RelatedTarget(); err == nil && element != nil {
				related, _ = element.ID()
				element.Close()
			}
		}); err != nil {
			t.Fatal(err)
		}
		press := func(kind MouseEventType, x, y float32, button MouseButton) {
			if err := document.DispatchMouse(kind, x, y, button, 0, 1, 0, 0); err != nil {
				t.Fatal(err)
			}
		}
		press(MousePressed, 20, 20, MouseButtonRight)
		press(MouseReleased, 20, 20, MouseButtonRight)
		if open, _ := menu.HasAttribute("open"); !open || related != "row" {
			t.Fatalf("右键应打开菜单且 RelatedTarget 是被右键的元素：open=%v related=%q", open, related)
		}
		press(MousePressed, 300, 300, MouseButtonLeft)
		press(MouseReleased, 300, 300, MouseButtonLeft)
		if open, _ := menu.HasAttribute("open"); open {
			t.Fatal("按在外面应关闭菜单")
		}
	})
}

// 表单控件的当前状态（读写），与内容属性 value="…" 的初值区分开。
func TestFormControlState(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-form-state", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		if err := document.LoadHTML(`<html><body><input id="name" type="text" value="initial">` +
			`<select id="size"><option value="s">S</option><option value="m" selected>M</option><option value="l">L</option></select>` +
			`<input id="agree" type="checkbox"><progress id="prog" value="3" max="10"></progress><div id="plain" value="attr"></div>` +
			`</body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.01)
		}
		get := func(id string) *Element {
			element, err := document.GetElementByID(id)
			if err != nil || element == nil {
				t.Fatalf("GetElementByID(%s): %v", id, err)
			}
			return element
		}
		name, size, agree, prog, plain := get("name"), get("size"), get("agree"), get("prog"), get("plain")
		defer name.Close()
		defer size.Close()
		defer agree.Close()
		defer prog.Close()
		defer plain.Close()

		if value, err := name.Value(); err != nil || value != "initial" {
			t.Fatalf("初值: %q %v", value, err)
		}
		if err := name.SetValue("Ada"); err != nil {
			t.Fatal(err)
		}
		if value, _ := name.Value(); value != "Ada" {
			t.Fatalf("当前值应是写入的 Ada，得到 %q", value)
		}
		if attribute, _ := name.Attribute("value"); attribute != "initial" {
			t.Fatalf("内容属性仍应是初值，得到 %q", attribute)
		}
		if index, err := size.SelectedIndex(); err != nil || index != 1 {
			t.Fatalf("select 初始下标: %d %v", index, err)
		}
		if err := size.SetSelectedIndex(2); err != nil {
			t.Fatal(err)
		}
		if value, _ := size.Value(); value != "l" {
			t.Fatalf("select 值: %q", value)
		}
		if err := agree.SetChecked(true); err != nil {
			t.Fatal(err)
		}
		if checked, err := agree.Checked(); err != nil || !checked {
			t.Fatalf("checked: %v %v", checked, err)
		}
		if err := prog.SetValue("7.5"); err != nil {
			t.Fatal(err)
		}
		if value, _ := prog.Value(); value != "7.5" {
			t.Fatalf("progress 值: %q", value)
		}
		if err := prog.SetValue("abc"); err == nil {
			t.Fatal("非数字应被拒绝")
		}
		var apiError *Error
		if _, err := plain.Value(); !errors.As(err, &apiError) || apiError.Status != ErrNotSupported {
			t.Fatalf("普通元素没有当前值状态: %v", err)
		}
		if _, err := name.SelectedIndex(); !errors.As(err, &apiError) || apiError.Status != ErrNotSupported {
			t.Fatalf("文本框没有 selectedIndex: %v", err)
		}
	})
}

// 宿主集成类接口：创建时隐藏、单帧渲染、滚入视图与惯性滚动、基准路径、脚本定时器、控制台与诊断。
func TestHostIntegration(t *testing.T) {
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
		radius := float32(8)
		window, err := NewWindow(WindowProps{Title: "kui-go-host", Width: 320, Height: 200, Frameless: true,
			Sizeable: true, Hidden: true, FrameRadius: &radius})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		if visible, _ := window.IsVisible(); visible {
			t.Fatal("Hidden 创建的窗口不应显示")
		}
		if err := window.ProcessFrame(); err != nil {
			t.Fatalf("ProcessFrame: %v", err)
		}
		if err := window.Show(); err != nil {
			t.Fatal(err)
		}
		if visible, _ := window.IsVisible(); !visible {
			t.Fatal("Show 之后应显示")
		}
		var apiError *Error
		if err := window.SetDevToolsTokenAuth(false); !errors.As(err, &apiError) || apiError.Status != ErrNotFound {
			t.Fatalf("DevTools 没启动时令牌开关应返回 NOT_FOUND: %v", err)
		}
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		if err := document.LoadHTML(`<html><body style="margin:0"><div id="box" style="width:200px;height:100px;overflow:auto">` +
			`<div style="height:1000px"></div><p id="end" style="margin:0;height:20px">end</p></div></body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.01)
		}
		box, _ := document.GetElementByID("box")
		end, _ := document.GetElementByID("end")
		if box == nil || end == nil {
			t.Fatal("找不到滚动夹具")
		}
		defer box.Close()
		defer end.Close()
		if err := end.ScrollIntoView(true); err != nil {
			t.Fatal(err)
		}
		if metrics, _ := box.ScrollMetrics(); metrics.Y <= 500 {
			t.Fatalf("滚入视图后容器应已向下滚动: %+v", metrics)
		}
		if err := box.ScrollTo(0, 0); err != nil {
			t.Fatal(err)
		}
		if err := box.StartMomentumScroll(0, 0.5); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 20; i++ {
			_ = WaitEvents(0.02)
		}
		if metrics, _ := box.ScrollMetrics(); metrics.Y <= 0 {
			t.Fatalf("惯性滚动应让容器动起来: %+v", metrics)
		}

		// 捕获阶段监听：从根往下先执行，目标处捕获在前、普通在后，再冒泡。
		var order []string
		tag := func(name string) func(*Event) { return func(*Event) { order = append(order, name) } }
		var ids [2]uint64
		var capture, plain uint64
		var addErr error
		if _, addErr = box.AddEventListener("probe", tag("box")); addErr != nil {
			t.Fatal(addErr)
		}
		if capture, addErr = box.AddCaptureEventListener("probe", tag("box-capture")); addErr != nil {
			t.Fatal(addErr)
		}
		if plain, addErr = end.AddEventListener("probe", tag("end")); addErr != nil {
			t.Fatal(addErr)
		}
		if ids[0], addErr = end.AddCaptureEventListener("probe", tag("end-capture")); addErr != nil {
			t.Fatal(addErr)
		}
		if _, err := end.DispatchEvent("probe", true, true); err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(order, " "); got != "box-capture end-capture end box" {
			t.Fatalf("捕获阶段监听的顺序: %q", got)
		}
		_ = box.RemoveEventListener("probe", capture)
		_ = end.RemoveEventListener("probe", plain)
		_ = end.RemoveEventListener("probe", ids[0])
		if _, err := end.AddCaptureEventListener("probe", nil); err == nil {
			t.Fatal("nil 回调应被拒绝")
		}

		if err := document.SetBasePath("assets/pages/"); err != nil {
			t.Fatal(err)
		}
		if path, err := document.BasePath(); err != nil || path != "assets/pages/" {
			t.Fatalf("基准路径读回: %q %v", path, err)
		}
		ratio, _ := window.DevicePixelRatio()
		if err := document.SetRenderDevicePixelRatio(ratio); err != nil {
			t.Fatalf("按原值设置渲染 DPR: %v", err)
		}

		engine, err := NewScriptEngine()
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		_ = engine.BindDocument(document)
		if result, err := engine.Eval("host_timer_fired = false; setTimeout(function() host_timer_fired = true end, 50)", "test"); err != nil || !result.Success {
			t.Fatalf("排定定时器: %v %+v", err, result)
		}
		_ = engine.TickTimers(0)
		if result, _ := engine.Eval("return tostring(host_timer_fired)", "test"); result.Value.String != "false" {
			t.Fatalf("定时器不该提前触发: %+v", result)
		}
		_ = engine.TickTimers(100)
		if result, _ := engine.Eval("return tostring(host_timer_fired)", "test"); result.Value.String != "true" {
			t.Fatalf("到期后的 tick 应触发定时器: %+v", result)
		}
		_ = engine.UnbindDocument()

		if err := SetConsoleEnabled(true); err != nil {
			t.Fatal(err)
		}
		if err := SetConsoleStderrOutput(false); err != nil {
			t.Fatal(err)
		}
		if err := SetConsoleMinLevel(LogDebug); err != nil {
			t.Fatal(err)
		}
		if err := SetConsoleMinLevel(LogLevel(9)); !errors.As(err, &apiError) || apiError.Status != ErrInvalidArg {
			t.Fatalf("未知日志级别应被拒绝: %v", err)
		}
		if _, err := InstallStackTrace(StackTraceConfig{OutputMode: StackTraceStderr}); err != nil {
			t.Fatal(err)
		}
		if err := UninstallStackTrace(); err != nil {
			t.Fatal(err)
		}
		if _, err := InstallCrashReporter(CrashReporterConfig{}); err != nil {
			t.Fatal(err)
		}
		if err := UninstallCrashReporter(); err != nil {
			t.Fatal(err)
		}
		if err := DisableIME(); err != nil && (!errors.As(err, &apiError) || apiError.Status != ErrNotSupported) {
			t.Fatalf("DisableIME: %v", err)
		}
	})
}
