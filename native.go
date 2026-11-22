//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

/*
#cgo CFLAGS: -std=c11 -I${SRCDIR}
#cgo linux LDFLAGS: -ldl
#include <stdlib.h>
#include "loader.h"
#include "callbacks.h"
*/
import "C"

import (
	"runtime"
	"runtime/cgo"
	"sync"
	"unsafe"
)

func inputString(value string) (*C.char, C.size_t, func()) {
	raw := C.CString(value)
	return raw, C.size_t(len(value)), func() { C.free(unsafe.Pointer(raw)) }
}

func statusError(status C.KuiStatus, context string) error {
	if status == C.KUI_OK {
		return nil
	}
	message := C.kui_status_message(status)
	text := "unknown error"
	if message != nil {
		text = C.GoString(message)
	}
	return &Error{Status: Status(status), Context: context, Message: text}
}

func bufferStatus(status C.KuiStatus, context string) error {
	if status == C.KUI_OK || status == C.KUI_ERR_BUFFER_TOO_SMALL {
		return nil
	}
	return statusError(status, context)
}

func loaderReady() error {
	if C.kui_go_is_loaded() != 0 {
		return nil
	}
	return &Error{Status: ErrNotSupported, Context: "runtime", Message: "the KUI C runtime (" + RuntimeLibraryName + ") is not loaded"}
}

// Load 解析并冻结全部 C ABI 导出（冻结清单 sdk/c/abi/kui_c_3.exports）。path 为空时加载可执行文件所在目录中的
// RuntimeLibraryName（Windows kui_c.dll、Linux libkui_c.so、macOS libkui_c.dylib），找不到再交给系统的库搜索规则。
func Load(path string) error {
	var raw *C.char
	if path != "" {
		var release func()
		raw, _, release = inputString(path)
		defer release()
	}
	if C.kui_go_load(raw) == 0 {
		return &Error{Status: ErrPluginMissing, Context: "Load", Message: C.GoString(C.kui_go_loader_error())}
	}
	return nil
}

func RuntimeLoaded() bool { return C.kui_go_is_loaded() != 0 }

// osThreadID 是当前 OS 线程的 ID（测试用来核对回调在哪个线程上执行）。
func osThreadID() uint64 { return uint64(C.kui_go_os_thread_id()) }

func RuntimeVersion() (Version, error) {
	if err := loaderReady(); err != nil {
		return Version{}, err
	}
	v := C.kui_get_version()
	return Version{Major: int(v.major), Minor: int(v.minor), Patch: int(v.patch)}, nil
}

// Init 将当前 goroutine 固定到 OS 线程并初始化平台。Shutdown 必须在同一 goroutine 调用。
func Init() error {
	if !RuntimeLoaded() {
		if err := Load(""); err != nil {
			return err
		}
	}
	runtime.LockOSThread()
	if err := statusError(C.kui_platform_init(), "Init"); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	return nil
}

func Shutdown() error {
	err := statusError(C.kui_platform_shutdown(), "Shutdown")
	runtime.UnlockOSThread()
	return err
}

func Run() (int, error) {
	var code C.int
	err := statusError(C.kui_platform_run(&code), "Run")
	return int(code), err
}
func Quit(code int) error { return statusError(C.kui_platform_quit(C.int(code)), "Quit") }
func PollEvents() error   { return statusError(C.kui_platform_poll_events(), "PollEvents") }
func WaitEvents(seconds float64) error {
	return statusError(C.kui_platform_wait_events(C.double(seconds)), "WaitEvents")
}
func SetF12DevToolsEnabled(enabled bool) error {
	return statusError(C.kui_platform_set_f12_devtools_enabled(boolInt(enabled)), "SetF12DevToolsEnabled")
}

func boolInt(value bool) C.int {
	if value {
		return 1
	}
	return 0
}

type timerState struct {
	fn   func()
	once bool
	id   uint64
}

var timerHandles sync.Map

// A panic must never unwind through a C callback frame. Native callback
// trampolines have no Go panic representation, so contain it at the binding
// boundary after all callback-local defers have released their handles.
func containCallbackPanic() {
	_ = recover()
}

//export goKuiTaskCallback
func goKuiTaskCallback(handle C.uintptr_t) {
	defer containCallbackPanic()
	h := cgo.Handle(handle)
	state, ok := h.Value().(*timerState)
	if !ok {
		return
	}
	if state.once {
		timerHandles.Delete(state.id)
		defer h.Delete()
	}
	state.fn()
}

// goKuiTaskDiscardCallback 在 Shutdown 丢弃尚未执行的 PostTask 时由 C ABI 调用（在调用 Shutdown 的线程上），
// 只负责释放 handle；任务本身不执行。
//
//export goKuiTaskDiscardCallback
func goKuiTaskDiscardCallback(handle C.uintptr_t) {
	defer containCallbackPanic()
	cgo.Handle(handle).Delete()
}

func addTimer(fn func(), delay uint32, interval bool) (uint64, error) {
	if fn == nil {
		return 0, &Error{Status: ErrInvalidArg, Context: "timer", Message: "nil callback"}
	}
	state := &timerState{fn: fn, once: !interval}
	h := cgo.NewHandle(state)
	var id C.KuiTimerId
	var status C.KuiStatus
	if interval {
		status = C.kui_platform_set_interval(C.kui_go_task_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(h)), C.uint32_t(delay), &id)
	} else {
		status = C.kui_platform_set_timeout(C.kui_go_task_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(h)), C.uint32_t(delay), &id)
	}
	if err := statusError(status, "timer"); err != nil {
		h.Delete()
		return 0, err
	}
	state.id = uint64(id)
	timerHandles.Store(state.id, h)
	return state.id, nil
}

func SetTimeout(fn func(), delayMS uint32) (uint64, error)     { return addTimer(fn, delayMS, false) }
func SetInterval(fn func(), intervalMS uint32) (uint64, error) { return addTimer(fn, intervalMS, true) }
func ClearTimer(id uint64) error {
	err := statusError(C.kui_platform_clear_timer(C.KuiTimerId(id)), "ClearTimer")
	if value, ok := timerHandles.LoadAndDelete(id); ok {
		value.(cgo.Handle).Delete()
	}
	return err
}

// PostTask 把 fn 交给 UI 线程（调用 Init 的那个 goroutine 所在的 OS 线程）执行，可以从任何 goroutine 调用，
// 用于把后台工作的结果带回界面。fn 在事件循环（Run / WaitEvents / PollEvents）里异步执行，同一 goroutine
// 投递的任务按投递顺序执行。Init 之前或 Shutdown 开始之后调用返回错误（Status 为 ErrNotSupported），fn 不会执行；
// Shutdown 时还没执行的任务被丢弃（不执行）。
func PostTask(fn func()) error {
	if fn == nil {
		return &Error{Status: ErrInvalidArg, Context: "PostTask", Message: "nil callback"}
	}
	state := &timerState{fn: fn, once: true}
	h := cgo.NewHandle(state)
	if err := statusError(C.kui_platform_post_to_main_thread(C.kui_go_task_trampoline(), C.kui_go_task_discard_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(h))), "PostTask"); err != nil {
		h.Delete()
		return err
	}
	return nil
}

// RequestIdle 在事件循环下次空闲时于 UI 线程执行 fn 一次。只能在 UI 线程调用；要从别的 goroutine
// 交回 UI 线程，用 PostTask。Shutdown 时还没执行的请求不会执行（C ABI 对 request_idle 没有丢弃回调，
// 这一个 handle 随进程结束回收）。
func RequestIdle(fn func()) error {
	if fn == nil {
		return &Error{Status: ErrInvalidArg, Context: "RequestIdle", Message: "nil callback"}
	}
	state := &timerState{fn: fn, once: true}
	h := cgo.NewHandle(state)
	if err := statusError(C.kui_platform_request_idle(C.kui_go_task_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(h))), "RequestIdle"); err != nil {
		h.Delete()
		return err
	}
	return nil
}

func monitorFromC(value C.KuiCMonitorInfo, name string) MonitorInfo {
	return MonitorInfo{
		Name:     name,
		Bounds:   Rect{X: float32(value.bounds_x), Y: float32(value.bounds_y), Width: float32(value.bounds_width), Height: float32(value.bounds_height)},
		WorkArea: Rect{X: float32(value.work_x), Y: float32(value.work_y), Width: float32(value.work_width), Height: float32(value.work_height)},
		Scale:    float32(value.scale), Primary: value.is_primary != 0,
	}
}

func PrimaryMonitor() (MonitorInfo, error) {
	var value C.KuiCMonitorInfo
	err := statusError(C.kui_platform_get_primary_monitor(&value), "PrimaryMonitor")
	return monitorFromC(value, ""), err
}

