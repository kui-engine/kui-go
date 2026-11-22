//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

/*
#cgo CFLAGS: -std=c11 -I${SRCDIR}
#include <stdlib.h>
#include "loader.h"
*/
import "C"

import "unsafe"

// 3.61.8 起的常用窗口、DOM 与平台操作（C ABI 第 9 节）。窗口关闭 / 聚焦通知请用
// Window.AddEventListener 注册 "close"（可 PreventDefault 阻止关闭）、"focus"、"blur"。

// Cursor 对应原生 CursorType（CSS cursor 关键字）。
type Cursor int

const (
	CursorAuto Cursor = iota
	CursorDefault
	CursorPointer
	CursorCrosshair
	CursorText
	CursorMove
	CursorNotAllowed
	CursorGrab
	CursorGrabbing
	CursorWait
	CursorProgress
	CursorHelp
	CursorContextMenu
	CursorNResize
	CursorEResize
	CursorSResize
	CursorWResize
	CursorNEResize
	CursorNWResize
	CursorSEResize
	CursorSWResize
	CursorEWResize
	CursorNSResize
	CursorNESWResize
	CursorNWSEResize
	CursorColResize
	CursorRowResize
)

// FrameRadii 是窗口四个角的圆角半径（CSS px）。
type FrameRadii struct {
	TopLeft, TopRight, BottomRight, BottomLeft float32
}

// FrameState 是无框窗口的三个开关。
type FrameState struct {
	Frameless, Shadow, Sizeable bool
}

func (w *Window) Focus() error { return statusError(C.kui_window_focus(w.raw()), "Window.Focus") }
func (w *Window) HasFocus() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_has_focus(w.raw(), &value), "Window.HasFocus")
	return value != 0, err
}

// SetCursor 立即切换光标；指针下一次进入别的内容时，CSS cursor 会重新生效。
func (w *Window) SetCursor(cursor Cursor) error {
	return statusError(C.kui_window_set_cursor(w.raw(), C.KuiCCursor(cursor)), "Window.SetCursor")
}
func (w *Window) SetCursorVisible(visible bool) error {
	return statusError(C.kui_window_set_cursor_visible(w.raw(), boolInt(visible)), "Window.SetCursorVisible")
}
func (w *Window) SetFullscreen(fullscreen bool) error {
	return statusError(C.kui_window_set_fullscreen(w.raw(), boolInt(fullscreen)), "Window.SetFullscreen")
}
func (w *Window) IsFullscreen() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_is_fullscreen(w.raw(), &value), "Window.IsFullscreen")
	return value != 0, err
}
func (w *Window) SetFrameless(frameless bool) error {
	return statusError(C.kui_window_set_frameless(w.raw(), boolInt(frameless)), "Window.SetFrameless")
}
func (w *Window) SetFramelessShadow(shadow bool) error {
	return statusError(C.kui_window_set_frameless_shadow(w.raw(), boolInt(shadow)), "Window.SetFramelessShadow")
}
func (w *Window) FrameState() (FrameState, error) {
	var frameless, shadow, sizeable C.int
	err := statusError(C.kui_window_get_frame_state(w.raw(), &frameless, &shadow, &sizeable), "Window.FrameState")
	return FrameState{Frameless: frameless != 0, Shadow: shadow != 0, Sizeable: sizeable != 0}, err
}
func (w *Window) SetFrameRadii(radii FrameRadii) error {
	return statusError(C.kui_window_set_frame_radii(w.raw(), C.float(radii.TopLeft), C.float(radii.TopRight),
		C.float(radii.BottomRight), C.float(radii.BottomLeft)), "Window.SetFrameRadii")
}

// FrameRadii 返回当前圆角；没有设置圆角时 ok 为 false。
func (w *Window) FrameRadii() (radii FrameRadii, ok bool, err error) {
	var tl, tr, br, bl C.float
	status := C.kui_window_get_frame_radii(w.raw(), &tl, &tr, &br, &bl)
	if status == C.KUI_ERR_NOT_FOUND {
		return FrameRadii{}, false, nil
	}
	if err := statusError(status, "Window.FrameRadii"); err != nil {
		return FrameRadii{}, false, err
	}
	return FrameRadii{TopLeft: float32(tl), TopRight: float32(tr), BottomRight: float32(br), BottomLeft: float32(bl)}, true, nil
}
func (w *Window) SetIMECandidatePos(x, y float32) error {
	return statusError(C.kui_window_set_ime_candidate_pos(w.raw(), C.float(x), C.float(y)), "Window.SetIMECandidatePos")
}

