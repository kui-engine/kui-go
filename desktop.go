//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

/*
#cgo CFLAGS: -std=c11 -I${SRCDIR}
#include <stdlib.h>
#include "loader.h"
#include "callbacks.h"
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// 3.61.8 起的桌面集成（C ABI 第 10 节）：原生菜单栏、托盘、窗口级快捷键、Toast、截图。
// 回调都在 UI 线程执行；它们的句柄在替换、移除或 Window.Close 时释放。

// MenuItem 描述一个菜单项。ID 由调用方指定，被点击时传给激活回调；Children 非空即为子菜单。
type MenuItem struct {
	Text      string // "&Open"：& 标记助记键
	Shortcut  string // 只用于显示，例如 "Ctrl+O"；真正的快捷键用 BindShortcut
	ID        uint32
	Separator bool
	Disabled  bool
	Checkable bool
	Checked   bool
	Children  []MenuItem
}

// Menu 是菜单栏里的一个顶层菜单。
type Menu struct {
	Text  string // 例如 "&File"
	Items []MenuItem
}

// TrayEvent 是托盘图标上的用户操作。
type TrayEvent int

const (
	TrayClick TrayEvent = iota
	TrayDoubleClick
	TrayRightClick
	TrayBalloonClick
)

type desktopState struct {
	menu      cgo.Handle
	trayMenu  cgo.Handle
	tray      cgo.Handle
	shortcuts map[uint32]cgo.Handle
}

func (s *desktopState) release() {
	if s == nil {
		return
	}
	for _, h := range []cgo.Handle{s.menu, s.trayMenu, s.tray} {
		if h != 0 {
			h.Delete()
		}
	}
	for id, h := range s.shortcuts {
		h.Delete()
		delete(s.shortcuts, id)
	}
	s.menu, s.trayMenu, s.tray = 0, 0, 0
}

func (w *Window) desktopHandles() *desktopState {
	if w.desktop == nil {
		w.desktop = &desktopState{shortcuts: map[uint32]cgo.Handle{}}
	}
	return w.desktop
}

// replaceHandle 把 slot 换成 next，并释放旧句柄。
func replaceHandle(slot *cgo.Handle, next cgo.Handle) {
	if *slot != 0 {
		slot.Delete()
	}
	*slot = next
}

//export goKuiMenuActivateCallback
func goKuiMenuActivateCallback(id C.uint32_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	if fn, ok := cgo.Handle(handle).Value().(func(uint32)); ok {
		fn(uint32(id))
	}
}

//export goKuiTrayEventCallback
func goKuiTrayEventCallback(event C.int, handle C.uintptr_t) {
	defer containCallbackPanic()
	if fn, ok := cgo.Handle(handle).Value().(func(TrayEvent)); ok {
		fn(TrayEvent(event))
	}
}

// menuArena 在 C 堆上搭菜单结构（C 层会复制，调用返回后即可释放）。
type menuArena struct{ blocks []unsafe.Pointer }

func (a *menuArena) alloc(size uintptr) unsafe.Pointer {
	p := C.calloc(1, C.size_t(size))
	a.blocks = append(a.blocks, p)
	return p
}

func (a *menuArena) str(value string) (*C.char, C.size_t) {
	raw := C.CString(value)
	a.blocks = append(a.blocks, unsafe.Pointer(raw))
	return raw, C.size_t(len(value))
}

func (a *menuArena) free() {
	for _, p := range a.blocks {
		C.free(p)
	}
	a.blocks = nil
}

func (a *menuArena) items(items []MenuItem) *C.KuiCMenuItem {
	if len(items) == 0 {
		return nil
	}
	size := unsafe.Sizeof(C.KuiCMenuItem{})
	base := a.alloc(size * uintptr(len(items)))
	out := unsafe.Slice((*C.KuiCMenuItem)(base), len(items))
	for i, item := range items {
		out[i].separator = boolInt(item.Separator)
		out[i].id = C.uint32_t(item.ID)
		out[i].disabled = boolInt(item.Disabled)
		out[i].checkable = boolInt(item.Checkable)
		out[i].checked = boolInt(item.Checked)
		if !item.Separator {
			out[i].text, out[i].text_len = a.str(item.Text)
			if item.Shortcut != "" {
				out[i].shortcut_label, out[i].shortcut_label_len = a.str(item.Shortcut)
			}
		}
		if len(item.Children) > 0 {
			out[i].children = a.items(item.Children)
			out[i].child_count = C.size_t(len(item.Children))
		}
	}
	return (*C.KuiCMenuItem)(base)
}

// SetMenuBar 替换窗口的原生菜单栏；onActivate 以菜单项 ID 在 UI 线程被调用。
func (w *Window) SetMenuBar(menus []Menu, onActivate func(id uint32)) error {
	var arena menuArena
	defer arena.free()
	var raw *C.KuiCMenu
	if len(menus) > 0 {
		base := arena.alloc(unsafe.Sizeof(C.KuiCMenu{}) * uintptr(len(menus)))
		out := unsafe.Slice((*C.KuiCMenu)(base), len(menus))
		for i, menu := range menus {
			out[i].text, out[i].text_len = arena.str(menu.Text)
			out[i].items = arena.items(menu.Items)
			out[i].item_count = C.size_t(len(menu.Items))
		}
		raw = (*C.KuiCMenu)(base)
	}
	var fn C.KuiMenuActivateFn
	var handle cgo.Handle
	if onActivate != nil {
		handle = cgo.NewHandle(onActivate)
		fn = C.kui_go_menu_activate_trampoline()
	}
	if err := statusError(C.kui_window_set_menu_bar(w.raw(), raw, C.size_t(len(menus)), fn, C.kui_go_handle_pointer(C.uintptr_t(handle))), "Window.SetMenuBar"); err != nil {
		if handle != 0 {
			handle.Delete()
		}
		return err
	}
	replaceHandle(&w.desktopHandles().menu, handle)
	return nil
}

func (w *Window) ClearMenuBar() error {
	err := statusError(C.kui_window_clear_menu_bar(w.raw()), "Window.ClearMenuBar")
	replaceHandle(&w.desktopHandles().menu, 0)
	return err
}

// SetMenuItemChecked / SetMenuItemEnabled 改菜单项状态并重建菜单；ID 不存在时返回 ErrNotFound。
func (w *Window) SetMenuItemChecked(id uint32, checked bool) error {
	return statusError(C.kui_window_menu_set_checked(w.raw(), C.uint32_t(id), boolInt(checked)), "Window.SetMenuItemChecked")
}
func (w *Window) SetMenuItemEnabled(id uint32, enabled bool) error {
	return statusError(C.kui_window_menu_set_enabled(w.raw(), C.uint32_t(id), boolInt(enabled)), "Window.SetMenuItemEnabled")
}

// CreateTray 为窗口创建托盘图标；iconPath 为空用应用图标。平台不支持时返回 ErrNotSupported。
func (w *Window) CreateTray(iconPath, tooltip string, visible bool) error {
	rawIcon, iconLen, releaseIcon := inputString(iconPath)
	defer releaseIcon()
	rawTip, tipLen, releaseTip := inputString(tooltip)
	defer releaseTip()
	return statusError(C.kui_window_tray_create(w.raw(), rawIcon, iconLen, rawTip, tipLen, boolInt(visible)), "Window.CreateTray")
}
func (w *Window) DestroyTray() error {
	err := statusError(C.kui_window_tray_destroy(w.raw()), "Window.DestroyTray")
	replaceHandle(&w.desktopHandles().trayMenu, 0)
	return err
}
func (w *Window) TrayActive() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_tray_is_active(w.raw(), &value), "Window.TrayActive")
	return value != 0, err
}
func (w *Window) SetTrayVisible(visible bool) error {
	return statusError(C.kui_window_tray_set_visible(w.raw(), boolInt(visible)), "Window.SetTrayVisible")
}
func (w *Window) TrayVisible() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_tray_is_visible(w.raw(), &value), "Window.TrayVisible")
	return value != 0, err
}
func (w *Window) SetTrayTooltip(tooltip string) error {
	raw, size, release := inputString(tooltip)
	defer release()
	return statusError(C.kui_window_tray_set_tooltip(w.raw(), raw, size), "Window.SetTrayTooltip")
}
func (w *Window) SetTrayIcon(iconPath string) error {
	raw, size, release := inputString(iconPath)
	defer release()
	return statusError(C.kui_window_tray_set_icon(w.raw(), raw, size), "Window.SetTrayIcon")
}

// SetTrayContextMenu 设置托盘右键菜单；onActivate 以菜单项 ID 被调用。
func (w *Window) SetTrayContextMenu(items []MenuItem, onActivate func(id uint32)) error {
	var arena menuArena
	defer arena.free()
	raw := arena.items(items)
	var fn C.KuiMenuActivateFn
	var handle cgo.Handle
	if onActivate != nil {
		handle = cgo.NewHandle(onActivate)
		fn = C.kui_go_menu_activate_trampoline()
	}
	if err := statusError(C.kui_window_tray_set_context_menu(w.raw(), raw, C.size_t(len(items)), fn, C.kui_go_handle_pointer(C.uintptr_t(handle))), "Window.SetTrayContextMenu"); err != nil {
		if handle != 0 {
			handle.Delete()
		}
		return err
	}
	replaceHandle(&w.desktopHandles().trayMenu, handle)
	return nil
}
func (w *Window) PopupTrayContextMenu() error {
	return statusError(C.kui_window_tray_popup_context_menu(w.raw()), "Window.PopupTrayContextMenu")
}

// ShowTrayBalloon 弹出托盘气泡；icon 0 无、1 信息、2 警告、3 错误。
func (w *Window) ShowTrayBalloon(title, message string, icon, durationMS int) error {
	rawTitle, titleLen, releaseTitle := inputString(title)
	defer releaseTitle()
	rawMessage, messageLen, releaseMessage := inputString(message)
	defer releaseMessage()
	return statusError(C.kui_window_tray_show_balloon(w.raw(), rawTitle, titleLen, rawMessage, messageLen, C.int(icon), C.int(durationMS)), "Window.ShowTrayBalloon")
}

// SetTrayEventHandler 设置托盘点击 / 双击 / 右键 / 气泡点击的回调；nil 清除。
func (w *Window) SetTrayEventHandler(fn func(TrayEvent)) error {
	var cfn C.KuiTrayEventFn
	var handle cgo.Handle
	if fn != nil {
		handle = cgo.NewHandle(fn)
		cfn = C.kui_go_tray_event_trampoline()
	}
	if err := statusError(C.kui_window_tray_set_event_callback(w.raw(), cfn, C.kui_go_handle_pointer(C.uintptr_t(handle))), "Window.SetTrayEventHandler"); err != nil {
		if handle != 0 {
			handle.Delete()
		}
		return err
	}
	replaceHandle(&w.desktopHandles().tray, handle)
	return nil
}

// SimulateTrayEvent 像用户操作一样触发托盘事件（自动化与测试）。
func (w *Window) SimulateTrayEvent(event TrayEvent) error {
	return statusError(C.kui_window_tray_simulate_event(w.raw(), C.KuiCTrayEvent(event)), "Window.SimulateTrayEvent")
}

// BindShortcut 绑定窗口级快捷键（"Ctrl+S"、"Ctrl+Shift+F5"……）。同一组合再绑会替换回调并返回同一 ID。
func (w *Window) BindShortcut(shortcut string, fn func()) (uint32, error) {
	if fn == nil {
		return 0, &Error{Status: ErrInvalidArg, Context: "Window.BindShortcut", Message: "nil callback"}
	}
	raw, size, release := inputString(shortcut)
	defer release()
	handle := cgo.NewHandle(&timerState{fn: fn})
	var id C.uint32_t
	if err := statusError(C.kui_window_bind_shortcut(w.raw(), raw, size, C.kui_go_task_trampoline(), C.kui_go_handle_pointer(C.uintptr_t(handle)), &id), "Window.BindShortcut"); err != nil {
		handle.Delete()
		return 0, err
	}
	state := w.desktopHandles()
	if old, ok := state.shortcuts[uint32(id)]; ok {
		old.Delete()
	}
	state.shortcuts[uint32(id)] = handle
	return uint32(id), nil
}
func (w *Window) UnbindShortcut(id uint32) error {
	err := statusError(C.kui_window_unbind_shortcut(w.raw(), C.uint32_t(id)), "Window.UnbindShortcut")
	state := w.desktopHandles()
	if h, ok := state.shortcuts[id]; ok {
		h.Delete()
		delete(state.shortcuts, id)
	}
	return err
}
func (w *Window) UnbindAllShortcuts() error {
	err := statusError(C.kui_window_unbind_all_shortcuts(w.raw()), "Window.UnbindAllShortcuts")
	state := w.desktopHandles()
	for id, h := range state.shortcuts {
		h.Delete()
		delete(state.shortcuts, id)
	}
	return err
}
func (w *Window) SetShortcutsEnabled(enabled bool) error {
	return statusError(C.kui_window_set_shortcuts_enabled(w.raw(), boolInt(enabled)), "Window.SetShortcutsEnabled")
}

// SaveScreenshot 把窗口当前内容写成 PNG；写不出时返回 false。
func (w *Window) SaveScreenshot(path string) (bool, error) {
	raw, size, release := inputString(path)
	defer release()
	var saved C.int
	err := statusError(C.kui_window_save_screenshot(w.raw(), raw, size, &saved), "Window.SaveScreenshot")
	return saved != 0, err
}

// ToastLevel 是 Toast 的样式级别。
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastOptions 对应原生 ToastConfig。DurationMS 为 0 时一直显示到被关闭；Position 为空即 "bottom-right"。
type ToastOptions struct {
	Message    string
	Title      string
	Level      ToastLevel
	DurationMS int
	Position   string
	Closable   bool
	Icon       string
}

func (d *Document) ShowToast(options ToastOptions) (uint32, error) {
	rawMessage, messageLen, releaseMessage := inputString(options.Message)
	defer releaseMessage()
	rawTitle, titleLen, releaseTitle := inputString(options.Title)
	defer releaseTitle()
	raw := C.KuiCToastOptions{
		message: rawMessage, message_len: messageLen,
		title: rawTitle, title_len: titleLen,
		level: C.KuiCToastLevel(options.Level), duration_ms: C.int(options.DurationMS),
		closable: boolInt(options.Closable),
	}
	if options.Position != "" {
		rawPosition, positionLen, releasePosition := inputString(options.Position)
		defer releasePosition()
		raw.position, raw.position_len = rawPosition, positionLen
	}
	if options.Icon != "" {
		rawIcon, iconLen, releaseIcon := inputString(options.Icon)
		defer releaseIcon()
		raw.icon, raw.icon_len = rawIcon, iconLen
	}
	var id C.uint32_t
	err := statusError(C.kui_document_show_toast(d.raw(), &raw, &id), "Document.ShowToast")
	return uint32(id), err
}
func (d *Document) DismissToast(id uint32) error {
	return statusError(C.kui_document_dismiss_toast(d.raw(), C.uint32_t(id)), "Document.DismissToast")
}
func (d *Document) DismissAllToasts() error {
	return statusError(C.kui_document_dismiss_all_toasts(d.raw()), "Document.DismissAllToasts")
}
func (d *Document) ToastCount() (int, error) {
	var value C.int
	err := statusError(C.kui_document_get_toast_count(d.raw(), &value), "Document.ToastCount")
	return int(value), err
}