func Monitors() ([]MonitorInfo, error) {
	var count C.size_t
	if err := statusError(C.kui_platform_get_monitor_count(&count), "Monitors"); err != nil {
		return nil, err
	}
	result := make([]MonitorInfo, 0, int(count))
	for index := C.size_t(0); index < count; index++ {
		var value C.KuiCMonitorInfo
		var size C.size_t
		status := C.kui_platform_get_monitor(index, &value, nil, 0, &size)
		if err := bufferStatus(status, "Monitors"); err != nil {
			return nil, err
		}
		buffer := make([]byte, int(size)+1)
		if err := statusError(C.kui_platform_get_monitor(index, &value, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "Monitors"); err != nil {
			return nil, err
		}
		result = append(result, monitorFromC(value, string(buffer[:int(size)])))
	}
	return result, nil
}

func PrimaryMonitorPhysicalSize() (Size, error) {
	var width, height C.float
	err := statusError(C.kui_platform_get_primary_monitor_physical_size(&width, &height), "PrimaryMonitorPhysicalSize")
	return Size{Width: float32(width), Height: float32(height)}, err
}

func readPlatformString(call func(*C.char, C.size_t, *C.size_t) C.KuiStatus, context string) (string, error) {
	var size C.size_t
	if err := bufferStatus(call(nil, 0, &size), context); err != nil {
		return "", err
	}
	buffer := make([]byte, int(size)+1)
	if err := statusError(call((*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), context); err != nil {
		return "", err
	}
	return string(buffer[:int(size)]), nil
}

func ClipboardText() (string, error) {
	return readPlatformString(func(buffer *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_platform_get_clipboard_text(buffer, capacity, out)
	}, "ClipboardText")
}
func SetClipboardText(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_platform_set_clipboard_text(raw, size), "SetClipboardText")
}
func HasClipboardText() (bool, error) {
	var value C.int
	err := statusError(C.kui_platform_has_clipboard_text(&value), "HasClipboardText")
	return value != 0, err
}
func OSName() (string, error) {
	return readPlatformString(func(buffer *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_platform_get_os_name(buffer, capacity, out)
	}, "OSName")
}
func OSVersion() (string, error) {
	return readPlatformString(func(buffer *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_platform_get_os_version(buffer, capacity, out)
	}, "OSVersion")
}

func fileDialog(options FileDialogOptions, call func(*C.KuiCFileDialogOptions, *C.char, C.size_t, *C.size_t) C.KuiStatus, context string) (string, bool, error) {
	title, titleLen, releaseTitle := inputString(options.Title)
	defer releaseTitle()
	defaultPath, defaultPathLen, releaseDefaultPath := inputString(options.DefaultPath)
	defer releaseDefaultPath()
	filters := make([]*C.char, len(options.Filters))
	for index, value := range options.Filters {
		filters[index] = C.CString(value)
		defer C.free(unsafe.Pointer(filters[index]))
	}
	var filterPointer **C.char
	if len(filters) != 0 {
		filterPointer = (**C.char)(unsafe.Pointer(&filters[0]))
	}
	rawOptions := C.KuiCFileDialogOptions{
		title: title, title_len: titleLen,
		default_path: defaultPath, default_path_len: defaultPathLen,
		filters: filterPointer, filter_count: C.size_t(len(filters)),
		allow_multiple: boolInt(options.AllowMultiple),
	}
	buffer := make([]byte, 32768)
	var size C.size_t
	status := call(&rawOptions, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size)
	if Status(status) == ErrNotFound {
		return "", false, nil
	}
	if err := statusError(status, context); err != nil {
		return "", false, err
	}
	if int(size) > len(buffer) {
		return "", false, &Error{Status: ErrBufferTooSmall, Context: context, Message: "dialog result exceeds Windows path buffer"}
	}
	return string(buffer[:int(size)]), true, nil
}

func OpenFileDialog(options FileDialogOptions) (string, bool, error) {
	return fileDialog(options, func(raw *C.KuiCFileDialogOptions, buffer *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_platform_open_file_dialog(raw, buffer, capacity, out)
	}, "OpenFileDialog")
}
func SaveFileDialog(options FileDialogOptions) (string, bool, error) {
	return fileDialog(options, func(raw *C.KuiCFileDialogOptions, buffer *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_platform_save_file_dialog(raw, buffer, capacity, out)
	}, "SaveFileDialog")
}
func SelectFolderDialog(titleValue string) (string, bool, error) {
	title, titleLen, release := inputString(titleValue)
	defer release()
	buffer := make([]byte, 32768)
	var size C.size_t
	status := C.kui_platform_select_folder_dialog(title, titleLen, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size)
	if Status(status) == ErrNotFound {
		return "", false, nil
	}
	if err := statusError(status, "SelectFolderDialog"); err != nil {
		return "", false, err
	}
	if int(size) > len(buffer) {
		return "", false, &Error{Status: ErrBufferTooSmall, Context: "SelectFolderDialog", Message: "dialog result exceeds Windows path buffer"}
	}
	return string(buffer[:int(size)]), true, nil
}

type Window struct {
	handle          unsafe.Pointer
	document        *Document
	resizeHandle    cgo.Handle
	windowListeners map[uint64]cgo.Handle
	desktop         *desktopState // 菜单 / 托盘 / 快捷键回调句柄（common_windows 之外的 desktop_windows.go）
}

func NewWindow(props WindowProps) (*Window, error) {
	title, titleLen, release := inputString(props.Title)
	defer release()
	rawProps := C.KuiWindowProps{
		title: title, title_len: titleLen,
		width: C.int(props.Width), height: C.int(props.Height),
		min_width: C.int(props.MinWidth), min_height: C.int(props.MinHeight),
		max_width: C.int(props.MaxWidth), max_height: C.int(props.MaxHeight),
		no_resize: boolInt(props.NoResize), no_decorations: boolInt(props.NoDecorations),
		transparent: boolInt(props.Transparent), always_on_top: boolInt(props.AlwaysOnTop),
		maximized: boolInt(props.Maximized), centered: boolInt(props.Centered),
		frameless: boolInt(props.Frameless), frameless_shadow: boolInt(props.FramelessShadow),
		opacity: C.float(props.Opacity),
	}
	extended := C.KuiWindowPropsEx{
		struct_size: C.size_t(unsafe.Sizeof(C.KuiWindowPropsEx{})), base: rawProps,
		sizeable: boolInt(props.Sizeable), hidden: boolInt(props.Hidden),
	}
	if props.FrameRadius != nil {
		extended.has_frame_radius = 1
		extended.frame_radius = C.float(*props.FrameRadius)
	}
	var raw *C.KuiWindow
	if err := statusError(C.kui_window_create_ex(&extended, &raw), "NewWindow"); err != nil {
		return nil, err
	}
	window := &Window{handle: unsafe.Pointer(raw), windowListeners: make(map[uint64]cgo.Handle)}
	doc, err := window.Document()
	if err != nil {
		C.kui_window_release(raw)
		return nil, err
	}
	window.document = doc
	return window, nil
}

func (w *Window) raw() *C.KuiWindow {
	if w == nil {
		return nil
	}
	return (*C.KuiWindow)(w.handle)
}
func (w *Window) Close() error {
	if w == nil || w.handle == nil {
		return nil
	}
	if w.resizeHandle != 0 {
		w.resizeHandle.Delete()
		w.resizeHandle = 0
	}
	for id, handle := range w.windowListeners {
		handle.Delete()
		delete(w.windowListeners, id)
	}
	if w.document != nil {
		_ = w.document.Close()
		w.document = nil
	}
	err := statusError(C.kui_window_release(w.raw()), "Window.Close")
	w.handle = nil
	w.desktop.release() // 窗口已销毁，它的菜单 / 托盘 / 快捷键回调不会再触发
	w.desktop = nil
	return err
}
func (w *Window) Document() (*Document, error) {
	if w.document != nil {
		return w.document, nil
	}
	var raw *C.KuiDocument
	if err := statusError(C.kui_window_get_document(w.raw(), &raw), "Window.Document"); err != nil {
		return nil, err
	}
	return &Document{handle: unsafe.Pointer(raw), state: &documentState{references: 1}}, nil
}
func (w *Window) Show() error { return statusError(C.kui_window_show(w.raw()), "Window.Show") }
func (w *Window) Hide() error { return statusError(C.kui_window_hide(w.raw()), "Window.Hide") }
func (w *Window) IsVisible() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_is_visible(w.raw(), &value), "Window.IsVisible")
	return value != 0, err
}
func (w *Window) SetTitle(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_window_set_title(w.raw(), raw, size), "Window.SetTitle")
}
func (w *Window) Title() (string, error) {
	var size C.size_t
	if err := bufferStatus(C.kui_window_get_title(w.raw(), nil, 0, &size), "Window.Title"); err != nil {
		return "", err
	}
	buffer := make([]byte, int(size)+1)
	if err := statusError(C.kui_window_get_title(w.raw(), (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "Window.Title"); err != nil {
		return "", err
	}
	return string(buffer[:int(size)]), nil
}
func (w *Window) Size() (int, int, error) {
	var width, height C.int
	err := statusError(C.kui_window_get_size(w.raw(), &width, &height), "Window.Size")
	return int(width), int(height), err
}
func (w *Window) Resize(width, height int) error {
	return statusError(C.kui_window_resize(w.raw(), C.int(width), C.int(height)), "Window.Resize")
}
func (w *Window) Position() (int, int, error) {
	var x, y C.int
	err := statusError(C.kui_window_get_position(w.raw(), &x, &y), "Window.Position")
	return int(x), int(y), err
}
func (w *Window) SetPosition(x, y int) error {
	return statusError(C.kui_window_set_position(w.raw(), C.int(x), C.int(y)), "Window.SetPosition")
}
func (w *Window) Minimize() error {
	return statusError(C.kui_window_minimize(w.raw()), "Window.Minimize")
}
func (w *Window) Maximize() error {
	return statusError(C.kui_window_maximize(w.raw()), "Window.Maximize")
}
func (w *Window) Restore() error { return statusError(C.kui_window_restore(w.raw()), "Window.Restore") }
func (w *Window) IsMinimized() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_is_minimized(w.raw(), &value), "Window.IsMinimized")
	return value != 0, err
}
func (w *Window) IsMaximized() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_is_maximized(w.raw(), &value), "Window.IsMaximized")
	return value != 0, err
}
func (w *Window) SetAlwaysOnTop(value bool) error {
	return statusError(C.kui_window_set_always_on_top(w.raw(), boolInt(value)), "Window.SetAlwaysOnTop")
}
func (w *Window) SetOpacity(value float32) error {
	return statusError(C.kui_window_set_opacity(w.raw(), C.float(value)), "Window.SetOpacity")
}
func (w *Window) DevicePixelRatio() (float32, error) {
	var value C.float
	err := statusError(C.kui_window_get_device_pixel_ratio(w.raw(), &value), "Window.DevicePixelRatio")
	return float32(value), err
}
func (w *Window) CloseRequest() error {
	return statusError(C.kui_window_close(w.raw()), "Window.CloseRequest")
}
func (w *Window) ShouldClose() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_should_close(w.raw(), &value), "Window.ShouldClose")
	return value != 0, err
}
func (w *Window) SetFramelessSizeable(value bool) error {
	return statusError(C.kui_window_set_frameless_sizeable(w.raw(), boolInt(value)), "Window.SetFramelessSizeable")
}
func (w *Window) SetFrameRadius(value float32) error {
	return statusError(C.kui_window_set_frame_radius(w.raw(), C.float(value)), "Window.SetFrameRadius")
}
func (w *Window) ClearFrameRadius() error {
	return statusError(C.kui_window_clear_frame_radius(w.raw()), "Window.ClearFrameRadius")
}
func (w *Window) SetScriptEngine(engine *ScriptEngine) error {
	var raw *C.KuiScriptEngine
	if engine != nil {
		raw = engine.raw()
	}
	return statusError(C.kui_window_set_script_engine(w.raw(), raw), "Window.SetScriptEngine")
}

// NativeHandle 返回原生窗口句柄：Windows 上是 HWND，macOS 上是窗口的内容视图 NSView*（不是 NSWindow*），
// Linux 上是 X11 Window（XID）。
func (w *Window) NativeHandle() (uintptr, error) {
	var value unsafe.Pointer
	err := statusError(C.kui_window_get_native_handle(w.raw(), &value), "Window.NativeHandle")
	return uintptr(value), err
}
func (w *Window) StartDevTools(port uint16) (bool, error) {
	var started C.int
	err := statusError(C.kui_window_start_devtools(w.raw(), C.uint16_t(port), &started), "Window.StartDevTools")
	return started != 0, err
}
func (w *Window) DevToolsPort() (uint16, error) {
	var port C.uint16_t
	err := statusError(C.kui_window_get_devtools_port(w.raw(), &port), "Window.DevToolsPort")
	return uint16(port), err
}

// StopDevTools 停止窗口的 DevTools 服务；StartDevTools 可再次启动。
func (w *Window) StopDevTools() error {
	return statusError(C.kui_window_stop_devtools(w.raw()), "Window.StopDevTools")
}