// InsertBefore 把 child 插到 ref 前面；ref 为 nil 时追加到末尾。已在树里的 child 会被移动。
func (e *Element) InsertBefore(child, ref *Element) error {
	var rawRef *C.KuiElement
	if ref != nil {
		rawRef = ref.raw()
	}
	return statusError(C.kui_element_insert_before(e.raw(), child.raw(), rawRef), "Element.InsertBefore")
}
func (e *Element) ReplaceChild(newChild, oldChild *Element) error {
	return statusError(C.kui_element_replace_child(e.raw(), newChild.raw(), oldChild.raw()), "Element.ReplaceChild")
}

// CloneNode 返回脱离文档的新元素（DOM cloneNode）；deep 为 true 时连子树一起复制。
// Clone() 则只是同一个元素的另一个句柄。
func (e *Element) CloneNode(deep bool) (*Element, error) {
	var raw *C.KuiElement
	status := C.kui_element_clone(e.raw(), boolInt(deep), &raw)
	return e.document.wrapElement(raw, status, "Element.CloneNode")
}
func (e *Element) Matches(selector string) (bool, error) {
	raw, size, release := inputString(selector)
	defer release()
	var value C.int
	err := statusError(C.kui_element_matches(e.raw(), raw, size, &value), "Element.Matches")
	return value != 0, err
}

// Closest 返回最近的匹配祖先（含自己）；没有时为 nil。
func (e *Element) Closest(selector string) (*Element, error) {
	raw, size, release := inputString(selector)
	defer release()
	var out *C.KuiElement
	status := C.kui_element_closest(e.raw(), raw, size, &out)
	return e.document.wrapElement(out, status, "Element.Closest")
}
func (e *Element) OuterHTML() (string, error) {
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_get_outer_html(e.raw(), buf, capacity, out)
	}, "Element.OuterHTML")
}

// ComputedStyleValue 等于 getComputedStyle(el).getPropertyValue(name)：按 CSS 属性名（短横线写法）
// 读计算后的值，与 Lua 的 getComputedStyle 同口径；逻辑属性按 direction 映射。未知属性名或不在
// 文档里的元素返回空串。
func (e *Element) ComputedStyleValue(property string) (string, error) {
	raw, size, release := inputString(property)
	defer release()
	return e.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_element_get_computed_style_value(e.raw(), raw, size, buf, capacity, out)
	}, "Element.ComputedStyleValue")
}

// InsertAdjacentHTML 的 position 是 "beforebegin"、"afterbegin"、"beforeend" 或 "afterend"。
func (e *Element) InsertAdjacentHTML(position, html string) error {
	rawPosition, positionLen, releasePosition := inputString(position)
	defer releasePosition()
	rawHTML, htmlLen, releaseHTML := inputString(html)
	defer releaseHTML()
	return statusError(C.kui_element_insert_adjacent_html(e.raw(), rawPosition, positionLen, rawHTML, htmlLen), "Element.InsertAdjacentHTML")
}

// AppendText 追加一个文本节点。
func (e *Element) AppendText(text string) error {
	raw, size, release := inputString(text)
	defer release()
	return statusError(C.kui_element_append_text(e.raw(), raw, size), "Element.AppendText")
}
func (e *Element) HasFocus() (bool, error) {
	var value C.int
	err := statusError(C.kui_element_has_focus(e.raw(), &value), "Element.HasFocus")
	return value != 0, err
}

// SetPointerCapture 让 pointerID 的后续移动与松开都发给这个元素，直到释放；鼠标的 pointerID 是 1。
func (e *Element) SetPointerCapture(pointerID int) error {
	return statusError(C.kui_element_set_pointer_capture(e.raw(), C.int(pointerID)), "Element.SetPointerCapture")
}
func (e *Element) ReleasePointerCapture(pointerID int) error {
	return statusError(C.kui_element_release_pointer_capture(e.raw(), C.int(pointerID)), "Element.ReleasePointerCapture")
}
func (e *Element) HasPointerCapture(pointerID int) (bool, error) {
	var value C.int
	err := statusError(C.kui_element_has_pointer_capture(e.raw(), C.int(pointerID), &value), "Element.HasPointerCapture")
	return value != 0, err
}

// GetElementsByClassName 返回带有全部给定类名（空格分隔）的元素。
func (d *Document) GetElementsByClassName(className string) ([]*Element, error) {
	raw, size, release := inputString(className)
	defer release()
	var count C.size_t
	first := C.kui_document_get_elements_by_class_name(d.raw(), raw, size, nil, 0, &count)
	return elementArray(d, first, count, func(items **C.KuiElement, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_document_get_elements_by_class_name(d.raw(), raw, size, items, capacity, out)
	}, "Document.GetElementsByClassName")
}