// DevToolsState 返回 DevTools 是否在运行、端口与连接数（停止时端口与连接数为 0）。访问 token 在 DevToolsURL 里。
func (w *Window) DevToolsState() (running bool, port uint16, clients uint32, err error) {
	var rawRunning C.int
	var rawPort C.uint16_t
	var rawClients C.uint32_t
	err = statusError(C.kui_window_get_devtools_state(w.raw(), &rawRunning, &rawPort, &rawClients), "Window.DevToolsState")
	return rawRunning != 0, uint16(rawPort), uint32(rawClients), err
}

func (w *Window) DevToolsURL() (string, error) {
	var size C.size_t
	if err := bufferStatus(C.kui_window_get_devtools_url(w.raw(), nil, 0, &size), "Window.DevToolsURL"); err != nil {
		return "", err
	}
	buffer := make([]byte, int(size)+1)
	if err := statusError(C.kui_window_get_devtools_url(w.raw(), (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "Window.DevToolsURL"); err != nil {
		return "", err
	}
	return string(buffer[:int(size)]), nil
}

type documentState struct {
	customHandles []cgo.Handle
	references    int
}
type Document struct {
	handle unsafe.Pointer
	state  *documentState
}

func NewDocument() (*Document, error) {
	var raw *C.KuiDocument
	if err := statusError(C.kui_document_create(&raw), "NewDocument"); err != nil {
		return nil, err
	}
	return &Document{handle: unsafe.Pointer(raw), state: &documentState{references: 1}}, nil
}
func (d *Document) raw() *C.KuiDocument {
	if d == nil {
		return nil
	}
	return (*C.KuiDocument)(d.handle)
}
func (d *Document) Close() error {
	if d == nil || d.handle == nil {
		return nil
	}
	if d.state != nil {
		d.state.references--
		if d.state.references == 0 {
			for _, handle := range d.state.customHandles {
				handle.Delete()
			}
			d.state.customHandles = nil
		}
	}
	err := statusError(C.kui_document_release(d.raw()), "Document.Close")
	d.handle = nil
	d.state = nil
	return err
}
func (d *Document) Clone() (*Document, error) {
	if err := statusError(C.kui_document_retain(d.raw()), "Document.Clone"); err != nil {
		return nil, err
	}
	if d.state == nil {
		d.state = &documentState{references: 1}
	}
	d.state.references++
	return &Document{handle: d.handle, state: d.state}, nil
}
func (d *Document) LoadHTML(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_document_load_html(d.raw(), raw, size), "Document.LoadHTML")
}
func (d *Document) LoadFile(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_document_load_html_file(d.raw(), raw, size), "Document.LoadFile")
}
func (d *Document) AddStyleSheet(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_document_add_stylesheet(d.raw(), raw, size), "Document.AddStyleSheet")
}
func (d *Document) ClearStyleSheets() error {
	return statusError(C.kui_document_clear_stylesheets(d.raw()), "Document.ClearStyleSheets")
}

func (d *Document) wrapElement(raw *C.KuiElement, status C.KuiStatus, context string) (*Element, error) {
	if status == C.KUI_ERR_NOT_FOUND {
		return nil, nil
	}
	if err := statusError(status, context); err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	return &Element{handle: unsafe.Pointer(raw), document: d, listeners: make(map[uint64]cgo.Handle)}, nil
}
func (d *Document) Root() (*Element, error) {
	var raw *C.KuiElement
	status := C.kui_document_root(d.raw(), &raw)
	return d.wrapElement(raw, status, "Document.Root")
}
func (d *Document) Body() (*Element, error) {
	var raw *C.KuiElement
	status := C.kui_document_body(d.raw(), &raw)
	return d.wrapElement(raw, status, "Document.Body")
}
func (d *Document) Head() (*Element, error) {
	var raw *C.KuiElement
	status := C.kui_document_head(d.raw(), &raw)
	return d.wrapElement(raw, status, "Document.Head")
}
func (d *Document) GetElementByID(id string) (*Element, error) {
	rawID, size, release := inputString(id)
	defer release()
	var raw *C.KuiElement
	status := C.kui_document_get_element_by_id(d.raw(), rawID, size, &raw)
	return d.wrapElement(raw, status, "Document.GetElementByID")
}
func (d *Document) QuerySelector(selector string) (*Element, error) {
	rawSelector, size, release := inputString(selector)
	defer release()
	var raw *C.KuiElement
	status := C.kui_document_query_selector(d.raw(), rawSelector, size, &raw)
	return d.wrapElement(raw, status, "Document.QuerySelector")
}
func (d *Document) CreateElement(tag string) (*Element, error) {
	rawTag, size, release := inputString(tag)
	defer release()
	var raw *C.KuiElement
	status := C.kui_document_create_element(d.raw(), rawTag, size, &raw)
	return d.wrapElement(raw, status, "Document.CreateElement")
}

func elementArray(document *Document, first C.KuiStatus, count C.size_t, fill func(**C.KuiElement, C.size_t, *C.size_t) C.KuiStatus, context string) ([]*Element, error) {
	if err := bufferStatus(first, context); err != nil {
		return nil, err
	}
	if count == 0 {
		return []*Element{}, nil
	}
	bytes := C.size_t(unsafe.Sizeof(uintptr(0))) * count
	memory := C.malloc(bytes)
	if memory == nil {
		return nil, &Error{Status: ErrOutOfMemory, Context: context, Message: "allocation failed"}
	}
	defer C.free(memory)
	if err := statusError(fill((**C.KuiElement)(memory), count, &count), context); err != nil {
		return nil, err
	}
	rawItems := unsafe.Slice((**C.KuiElement)(memory), int(count))
	result := make([]*Element, 0, int(count))
	for _, raw := range rawItems {
		result = append(result, &Element{handle: unsafe.Pointer(raw), document: document, listeners: make(map[uint64]cgo.Handle)})
	}
	return result, nil
}

func (d *Document) QuerySelectorAll(selector string) ([]*Element, error) {
	rawSelector, size, release := inputString(selector)
	defer release()
	var count C.size_t
	first := C.kui_document_query_selector_all(d.raw(), rawSelector, size, nil, 0, &count)
	return elementArray(d, first, count, func(items **C.KuiElement, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_document_query_selector_all(d.raw(), rawSelector, size, items, capacity, out)
	}, "Document.QuerySelectorAll")
}

// Deprecated: 用 DispatchMouseEvent（MouseInjection）；本函数保留一个版本。
func (d *Document) DispatchMouse(kind MouseEventType, x, y float32, button MouseButton, modifiers, clickCount int, deltaX, deltaY float32) error {
	return statusError(C.kui_devtools_dispatch_mouse_event(d.raw(), C.KuiCMouseEventType(kind), C.float(x), C.float(y), C.KuiCMouseButton(button), C.int(modifiers), C.int(clickCount), C.float(deltaX), C.float(deltaY)), "Document.DispatchMouse")
}

// Deprecated: 用 DispatchKeyEvent（KeyInjection）；本函数保留一个版本。
func (d *Document) DispatchKey(kind KeyEventType, key, code string, modifiers int, text string, autoRepeat bool) error {
	rawKey, keyLen, releaseKey := inputString(key)
	defer releaseKey()
	rawCode, codeLen, releaseCode := inputString(code)
	defer releaseCode()
	rawText, textLen, releaseText := inputString(text)
	defer releaseText()
	return statusError(C.kui_devtools_dispatch_key_event(d.raw(), C.KuiCKeyEventType(kind), rawKey, keyLen, rawCode, codeLen, C.int(modifiers), rawText, textLen, boolInt(autoRepeat)), "Document.DispatchKey")
}
func (d *Document) DispatchDevToolsCommand(command string) (string, error) {
	rawCommand, _, release := inputString(command)
	defer release()
	var size C.size_t
	if err := bufferStatus(C.kui_devtools_dispatch_command(d.raw(), rawCommand, nil, 0, &size), "Document.DispatchDevToolsCommand"); err != nil {
		return "", err
	}
	buffer := make([]byte, int(size)+1)
	if err := statusError(C.kui_devtools_dispatch_command(d.raw(), rawCommand, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "Document.DispatchDevToolsCommand"); err != nil {
		return "", err
	}
	return string(buffer[:int(size)]), nil
}

// Deprecated: 用 ElementBoxModel（BoxModelInfo）；本函数保留一个版本。
func (d *Document) BoxModel(selector string) ([32]float32, Size, error) {
	var result [32]float32
	raw, _, release := inputString(selector)
	defer release()
	var width, height C.float
	err := statusError(C.kui_devtools_get_box_model(d.raw(), raw, (*C.float)(unsafe.Pointer(&result[0])), &width, &height), "Document.BoxModel")
	return result, Size{Width: float32(width), Height: float32(height)}, err
}

// 3.61.8 起的扩展生命周期（C ABI 第 14 节）；需要先 InitExtensions（空目录也行）。

// LoadExtension 加载单个扩展 DLL；加载失败时返回 false。
func (d *Document) LoadExtension(path string) (bool, error) {
	raw, size, release := inputString(path)
	defer release()
	var loaded C.int
	err := statusError(C.kui_document_load_extension(d.raw(), raw, size, &loaded), "Document.LoadExtension")
	return loaded != 0, err
}

// UnloadExtension 按自定义元素标签卸载扩展；没加载时返回 ErrNotFound。
func (d *Document) UnloadExtension(tagName string) error {
	raw, size, release := inputString(tagName)
	defer release()
	return statusError(C.kui_document_unload_extension(d.raw(), raw, size), "Document.UnloadExtension")
}

// ReloadExtension 卸载该 DLL 的扩展并重新加载。
func (d *Document) ReloadExtension(path string) (bool, error) {
	raw, size, release := inputString(path)
	defer release()
	var reloaded C.int
	err := statusError(C.kui_document_reload_extension(d.raw(), raw, size, &reloaded), "Document.ReloadExtension")
	return reloaded != 0, err
}

// UnloadAllExtensions 卸载全部扩展。
func (d *Document) UnloadAllExtensions() error {
	return statusError(C.kui_document_unload_all_extensions(d.raw()), "Document.UnloadAllExtensions")
}

// WatchExtensions 监视目录，DLL 变化时在下一帧自动重载。
func (d *Document) WatchExtensions(directory string) error {
	raw, size, release := inputString(directory)
	defer release()
	return statusError(C.kui_document_watch_extensions(d.raw(), raw, size), "Document.WatchExtensions")
}

// StopWatchingExtensions 停止监视。
func (d *Document) StopWatchingExtensions() error {
	return statusError(C.kui_document_watch_extensions(d.raw(), nil, 0), "Document.StopWatchingExtensions")
}

// ExtensionState 返回已加载的扩展数与是否在监视目录。
func (d *Document) ExtensionState() (count int, watching bool, err error) {
	var rawCount C.size_t
	var rawWatching C.int
	err = statusError(C.kui_document_get_extension_state(d.raw(), &rawCount, &rawWatching), "Document.ExtensionState")
	return int(rawCount), rawWatching != 0, err
}

func (d *Document) InitExtensions(directory string) (int, error) {
	raw, size, release := inputString(directory)
	defer release()
	var count C.size_t
	err := statusError(C.kui_document_init_extensions(d.raw(), raw, size, &count), "Document.InitExtensions")
	return int(count), err
}
func (d *Document) ExtensionLoaded(tag string) (bool, error) {
	raw, size, release := inputString(tag)
	defer release()
	var loaded C.int
	err := statusError(C.kui_document_extension_is_loaded(d.raw(), raw, size, &loaded), "Document.ExtensionLoaded")
	return loaded != 0, err
}
func (d *Document) ExtensionPerfStats(tag string) (ExtensionPerfStats, error) {
	rawTag, size, release := inputString(tag)
	defer release()
	var value C.KuiCExtensionPerfStats
	err := statusError(C.kui_document_extension_get_perf_stats(d.raw(), rawTag, size, &value), "Document.ExtensionPerfStats")
	return ExtensionPerfStats{LastTickMS: float64(value.last_tick_ms), LastPaintMS: float64(value.last_paint_ms), MaxTickMS: float64(value.max_tick_ms), MaxPaintMS: float64(value.max_paint_ms), TickCount: uint64(value.tick_count), PaintCount: uint64(value.paint_count), TotalTickMS: float64(value.total_tick_ms), TotalPaintMS: float64(value.total_paint_ms)}, err
}
func (d *Document) ResetExtensionPerfStats() error {
	return statusError(C.kui_document_extension_reset_perf_stats(d.raw()), "Document.ResetExtensionPerfStats")
}
func (d *Document) PrintToPDF(path string, options PDFOptions) (bool, error) {
	rawPath, pathLen, releasePath := inputString(path)
	defer releasePath()
	values := []string{options.Title, options.Author, options.Subject, options.Keywords, options.Creator, options.CreationDate}
	cvalues := make([]*C.char, len(values))
	for index, value := range values {
		cvalues[index] = C.CString(value)
		defer C.free(unsafe.Pointer(cvalues[index]))
	}
	rawOptions := C.KuiCPDFOptions{title: cvalues[0], author: cvalues[1], subject: cvalues[2], keywords: cvalues[3], creator: cvalues[4], creation_date: cvalues[5], page_width_pt: C.float(options.PageWidthPT), page_height_pt: C.float(options.PageHeightPT), rasterize_fonts: boolInt(options.RasterizeFonts), generate_outline: boolInt(options.GenerateOutline)}
	var written C.int
	err := statusError(C.kui_document_print_to_pdf(d.raw(), rawPath, pathLen, &rawOptions, &written), "Document.PrintToPDF")
	return written != 0, err
}

type Element struct {
	handle    unsafe.Pointer
	document  *Document
	listeners map[uint64]cgo.Handle
	borrowed  bool
}

func (e *Element) Clone() (*Element, error) {
	if err := statusError(C.kui_element_retain(e.raw()), "Element.Clone"); err != nil {
		return nil, err
	}
	return &Element{handle: e.handle, document: e.document, listeners: make(map[uint64]cgo.Handle)}, nil
}

func (e *Element) raw() *C.KuiElement {
	if e == nil {
		return nil
	}
	return (*C.KuiElement)(e.handle)
}
func (e *Element) Close() error {
	if e == nil || e.handle == nil {
		return nil
	}
	if e.borrowed {
		e.handle = nil
		return nil
	}
	for id, handle := range e.listeners {
		state := handle.Value().(*eventState)
		_ = e.RemoveEventListener(state.name, id)
	}
	err := statusError(C.kui_element_release(e.raw()), "Element.Close")
	e.handle = nil
	return err
}
func (e *Element) readString(call func(*C.char, C.size_t, *C.size_t) C.KuiStatus, context string) (string, error) {
	var size C.size_t
	if err := bufferStatus(call(nil, 0, &size), context); err != nil {
		return "", err
	}
	buffer := make([]byte, int(size)+1)
	if err := statusError(call((*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), context); err != nil {
		return "", err
	}
	return string(buffer[:int(size)]), nil
}
func (e *Element) TagName() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, size *C.size_t) C.KuiStatus {
		return C.kui_element_get_tag_name(e.raw(), buf, capacity, size)
	}, "Element.TagName")
}
func (e *Element) ID() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, size *C.size_t) C.KuiStatus {
		return C.kui_element_get_id(e.raw(), buf, capacity, size)
	}, "Element.ID")
}
func (e *Element) ClassName() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, size *C.size_t) C.KuiStatus {
		return C.kui_element_get_class_name(e.raw(), buf, capacity, size)
	}, "Element.ClassName")
}
func (e *Element) SetClassName(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_element_set_class_name(e.raw(), raw, size), "Element.SetClassName")
}
func (e *Element) SetAttribute(name, value string) error {
	rawName, nameLen, releaseName := inputString(name)
	defer releaseName()
	rawValue, valueLen, releaseValue := inputString(value)
	defer releaseValue()
	return statusError(C.kui_element_set_attribute(e.raw(), rawName, nameLen, rawValue, valueLen), "Element.SetAttribute")
}
func (e *Element) Attribute(name string) (string, error) {
	raw, size, release := inputString(name)
	defer release()
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_get_attribute(e.raw(), raw, size, buf, capacity, out)
	}, "Element.Attribute")
}
func (e *Element) HasAttribute(name string) (bool, error) {
	raw, size, release := inputString(name)
	defer release()
	var value C.int
	err := statusError(C.kui_element_has_attribute(e.raw(), raw, size, &value), "Element.HasAttribute")
	return value != 0, err
}
func (e *Element) RemoveAttribute(name string) error {
	raw, size, release := inputString(name)
	defer release()
	return statusError(C.kui_element_remove_attribute(e.raw(), raw, size), "Element.RemoveAttribute")
}

// Value 读表单控件的当前值（用户键入 / 选择的），不是内容属性 value="…" 的初值（那个用 Attribute("value")）：
// <input> / <textarea> / <select> / <output> 取 value，<progress> / <meter> 取数值的十进制文本。
// 其他元素返回 KUI_ERR_NOT_SUPPORTED。
func (e *Element) Value() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_get_value(e.raw(), buf, capacity, out)
	}, "Element.Value")
}

// SetValue 写控件的当前值；不派发 input / change 事件（与给 DOM 属性赋值一致）。
func (e *Element) SetValue(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_element_set_value(e.raw(), raw, size), "Element.SetValue")
}

// Checked 读 <input> 的选中状态（复选框 / 单选有意义）。其他元素返回 KUI_ERR_NOT_SUPPORTED。
func (e *Element) Checked() (bool, error) {
	var value C.int
	err := statusError(C.kui_element_get_checked(e.raw(), &value), "Element.Checked")
	return value != 0, err
}

// SetChecked 写 <input> 的选中状态；不派发事件。
func (e *Element) SetChecked(checked bool) error {
	flag := C.int(0)
	if checked {
		flag = 1
	}
	return statusError(C.kui_element_set_checked(e.raw(), flag), "Element.SetChecked")
}

// SelectedIndex 读 <select> 当前选中的选项下标（-1 表示没有）。其他元素返回 KUI_ERR_NOT_SUPPORTED。
func (e *Element) SelectedIndex() (int, error) {
	var value C.int32_t
	err := statusError(C.kui_element_get_selected_index(e.raw(), &value), "Element.SelectedIndex")
	return int(value), err
}

// SetSelectedIndex 写 <select> 当前选项（越界按 DOM 规则钳制）；不派发 change 事件。
func (e *Element) SetSelectedIndex(index int) error {
	return statusError(C.kui_element_set_selected_index(e.raw(), C.int32_t(index)), "Element.SetSelectedIndex")
}
func (e *Element) SetTextContent(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_element_set_text_content(e.raw(), raw, size), "Element.SetTextContent")
}
func (e *Element) TextContent() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_get_text_content(e.raw(), buf, capacity, out)
	}, "Element.TextContent")
}
func (e *Element) SetInnerHTML(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_element_set_inner_html(e.raw(), raw, size), "Element.SetInnerHTML")
}
func (e *Element) InnerHTML() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_get_inner_html(e.raw(), buf, capacity, out)
	}, "Element.InnerHTML")
}
func (e *Element) AddClass(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_element_class_list_add(e.raw(), raw, size), "Element.AddClass")
}
func (e *Element) RemoveClass(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_element_class_list_remove(e.raw(), raw, size), "Element.RemoveClass")
}
func (e *Element) ToggleClass(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_element_class_list_toggle(e.raw(), raw, size), "Element.ToggleClass")
}
func (e *Element) HasClass(value string) (bool, error) {
	raw, size, release := inputString(value)
	defer release()
	var found C.int
	err := statusError(C.kui_element_class_list_contains(e.raw(), raw, size, &found), "Element.HasClass")
	return found != 0, err
}