// GetElementsByTagName 返回指定标签名的元素。
func (d *Document) GetElementsByTagName(tagName string) ([]*Element, error) {
	raw, size, release := inputString(tagName)
	defer release()
	var count C.size_t
	first := C.kui_document_get_elements_by_tag_name(d.raw(), raw, size, nil, 0, &count)
	return elementArray(d, first, count, func(items **C.KuiElement, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_document_get_elements_by_tag_name(d.raw(), raw, size, items, capacity, out)
	}, "Document.GetElementsByTagName")
}

// ElementFromPoint 返回视口坐标处最上层的元素；没有时为 nil。
func (d *Document) ElementFromPoint(x, y float32) (*Element, error) {
	var raw *C.KuiElement
	status := C.kui_document_element_from_point(d.raw(), C.float(x), C.float(y), &raw)
	return d.wrapElement(raw, status, "Document.ElementFromPoint")
}

// MessageBoxType、MessageBoxButtons 与 MessageBoxResult 对应原生 Platform::MsgBox*。
type MessageBoxType int

const (
	MessageBoxInfo MessageBoxType = iota
	MessageBoxWarning
	MessageBoxError
	MessageBoxQuestion
)

type MessageBoxButtons int

const (
	MessageBoxOK MessageBoxButtons = iota
	MessageBoxOKCancel
	MessageBoxYesNo
	MessageBoxYesNoCancel
)

type MessageBoxResult int

const (
	MessageBoxResultOK MessageBoxResult = iota
	MessageBoxResultCancel
	MessageBoxResultYes
	MessageBoxResultNo
)

// MessageBox 弹出模态消息框并等待用户回答。
func MessageBox(title, message string, kind MessageBoxType, buttons MessageBoxButtons) (MessageBoxResult, error) {
	rawTitle, titleLen, releaseTitle := inputString(title)
	defer releaseTitle()
	rawMessage, messageLen, releaseMessage := inputString(message)
	defer releaseMessage()
	var result C.KuiCMessageBoxResult
	err := statusError(C.kui_platform_message_box(rawTitle, titleLen, rawMessage, messageLen,
		C.KuiCMessageBoxType(kind), C.KuiCMessageBoxButtons(buttons), &result), "MessageBox")
	return MessageBoxResult(result), err
}

// OpenURL 用系统默认程序打开网址。
func OpenURL(url string) error {
	raw, size, release := inputString(url)
	defer release()
	return statusError(C.kui_platform_open_url(raw, size), "OpenURL")
}

// Now 返回单调时钟的毫秒数（起点任意，与平台定时器同一时钟）。
func Now() (float64, error) {
	var value C.double
	err := statusError(C.kui_platform_now(&value), "Now")
	return float64(value), err
}

func monitorLookup(call func(*C.KuiCMonitorInfo, *C.char, C.size_t, *C.size_t) C.KuiStatus, context string) (MonitorInfo, error) {
	var value C.KuiCMonitorInfo
	var size C.size_t
	if err := bufferStatus(call(&value, nil, 0, &size), context); err != nil {
		return MonitorInfo{}, err
	}
	buffer := make([]byte, int(size)+1)
	if err := statusError(call(&value, (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), context); err != nil {
		return MonitorInfo{}, err
	}
	return monitorFromC(value, string(buffer[:int(size)])), nil
}

// MonitorFromPoint 返回包含虚拟屏幕坐标 (x, y) 的显示器。
func MonitorFromPoint(x, y float32) (MonitorInfo, error) {
	return monitorLookup(func(out *C.KuiCMonitorInfo, buf *C.char, capacity C.size_t, size *C.size_t) C.KuiStatus {
		return C.kui_platform_get_monitor_from_point(C.float(x), C.float(y), out, buf, capacity, size)
	}, "MonitorFromPoint")
}

// MonitorFromWindow 返回窗口所在的显示器。
func MonitorFromWindow(window *Window) (MonitorInfo, error) {
	// 直接取原生句柄指针，不经 NativeHandle() 的 uintptr 往返。
	var handle unsafe.Pointer
	if err := statusError(C.kui_window_get_native_handle(window.raw(), &handle), "MonitorFromWindow"); err != nil {
		return MonitorInfo{}, err
	}
	return monitorLookup(func(out *C.KuiCMonitorInfo, buf *C.char, capacity C.size_t, size *C.size_t) C.KuiStatus {
		return C.kui_platform_get_monitor_from_window(handle, out, buf, capacity, size)
	}, "MonitorFromWindow")
}