func (e *Element) QuerySelector(selector string) (*Element, error) {
	rawSelector, size, release := inputString(selector)
	defer release()
	var raw *C.KuiElement
	status := C.kui_element_query_selector(e.raw(), rawSelector, size, &raw)
	return e.document.wrapElement(raw, status, "Element.QuerySelector")
}
func (e *Element) QuerySelectorAll(selector string) ([]*Element, error) {
	rawSelector, size, release := inputString(selector)
	defer release()
	var count C.size_t
	first := C.kui_element_query_selector_all(e.raw(), rawSelector, size, nil, 0, &count)
	return elementArray(e.document, first, count, func(items **C.KuiElement, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_query_selector_all(e.raw(), rawSelector, size, items, capacity, out)
	}, "Element.QuerySelectorAll")
}
func (e *Element) related(call func(**C.KuiElement) C.KuiStatus, context string) (*Element, error) {
	var raw *C.KuiElement
	status := call(&raw)
	return e.document.wrapElement(raw, status, context)
}
func (e *Element) Parent() (*Element, error) {
	return e.related(func(out **C.KuiElement) C.KuiStatus { return C.kui_element_parent(e.raw(), out) }, "Element.Parent")
}
func (e *Element) FirstChild() (*Element, error) {
	return e.related(func(out **C.KuiElement) C.KuiStatus { return C.kui_element_first_child(e.raw(), out) }, "Element.FirstChild")
}
func (e *Element) LastChild() (*Element, error) {
	return e.related(func(out **C.KuiElement) C.KuiStatus { return C.kui_element_last_child(e.raw(), out) }, "Element.LastChild")
}
func (e *Element) NextSibling() (*Element, error) {
	return e.related(func(out **C.KuiElement) C.KuiStatus { return C.kui_element_next_sibling(e.raw(), out) }, "Element.NextSibling")
}
func (e *Element) PreviousSibling() (*Element, error) {
	return e.related(func(out **C.KuiElement) C.KuiStatus { return C.kui_element_previous_sibling(e.raw(), out) }, "Element.PreviousSibling")
}
func (e *Element) ChildCount() (int, error) {
	var count C.size_t
	err := statusError(C.kui_element_child_count(e.raw(), &count), "Element.ChildCount")
	return int(count), err
}
func (e *Element) AppendChild(child *Element) error {
	return statusError(C.kui_element_append_child(e.raw(), child.raw()), "Element.AppendChild")
}
func (e *Element) RemoveChild(child *Element) error {
	return statusError(C.kui_element_remove_child(e.raw(), child.raw()), "Element.RemoveChild")
}
func (e *Element) Remove() error { return statusError(C.kui_element_remove(e.raw()), "Element.Remove") }
func (e *Element) Focus() error  { return statusError(C.kui_element_focus(e.raw()), "Element.Focus") }
func (e *Element) Blur() error   { return statusError(C.kui_element_blur(e.raw()), "Element.Blur") }
func (e *Element) ScrollTo(left, top float32) error {
	return statusError(C.kui_element_set_scroll_position(e.raw(), C.float(left), C.float(top)), "Element.ScrollTo")
}
func (e *Element) ScrollBy(dx, dy float32) error {
	return statusError(C.kui_element_scroll_by(e.raw(), C.float(dx), C.float(dy)), "Element.ScrollBy")
}
func (e *Element) BoundingRect() (Rect, error) {
	var value C.KuiRect
	err := statusError(C.kui_element_get_bounding_rect(e.raw(), &value), "Element.BoundingRect")
	return Rect{X: float32(value.x), Y: float32(value.y), Width: float32(value.width), Height: float32(value.height)}, err
}
func (e *Element) ClientMetrics() (Rect, error) {
	var left, top, width, height C.float
	err := statusError(C.kui_element_get_client_metrics(e.raw(), &left, &top, &width, &height), "Element.ClientMetrics")
	return Rect{X: float32(left), Y: float32(top), Width: float32(width), Height: float32(height)}, err
}
func (e *Element) OffsetMetrics() (Rect, error) {
	var left, top, width, height C.float
	err := statusError(C.kui_element_get_offset_metrics(e.raw(), &left, &top, &width, &height), "Element.OffsetMetrics")
	return Rect{X: float32(left), Y: float32(top), Width: float32(width), Height: float32(height)}, err
}
func (e *Element) ScrollMetrics() (Rect, error) {
	var left, top, width, height C.float
	err := statusError(C.kui_element_get_scroll_metrics(e.raw(), &left, &top, &width, &height), "Element.ScrollMetrics")
	return Rect{X: float32(left), Y: float32(top), Width: float32(width), Height: float32(height)}, err
}

func (e *Element) SetStyleProperty(name, value string, important bool) error {
	rawName, nameLen, releaseName := inputString(name)
	defer releaseName()
	rawValue, valueLen, releaseValue := inputString(value)
	defer releaseValue()
	return statusError(C.kui_element_set_style_property(e.raw(), rawName, nameLen, rawValue, valueLen, boolInt(important)), "Element.SetStyleProperty")
}
func (e *Element) StyleProperty(name string) (string, error) {
	raw, size, release := inputString(name)
	defer release()
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_get_style_property(e.raw(), raw, size, buf, capacity, out)
	}, "Element.StyleProperty")
}
func (e *Element) RemoveStyleProperty(name string) error {
	raw, size, release := inputString(name)
	defer release()
	return statusError(C.kui_element_remove_style_property(e.raw(), raw, size), "Element.RemoveStyleProperty")
}
func (e *Element) StylePropertyImportant(name string) (bool, error) {
	raw, size, release := inputString(name)
	defer release()
	var value C.int
	err := statusError(C.kui_element_style_property_is_important(e.raw(), raw, size, &value), "Element.StylePropertyImportant")
	return value != 0, err
}
func (e *Element) StyleText() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_get_style_text(e.raw(), buf, capacity, out)
	}, "Element.StyleText")
}
func (e *Element) SetStyleText(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_element_set_style_text(e.raw(), raw, size), "Element.SetStyleText")
}
func (e *Element) StylePropertyNames() ([]string, error) {
	var count C.size_t
	if err := statusError(C.kui_element_get_style_property_count(e.raw(), &count), "Element.StylePropertyNames"); err != nil {
		return nil, err
	}
	result := make([]string, 0, int(count))
	for index := C.size_t(0); index < count; index++ {
		value, err := e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
			return C.kui_element_get_style_property_name(e.raw(), index, buf, capacity, out)
		}, "Element.StylePropertyNames")
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

func (e *Element) PositionRelativeTo(anchor *Element, placement string, options PositioningOptions) error {
	rawPlacement, size, release := inputString(placement)
	defer release()
	rawOptions := C.KuiCPositioningOptions{flip: boolInt(options.Flip), clamp_to_viewport: boolInt(options.ClampToViewport), offset: C.float(options.Offset), viewport_margin: C.float(options.ViewportMargin)}
	return statusError(C.kui_element_position_relative_to(e.raw(), anchor.raw(), rawPlacement, size, &rawOptions), "Element.PositionRelativeTo")
}
func (e *Element) MediaPlay() error {
	return statusError(C.kui_media_play(e.raw()), "Element.MediaPlay")
}
func (e *Element) MediaPause() error {
	return statusError(C.kui_media_pause(e.raw()), "Element.MediaPause")
}
func (e *Element) SetMediaSource(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_media_set_src(e.raw(), raw, size), "Element.SetMediaSource")
}
func (e *Element) SetMediaControls(value bool) error {
	return statusError(C.kui_media_set_controls(e.raw(), boolInt(value)), "Element.SetMediaControls")
}
func (e *Element) SetCurrentTime(value float64) error {
	return statusError(C.kui_media_set_current_time(e.raw(), C.double(value)), "Element.SetCurrentTime")
}
func (e *Element) CurrentTime() (float64, error) {
	var value C.double
	err := statusError(C.kui_media_get_current_time(e.raw(), &value), "Element.CurrentTime")
	return float64(value), err
}
func (e *Element) Duration() (float64, error) {
	var value C.double
	err := statusError(C.kui_media_get_duration(e.raw(), &value), "Element.Duration")
	return float64(value), err
}
func (e *Element) SetVolume(value float32) error {
	return statusError(C.kui_media_set_volume(e.raw(), C.float(value)), "Element.SetVolume")
}
func (e *Element) Volume() (float32, error) {
	var value C.float
	err := statusError(C.kui_media_get_volume(e.raw(), &value), "Element.Volume")
	return float32(value), err
}
func (e *Element) SetMuted(value bool) error {
	return statusError(C.kui_media_set_muted(e.raw(), boolInt(value)), "Element.SetMuted")
}
func (e *Element) Muted() (bool, error) {
	var value C.int
	err := statusError(C.kui_media_get_muted(e.raw(), &value), "Element.Muted")
	return value != 0, err
}
func (e *Element) Paused() (bool, error) {
	var value C.int
	err := statusError(C.kui_media_get_paused(e.raw(), &value), "Element.Paused")
	return value != 0, err
}
func (e *Element) MediaReadyState() (int, error) {
	var value C.int
	err := statusError(C.kui_media_get_ready_state(e.raw(), &value), "Element.MediaReadyState")
	return int(value), err
}
func (e *Element) MediaNetworkState() (int, error) {
	var value C.int
	err := statusError(C.kui_media_get_network_state(e.raw(), &value), "Element.MediaNetworkState")
	return int(value), err
}
func (e *Element) MediaAvailable() (bool, error) {
	var value C.int
	err := statusError(C.kui_media_is_available(e.raw(), &value), "Element.MediaAvailable")
	return value != 0, err
}
func (e *Element) SetLottieSource(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_lottie_set_src(e.raw(), raw, size), "Element.SetLottieSource")
}
func (e *Element) LottiePlay() error {
	return statusError(C.kui_lottie_play(e.raw()), "Element.LottiePlay")
}
func (e *Element) LottiePause() error {
	return statusError(C.kui_lottie_pause(e.raw()), "Element.LottiePause")
}
func (e *Element) LottieComplete() (bool, error) {
	var value C.int
	err := statusError(C.kui_lottie_get_complete(e.raw(), &value), "Element.LottieComplete")
	return value != 0, err
}
func (e *Element) LottieNaturalSize() (Size, error) {
	var width, height C.float
	err := statusError(C.kui_lottie_get_natural_size(e.raw(), &width, &height), "Element.LottieNaturalSize")
	return Size{Width: float32(width), Height: float32(height)}, err
}

// 3.61.8 起的媒体与 Lottie 播放控制（C ABI 第 13 节）。

// MediaLoad 重新加载当前 src（HTMLMediaElement.load()）。
func (e *Element) MediaLoad() error {
	return statusError(C.kui_media_load(e.raw()), "Element.MediaLoad")
}

// SetMediaLoop 设置是否循环播放。
func (e *Element) SetMediaLoop(loop bool) error {
	return statusError(C.kui_media_set_loop(e.raw(), boolInt(loop)), "Element.SetMediaLoop")
}

// SetMediaAutoplay 设置是否自动播放。
func (e *Element) SetMediaAutoplay(autoplay bool) error {
	return statusError(C.kui_media_set_autoplay(e.raw(), boolInt(autoplay)), "Element.SetMediaAutoplay")
}

// MediaPlaybackFlags 返回 loop / autoplay / ended（不循环且播到结尾时 ended 为 true）。
func (e *Element) MediaPlaybackFlags() (loop, autoplay, ended bool, err error) {
	var l, a, n C.int
	err = statusError(C.kui_media_get_playback_flags(e.raw(), &l, &a, &n), "Element.MediaPlaybackFlags")
	return l != 0, a != 0, n != 0, err
}

// LottieLoadState 是 Lottie 资源的加载状态。
type LottieLoadState int

const (
	LottieIdle LottieLoadState = iota
	LottieLoading
	LottieComplete
	LottieError
)

// LottieState 是 Lottie 的播放状态；Position 为 0-1。
type LottieState struct {
	Playing, Looping, Autoplay bool
	Speed                      float32
	Position                   float64
	Load                       LottieLoadState
}

// LottieStop 停止并回到第一帧。
func (e *Element) LottieStop() error {
	return statusError(C.kui_lottie_stop(e.raw()), "Element.LottieStop")
}

// LottieSeek 跳到 0-1 之间的位置（暂停时也生效）。
func (e *Element) LottieSeek(position float64) error {
	return statusError(C.kui_lottie_seek(e.raw(), C.double(position)), "Element.LottieSeek")
}

// SetLottieSpeed 设置播放速度倍率（≥ 0）。
func (e *Element) SetLottieSpeed(speed float32) error {
	return statusError(C.kui_lottie_set_speed(e.raw(), C.float(speed)), "Element.SetLottieSpeed")
}

// SetLottieLoop 设置是否循环。
func (e *Element) SetLottieLoop(loop bool) error {
	return statusError(C.kui_lottie_set_loop(e.raw(), boolInt(loop)), "Element.SetLottieLoop")
}

// LottieState 返回播放状态。
func (e *Element) LottieState() (LottieState, error) {
	var raw C.KuiCLottieState
	err := statusError(C.kui_lottie_get_state(e.raw(), &raw), "Element.LottieState")
	return LottieState{Playing: raw.playing != 0, Looping: raw.looping != 0, Autoplay: raw.autoplay != 0,
		Speed: float32(raw.speed), Position: float64(raw.position), Load: LottieLoadState(raw.load_state)}, err
}

type eventState struct {
	fn       func(*Event)
	name     string
	document *Document
}

//export goKuiEventCallback
func goKuiEventCallback(raw *C.KuiEvent, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*eventState)
	if ok {
		state.fn(&Event{handle: unsafe.Pointer(raw), document: state.document})
	}
}

func (e *Element) AddEventListener(name string, fn func(*Event)) (uint64, error) {
	return e.addEventListener(name, fn, false, "Element.AddEventListener")
}

// AddCaptureEventListener 登记捕获阶段监听：事件从根往下走时执行（在目标自己身上最先执行，早于普通监听）；
// 在里面 StopPropagation 会终止下行。用 RemoveEventListener 摘除。
func (e *Element) AddCaptureEventListener(name string, fn func(*Event)) (uint64, error) {
	return e.addEventListener(name, fn, true, "Element.AddCaptureEventListener")
}

func (e *Element) addEventListener(name string, fn func(*Event), capture bool, context string) (uint64, error) {
	if fn == nil {
		return 0, &Error{Status: ErrInvalidArg, Context: context, Message: "nil callback"}
	}
	rawName, size, release := inputString(name)
	defer release()
	state := &eventState{fn: fn, name: name, document: e.document}
	handle := cgo.NewHandle(state)
	var id C.KuiListenerId
	var status C.KuiStatus
	if capture {
		status = C.kui_element_add_event_listener_ex(e.raw(), rawName, size, C.kui_go_event_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(handle)), 1, &id)
	} else {
		status = C.kui_element_add_event_listener(e.raw(), rawName, size, C.kui_go_event_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(handle)), &id)
	}
	if err := statusError(status, context); err != nil {
		handle.Delete()
		return 0, err
	}
	e.listeners[uint64(id)] = handle
	return uint64(id), nil
}
func (e *Element) RemoveEventListener(name string, id uint64) error {
	rawName, size, release := inputString(name)
	defer release()
	err := statusError(C.kui_element_remove_event_listener(e.raw(), rawName, size, C.KuiListenerId(id)), "Element.RemoveEventListener")
	if handle, ok := e.listeners[id]; ok {
		handle.Delete()
		delete(e.listeners, id)
	}
	return err
}
func (e *Element) RemoveEventListeners(name string) error {
	rawName, size, release := inputString(name)
	defer release()
	err := statusError(C.kui_element_remove_event_listeners(e.raw(), rawName, size), "Element.RemoveEventListeners")
	for id, handle := range e.listeners {
		if handle.Value().(*eventState).name == name {
			handle.Delete()
			delete(e.listeners, id)
		}
	}
	return err
}
func (e *Element) DispatchEvent(name string, bubbles, cancelable bool) (bool, error) {
	rawName, size, release := inputString(name)
	defer release()
	var prevented C.int
	err := statusError(C.kui_element_dispatch_event(e.raw(), rawName, size, boolInt(bubbles), boolInt(cancelable), &prevented), "Element.DispatchEvent")
	return prevented != 0, err
}

type Event struct {
	handle   unsafe.Pointer
	document *Document
}

func (e *Event) raw() *C.KuiEvent {
	if e == nil {
		return nil
	}
	return (*C.KuiEvent)(e.handle)
}
func (e *Event) readString(call func(*C.char, C.size_t, *C.size_t) C.KuiStatus, context string) (string, error) {
	var size C.size_t
	if err := bufferStatus(call(nil, 0, &size), context); err != nil {
		return "", err
	}
	buffer := make([]byte, int(size)+1)
	if err := statusError(call((*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), context); err != nil {
		return "", err
	}
	return string(buffer[:int(size)]), nil
}
func (e *Event) Type() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_event_get_type(e.raw(), buf, capacity, out)
	}, "Event.Type")
}
func (e *Event) TypeTag() (EventTypeTag, error) {
	var value C.KuiCEventTypeTag
	err := statusError(C.kui_event_get_type_tag(e.raw(), &value), "Event.TypeTag")
	return EventTypeTag(value), err
}
func (e *Event) target(call func(*C.KuiEvent, **C.KuiElement) C.KuiStatus, context string) (*Element, error) {
	var raw *C.KuiElement
	status := call(e.raw(), &raw)
	if Status(status) == ErrNotFound {
		return nil, nil
	}
	if err := statusError(status, context); err != nil {
		return nil, err
	}
	return &Element{handle: unsafe.Pointer(raw), document: e.document, listeners: make(map[uint64]cgo.Handle)}, nil
}
func (e *Event) Target() (*Element, error) {
	return e.target(func(raw *C.KuiEvent, out **C.KuiElement) C.KuiStatus {
		return C.kui_event_get_target(raw, out)
	}, "Event.Target")
}
func (e *Event) CurrentTarget() (*Element, error) {
	return e.target(func(raw *C.KuiEvent, out **C.KuiElement) C.KuiStatus {
		return C.kui_event_get_current_target(raw, out)
	}, "Event.CurrentTarget")
}

// RelatedTarget 是鼠标事件（mouseover / mouseout / mouseenter / mouseleave）与焦点事件（focus / blur）的另一端；
// <kui-context-menu> 的 menu-open 里是被右键的元素。没有时返回 nil, nil。
func (e *Event) RelatedTarget() (*Element, error) {
	return e.target(func(raw *C.KuiEvent, out **C.KuiElement) C.KuiStatus {
		return C.kui_event_get_related_target(raw, out)
	}, "Event.RelatedTarget")
}
func (e *Event) PreventDefault() error {
	return statusError(C.kui_event_prevent_default(e.raw()), "Event.PreventDefault")
}
func (e *Event) StopPropagation() error {
	return statusError(C.kui_event_stop_propagation(e.raw()), "Event.StopPropagation")
}

// Deprecated: 用 FlagValues（EventFlags）；本函数保留一个版本。
func (e *Event) Flags() (bool, bool, bool, error) {
	var bubbles, cancelable, prevented C.int
	err := statusError(C.kui_event_get_flags(e.raw(), &bubbles, &cancelable, &prevented), "Event.Flags")
	return bubbles != 0, cancelable != 0, prevented != 0, err
}
func (e *Event) MouseClient() (Point, error) {
	var x, y C.float
	err := statusError(C.kui_event_mouse_get_client_xy(e.raw(), &x, &y), "Event.MouseClient")
	return Point{X: float32(x), Y: float32(y)}, err
}
func (e *Event) MousePage() (Point, error) {
	var x, y C.float
	err := statusError(C.kui_event_mouse_get_page_xy(e.raw(), &x, &y), "Event.MousePage")
	return Point{X: float32(x), Y: float32(y)}, err
}

// WheelDelta 是 "wheel" 事件的滚动量；不是滚轮事件返回 KUI_ERR_NOT_SUPPORTED。
func (e *Event) WheelDelta() (Point, error) {
	var x, y C.float
	err := statusError(C.kui_event_wheel_get_delta(e.raw(), &x, &y), "Event.WheelDelta")
	return Point{X: float32(x), Y: float32(y)}, err
}
func (e *Event) MouseOffset() (Point, error) {
	var x, y C.float
	err := statusError(C.kui_event_mouse_get_offset_xy(e.raw(), &x, &y), "Event.MouseOffset")
	return Point{X: float32(x), Y: float32(y)}, err
}
func (e *Event) MouseButton() (MouseButton, error) {
	var value C.int
	err := statusError(C.kui_event_mouse_get_button(e.raw(), &value), "Event.MouseButton")
	return MouseButton(value), err
}
func (e *Event) Key() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_event_key_get_key(e.raw(), buf, capacity, out)
	}, "Event.Key")
}
func (e *Event) Code() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_event_key_get_code(e.raw(), buf, capacity, out)
	}, "Event.Code")
}

// Deprecated: 用 ModifierKeys（KeyModifiers）；本函数保留一个版本。
func (e *Event) Modifiers() (bool, bool, bool, bool, error) {
	var ctrl, shift, alt, meta C.int
	err := statusError(C.kui_event_key_get_modifiers(e.raw(), &ctrl, &shift, &alt, &meta), "Event.Modifiers")
	return ctrl != 0, shift != 0, alt != 0, meta != 0, err
}
func (e *Event) DragFiles() ([]string, error) {
	var count C.size_t
	if err := statusError(C.kui_event_drag_get_file_count(e.raw(), &count), "Event.DragFiles"); err != nil {
		return nil, err
	}
	result := make([]string, 0, int(count))
	for index := C.size_t(0); index < count; index++ {
		value, err := e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
			return C.kui_event_drag_get_file(e.raw(), index, buf, capacity, out)
		}, "Event.DragFiles")
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

// SetDragData 在 dragstart 里设置拖动数据，format 是 MIME 类型（如 "text/plain"）。
// 不是拖放事件时返回 ErrNotSupported。
func (e *Event) SetDragData(format, data string) error {
	rawFormat, formatLen, releaseFormat := inputString(format)
	defer releaseFormat()
	rawData, dataLen, releaseData := inputString(data)
	defer releaseData()
	return statusError(C.kui_event_drag_set_data(e.raw(), rawFormat, formatLen, rawData, dataLen), "Event.SetDragData")
}

// DragData 读取拖动数据；没设置过该格式时 found 为 false。
func (e *Event) DragData(format string) (value string, found bool, err error) {
	rawFormat, formatLen, releaseFormat := inputString(format)
	defer releaseFormat()
	var size C.size_t
	if C.kui_event_drag_get_data(e.raw(), rawFormat, formatLen, nil, 0, &size) == C.KUI_ERR_NOT_FOUND {
		return "", false, nil
	}
	value, err = e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_event_drag_get_data(e.raw(), rawFormat, formatLen, buf, capacity, out)
	}, "Event.DragData")
	return value, err == nil, err
}

// ClearDragData 清除一种格式的拖动数据。
func (e *Event) ClearDragData(format string) error {
	rawFormat, formatLen, releaseFormat := inputString(format)
	defer releaseFormat()
	return statusError(C.kui_event_drag_clear_data(e.raw(), rawFormat, formatLen), "Event.ClearDragData")
}

// ClearAllDragData 清除全部拖动数据。
func (e *Event) ClearAllDragData() error {
	return statusError(C.kui_event_drag_clear_data(e.raw(), nil, 0), "Event.ClearAllDragData")
}

// DragTypes 是已设置的数据格式，按设置顺序（DataTransfer.types）。
func (e *Event) DragTypes() ([]string, error) {
	var count C.size_t
	if err := statusError(C.kui_event_drag_get_type_count(e.raw(), &count), "Event.DragTypes"); err != nil {
		return nil, err
	}
	result := make([]string, 0, int(count))
	for index := C.size_t(0); index < count; index++ {
		value, err := e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
			return C.kui_event_drag_get_type(e.raw(), index, buf, capacity, out)
		}, "Event.DragTypes")
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}
func (e *Event) SetDropEffect(effect DropEffect) error {
	return statusError(C.kui_event_drag_set_drop_effect(e.raw(), C.KuiCDropEffect(effect)), "Event.SetDropEffect")
}
func (e *Event) DropEffect() (DropEffect, error) {
	var value C.KuiCDropEffect
	err := statusError(C.kui_event_drag_get_drop_effect(e.raw(), &value), "Event.DropEffect")
	return DropEffect(value), err
}
func (e *Event) SetEffectAllowed(effect EffectAllowed) error {
	return statusError(C.kui_event_drag_set_effect_allowed(e.raw(), C.KuiCEffectAllowed(effect)), "Event.SetEffectAllowed")
}
func (e *Event) EffectAllowed() (EffectAllowed, error) {
	var value C.KuiCEffectAllowed
	err := statusError(C.kui_event_drag_get_effect_allowed(e.raw(), &value), "Event.EffectAllowed")
	return EffectAllowed(value), err
}

type resizeState struct{ fn func(uint32, uint32) }

//export goKuiResizeCallback
func goKuiResizeCallback(width, height C.uint32_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*resizeState)
	if ok {
		state.fn(uint32(width), uint32(height))
	}
}
func (w *Window) OnResize(fn func(uint32, uint32)) error {
	if w.resizeHandle != 0 {
		w.resizeHandle.Delete()
		w.resizeHandle = 0
	}
	if fn == nil {
		return statusError(C.kui_window_on_resize(w.raw(), nil, nil), "Window.OnResize")
	}
	handle := cgo.NewHandle(&resizeState{fn: fn})
	if err := statusError(C.kui_window_on_resize(w.raw(), C.kui_go_resize_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(handle))), "Window.OnResize"); err != nil {
		handle.Delete()
		return err
	}
	w.resizeHandle = handle
	return nil
}

type windowEventState struct {
	fn     func(*Event)
	name   string
	once   bool
	window *Window
	id     uint64
}

//export goKuiWindowEventCallback
func goKuiWindowEventCallback(raw *C.KuiEvent, handle C.uintptr_t) {
	defer containCallbackPanic()
	h := cgo.Handle(handle)
	state, ok := h.Value().(*windowEventState)
	if !ok {
		return
	}
	if state.once {
		delete(state.window.windowListeners, state.id)
		defer h.Delete()
	}
	state.fn(&Event{handle: unsafe.Pointer(raw), document: state.window.document})
}
func (w *Window) AddEventListener(name string, once bool, fn func(*Event)) (uint64, error) {
	if fn == nil {
		return 0, &Error{Status: ErrInvalidArg, Context: "Window.AddEventListener", Message: "nil callback"}
	}
	rawName, size, release := inputString(name)
	defer release()
	state := &windowEventState{fn: fn, name: name, once: once, window: w}
	handle := cgo.NewHandle(state)
	var id C.KuiListenerId
	status := C.kui_window_add_event_listener(w.raw(), rawName, size, C.kui_go_window_event_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(handle)), boolInt(once), &id)
	if err := statusError(status, "Window.AddEventListener"); err != nil {
		handle.Delete()
		return 0, err
	}
	state.id = uint64(id)
	w.windowListeners[state.id] = handle
	return state.id, nil
}
func (w *Window) RemoveEventListener(name string, id uint64) error {
	rawName, size, release := inputString(name)
	defer release()
	err := statusError(C.kui_window_remove_event_listener(w.raw(), rawName, size, C.KuiListenerId(id)), "Window.RemoveEventListener")
	if handle, ok := w.windowListeners[id]; ok {
		handle.Delete()
		delete(w.windowListeners, id)
	}
	return err
}
func (w *Window) RemoveEventListeners(name string) error {
	rawName, size, release := inputString(name)
	defer release()
	err := statusError(C.kui_window_remove_event_listeners(w.raw(), rawName, size), "Window.RemoveEventListeners")
	for id, handle := range w.windowListeners {
		if handle.Value().(*windowEventState).name == name {
			handle.Delete()
			delete(w.windowListeners, id)
		}
	}
	return err
}

type ScriptEngine struct {
	handle      unsafe.Pointer
	errorHandle cgo.Handle
}

func NewScriptEngine() (*ScriptEngine, error) {
	var raw *C.KuiScriptEngine
	if err := statusError(C.kui_script_engine_create(&raw), "NewScriptEngine"); err != nil {
		return nil, err
	}
	return &ScriptEngine{handle: unsafe.Pointer(raw)}, nil
}
func (s *ScriptEngine) raw() *C.KuiScriptEngine {
	if s == nil {
		return nil
	}
	return (*C.KuiScriptEngine)(s.handle)
}
func (s *ScriptEngine) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	if s.errorHandle != 0 {
		s.errorHandle.Delete()
		s.errorHandle = 0
	}
	err := statusError(C.kui_script_engine_release(s.raw()), "ScriptEngine.Close")
	s.handle = nil
	return err
}
func (s *ScriptEngine) BindDocument(document *Document) error {
	return statusError(C.kui_script_engine_bind_document(s.raw(), document.raw()), "ScriptEngine.BindDocument")
}
func (s *ScriptEngine) UnbindDocument() error {
	return statusError(C.kui_script_engine_unbind_document(s.raw()), "ScriptEngine.UnbindDocument")
}
func (s *ScriptEngine) Eval(code, source string) (ScriptResult, error) {
	rawCode, codeLen, releaseCode := inputString(code)
	defer releaseCode()
	rawSource, sourceLen, releaseSource := inputString(source)
	defer releaseSource()
	var success C.int
	var out C.KuiCScriptValue
	status := C.kui_script_engine_eval_value(s.raw(), rawCode, codeLen, rawSource, sourceLen, &success, &out)
	if err := statusError(status, "ScriptEngine.Eval"); err != nil {
		return ScriptResult{}, err
	}
	return s.finishRun(success, &out)
}
func (s *ScriptEngine) ExecuteScripts() error {
	return statusError(C.kui_script_engine_execute_scripts(s.raw()), "ScriptEngine.ExecuteScripts")
}

type scriptErrorState struct{ fn func(string, string, int) }

//export goKuiScriptErrorCallback
func goKuiScriptErrorCallback(rawError *C.char, errorLen C.size_t, rawSource *C.char, sourceLen C.size_t, line C.int, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*scriptErrorState)
	if ok {
		state.fn(C.GoStringN(rawError, C.int(errorLen)), C.GoStringN(rawSource, C.int(sourceLen)), int(line))
	}
}
func (s *ScriptEngine) SetErrorHandler(fn func(string, string, int)) error {
	if s.errorHandle != 0 {
		s.errorHandle.Delete()
		s.errorHandle = 0
	}
	if fn == nil {
		return statusError(C.kui_script_engine_set_error_handler(s.raw(), nil, nil), "ScriptEngine.SetErrorHandler")
	}
	handle := cgo.NewHandle(&scriptErrorState{fn: fn})
	if err := statusError(C.kui_script_engine_set_error_handler(s.raw(), C.kui_go_script_error_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(handle))), "ScriptEngine.SetErrorHandler"); err != nil {
		handle.Delete()
		return err
	}
	s.errorHandle = handle
	return nil
}

type HTTPClient struct {
	handle  unsafe.Pointer
	pending sync.Map
}

func NewHTTPClient() (*HTTPClient, error) {
	var raw *C.KuiHttpClient
	if err := statusError(C.kui_http_client_create(&raw), "NewHTTPClient"); err != nil {
		return nil, err
	}
	return &HTTPClient{handle: unsafe.Pointer(raw)}, nil
}
func (h *HTTPClient) raw() *C.KuiHttpClient {
	if h == nil {
		return nil
	}
	return (*C.KuiHttpClient)(h.handle)
}
func (h *HTTPClient) Close() error {
	if h == nil || h.handle == nil {
		return nil
	}
	_ = h.CancelAll()
	err := statusError(C.kui_http_client_release(h.raw()), "HTTPClient.Close")
	h.handle = nil
	return err
}
func (h *HTTPClient) SetUserAgent(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_http_client_set_user_agent(h.raw(), raw, size), "HTTPClient.SetUserAgent")
}
func takeHTTPResponse(raw *C.KuiHttpResponse) (HTTPResponse, error) {
	if raw == nil {
		return HTTPResponse{}, &Error{Status: ErrInternal, Context: "HTTPResponse", Message: "nil response"}
	}
	defer C.kui_http_response_release(raw)
	var statusCode C.int
	if err := statusError(C.kui_http_response_get_status(raw, &statusCode), "HTTPResponse.Status"); err != nil {
		return HTTPResponse{}, err
	}
	readBytes := func(call func(*C.uint8_t, C.size_t, *C.size_t) C.KuiStatus) ([]byte, error) {
		var size C.size_t
		if err := bufferStatus(call(nil, 0, &size), "HTTPResponse.Body"); err != nil {
			return nil, err
		}
		if size == 0 {
			return []byte{}, nil
		}
		// The C ABI body getter follows the shared output-buffer contract:
		// capacity includes a trailing NUL even though out_len is byte data.
		buffer := make([]byte, int(size)+1)
		if err := statusError(call((*C.uint8_t)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "HTTPResponse.Body"); err != nil {
			return nil, err
		}
		return buffer[:int(size)], nil
	}
	body, err := readBytes(func(buf *C.uint8_t, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_http_response_get_body(raw, buf, capacity, out)
	})
	if err != nil {
		return HTTPResponse{}, err
	}
	readString := func(call func(*C.char, C.size_t, *C.size_t) C.KuiStatus, context string) (string, error) {
		var size C.size_t
		if err := bufferStatus(call(nil, 0, &size), context); err != nil {
			return "", err
		}
		buffer := make([]byte, int(size)+1)
		if err := statusError(call((*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), context); err != nil {
			return "", err
		}
		return string(buffer[:int(size)]), nil
	}
	contentType, err := readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_http_response_get_content_type(raw, buf, capacity, out)
	}, "HTTPResponse.ContentType")
	if err != nil {
		return HTTPResponse{}, err
	}
	errorText, err := readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_http_response_get_error(raw, buf, capacity, out)
	}, "HTTPResponse.Error")
	if err != nil {
		return HTTPResponse{}, err
	}
	return HTTPResponse{Status: int(statusCode), Body: body, ContentType: contentType, Error: errorText}, nil
}
func (h *HTTPClient) Fetch(url string, timeoutSeconds float64) (HTTPResponse, error) {
	rawURL, size, release := inputString(url)
	defer release()
	var raw *C.KuiHttpResponse
	if err := statusError(C.kui_http_client_fetch_sync(h.raw(), rawURL, size, C.double(timeoutSeconds), &raw), "HTTPClient.Fetch"); err != nil {
		return HTTPResponse{}, err
	}
	return takeHTTPResponse(raw)
}

type httpCallbackState struct {
	fn     func(HTTPResponse, error)
	client *HTTPClient
	handle cgo.Handle
}

//export goKuiHttpResponseCallback
func goKuiHttpResponseCallback(raw *C.KuiHttpResponse, handle C.uintptr_t) {
	defer containCallbackPanic()
	h := cgo.Handle(handle)
	state, ok := h.Value().(*httpCallbackState)
	if !ok {
		return
	}
	state.client.pending.Delete(uintptr(handle))
	defer h.Delete()
	response, err := takeHTTPResponse(raw)
	state.fn(response, err)
}

// FetchAsync 发起异步 GET。fn 在 HTTP 工作线程上执行，不在 UI 线程：可以处理响应数据，但要操作 DOM / 窗口
// 必须经 PostTask 交回 UI 线程，或直接改用 FetchAsyncOnMainThread。
func (h *HTTPClient) FetchAsync(url string, timeoutSeconds float64, fn func(HTTPResponse, error)) error {
	return h.fetchAsync(url, timeoutSeconds, fn, false, "HTTPClient.FetchAsync")
}

// FetchAsyncOnMainThread 与 FetchAsync 相同，但 fn 在 UI 线程的事件循环里执行，可以直接操作 DOM / 窗口。
// Init 之前或 Shutdown 开始之后调用返回错误（ErrNotSupported）；结果还在排队时遇到 Shutdown，fn 在 Shutdown
// 内以错误调用一次；请求在 Shutdown 之后才完成时 fn 不再调用。
func (h *HTTPClient) FetchAsyncOnMainThread(url string, timeoutSeconds float64, fn func(HTTPResponse, error)) error {
	return h.fetchAsync(url, timeoutSeconds, fn, true, "HTTPClient.FetchAsyncOnMainThread")
}

func (h *HTTPClient) fetchAsync(url string, timeoutSeconds float64, fn func(HTTPResponse, error), onMainThread bool, context string) error {
	if fn == nil {
		return &Error{Status: ErrInvalidArg, Context: context, Message: "nil callback"}
	}
	rawURL, size, release := inputString(url)
	defer release()
	state := &httpCallbackState{fn: fn, client: h}
	handle := cgo.NewHandle(state)
	state.handle = handle
	h.pending.Store(uintptr(handle), handle)
	var status C.KuiStatus
	if onMainThread {
		status = C.kui_http_client_fetch_async_on_main_thread(h.raw(), rawURL, size, C.double(timeoutSeconds), C.kui_go_http_response_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(handle)))
	} else {
		status = C.kui_http_client_fetch_async(h.raw(), rawURL, size, C.double(timeoutSeconds), C.kui_go_http_response_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(handle)))
	}
	if err := statusError(status, context); err != nil {
		h.pending.Delete(uintptr(handle))
		handle.Delete()
		return err
	}
	return nil
}
func (h *HTTPClient) CancelAll() error {
	// The native contract completes every accepted asynchronous request by
	// invoking its callback, including cancelled requests. The callback owns and
	// deletes the cgo.Handle; deleting it here would make the later trampoline
	// dereference an invalid handle.
	return statusError(C.kui_http_client_cancel_all(h.raw()), "HTTPClient.CancelAll")
}

func (d *Document) InvokeExtension(element *Element, method string, args []ExtensionValue, maxResults int) ([]ExtensionValue, error) {
	rawMethod, methodLen, releaseMethod := inputString(method)
	defer releaseMethod()
	rawArgs := make([]C.KuiCExtensionValue, len(args))
	strings := make([]*C.char, 0, len(args))
	defer func() {
		for _, value := range strings {
			C.free(unsafe.Pointer(value))
		}
	}()
	for index, value := range args {
		rawArgs[index]._type = C.KuiCExtensionValueType(value.Type)
		rawArgs[index].int_value = C.int64_t(value.Int)
		rawArgs[index].float_value = C.double(value.Float)
		rawArgs[index].bool_value = boolInt(value.Bool)
		if value.Type == ExtensionString {
			raw := C.CString(value.String)
			strings = append(strings, raw)
			rawArgs[index].string_value = raw
			rawArgs[index].string_len = C.size_t(len(value.String))
		}
	}
	var argsPtr *C.KuiCExtensionValue
	if len(rawArgs) != 0 {
		argsPtr = &rawArgs[0]
	}
	var result *C.KuiExtensionResult
	if err := statusError(C.kui_document_extension_invoke(d.raw(), element.raw(), rawMethod, methodLen, argsPtr, C.size_t(len(rawArgs)), C.size_t(maxResults), &result), "Document.InvokeExtension"); err != nil {
		return nil, err
	}
	defer C.kui_extension_result_release(result)
	var extensionStatus C.int
	if err := statusError(C.kui_extension_result_get_status(result, &extensionStatus), "ExtensionResult.Status"); err != nil {
		return nil, err
	}
	if extensionStatus != 0 {
		errorText, readErr := func() (string, error) {
			var size C.size_t
			if err := bufferStatus(C.kui_extension_result_get_error(result, nil, 0, &size), "ExtensionResult.Error"); err != nil {
				return "", err
			}
			buffer := make([]byte, int(size)+1)
			if err := statusError(C.kui_extension_result_get_error(result, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "ExtensionResult.Error"); err != nil {
				return "", err
			}
			return string(buffer[:int(size)]), nil
		}()
		if readErr != nil {
			return nil, readErr
		}
		return nil, &Error{Status: ErrInternal, Context: "Document.InvokeExtension", Message: errorText}
	}
	var count C.size_t
	if err := statusError(C.kui_extension_result_get_count(result, &count), "ExtensionResult.Count"); err != nil {
		return nil, err
	}
	values := make([]ExtensionValue, 0, int(count))
	for index := C.size_t(0); index < count; index++ {
		var value C.KuiCExtensionValue
		if err := statusError(C.kui_extension_result_get_value(result, index, &value), "ExtensionResult.Value"); err != nil {
			return nil, err
		}
		item := ExtensionValue{Type: ExtensionValueType(value._type), Int: int64(value.int_value), Float: float64(value.float_value), Bool: value.bool_value != 0}
		if item.Type == ExtensionString {
			var size C.size_t
			if err := bufferStatus(C.kui_extension_result_get_string(result, index, nil, 0, &size), "ExtensionResult.String"); err != nil {
				return nil, err
			}
			buffer := make([]byte, int(size)+1)
			if err := statusError(C.kui_extension_result_get_string(result, index, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "ExtensionResult.String"); err != nil {
				return nil, err
			}
			item.String = string(buffer[:int(size)])
		}
		values = append(values, item)
	}
	return values, nil
}

type CustomElementDefinition struct {
	ObservedAttributes []string
	Connected          func(*Element)
	Disconnected       func(*Element)
	AttributeChanged   func(*Element, string, string, string)

	// 3.61.8（C ABI 第 17 节）：自绘、tick 与输入，见 paint_windows.go。
	// Tick 的 deltaSeconds 是距上次 tick 的秒数：元素约每 250 ms tick 一次，tick 里调用了
	// RequestRepaint / InvalidateRect 时下一帧也会 tick。输入回调返回 true 表示已处理。
	Paint     func(element *Element, paint *PaintContext)
	Tick      func(element *Element, deltaSeconds float64)
	NeedsTick func(element *Element) bool
	Key       func(element *Element, event CustomKeyEvent) bool
	Mouse     func(element *Element, event CustomMouseEvent) bool
	Wheel     func(element *Element, event CustomWheelEvent) bool
	IME       func(element *Element, event CustomIMEEvent) bool
}
type customElementState struct {
	document   *Document
	definition CustomElementDefinition
}

//export goKuiCustomConnectedCallback
func goKuiCustomConnectedCallback(raw *C.KuiElement, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*customElementState)
	if !ok || state.definition.Connected == nil {
		return
	}
	state.definition.Connected(&Element{handle: unsafe.Pointer(raw), document: state.document, borrowed: true})
}

//export goKuiCustomDisconnectedCallback
func goKuiCustomDisconnectedCallback(raw *C.KuiElement, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*customElementState)
	if !ok || state.definition.Disconnected == nil {
		return
	}
	state.definition.Disconnected(&Element{handle: unsafe.Pointer(raw), document: state.document, borrowed: true})
}

//export goKuiCustomAttributeCallback
func goKuiCustomAttributeCallback(raw *C.KuiElement, name *C.char, nameLen C.size_t, oldValue *C.char, oldLen C.size_t, newValue *C.char, newLen C.size_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*customElementState)
	if !ok || state.definition.AttributeChanged == nil {
		return
	}
	element := &Element{handle: unsafe.Pointer(raw), document: state.document, borrowed: true}
	state.definition.AttributeChanged(element, C.GoStringN(name, C.int(nameLen)), C.GoStringN(oldValue, C.int(oldLen)), C.GoStringN(newValue, C.int(newLen)))
}
func (d *Document) DefineCustomElement(tag string, definition CustomElementDefinition) error {
	rawTag, tagLen, releaseTag := inputString(tag)
	defer releaseTag()
	state := &customElementState{document: d, definition: definition}
	handle := cgo.NewHandle(state)
	rawObserved := make([]*C.char, len(definition.ObservedAttributes))
	for index, value := range definition.ObservedAttributes {
		rawObserved[index] = C.CString(value)
		defer C.free(unsafe.Pointer(rawObserved[index]))
	}
	var observed **C.char
	if len(rawObserved) != 0 {
		observed = (**C.char)(unsafe.Pointer(&rawObserved[0]))
	}
	var rawDefinition C.KuiCCustomElementDefinition
	rawDefinition.callbacks = C.KuiCCustomElementCallbacks{observed_attributes: observed, observed_attribute_count: C.size_t(len(rawObserved)), user_data: C.kui_go_handle_pointer(C.uintptr_t(handle))}
	if definition.Connected != nil {
		rawDefinition.callbacks.connected = C.kui_go_custom_connected_trampoline()
	}
	if definition.Disconnected != nil {
		rawDefinition.callbacks.disconnected = C.kui_go_custom_disconnected_trampoline()
	}
	if definition.AttributeChanged != nil {
		rawDefinition.callbacks.attribute_changed = C.kui_go_custom_attribute_trampoline()
	}
	customHooks(definition, &rawDefinition)
	// 回调句柄随文档关闭释放（customHandles），不交给 release_fn。
	if err := statusError(C.kui_document_define_custom_element_ex(d.raw(), rawTag, tagLen, &rawDefinition), "Document.DefineCustomElement"); err != nil {
		handle.Delete()
		return err
	}
	d.state.customHandles = append(d.state.customHandles, handle)
	return nil
}
func (d *Document) CustomElementDefined(tag string) (bool, error) {
	raw, size, release := inputString(tag)
	defer release()
	var value C.int
	err := statusError(C.kui_document_custom_element_is_defined(d.raw(), raw, size, &value), "Document.CustomElementDefined")
	return value != 0, err
}
