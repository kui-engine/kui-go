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

// 3.61.8 起的自定义元素自绘、tick 与输入（C ABI 第 17 节），回调都在 UI 线程执行。
// 键盘与输入法事件只到获得焦点的元素（tabindex + Focus）；鼠标与滚轮坐标在元素的内容框里。

// InputModifier 是自定义元素输入事件的修饰键位（与 DispatchMouse / DispatchKey 的位不同）。
const (
	InputModCtrl  = 1
	InputModShift = 2
	InputModAlt   = 4
	InputModMeta  = 8
)

type CustomKeyType int32

const (
	CustomKeyDown CustomKeyType = iota
	CustomKeyUp
)

type CustomMouseType int32

const (
	CustomMouseDown CustomMouseType = iota
	CustomMouseUp
	CustomMouseMove
	CustomMouseEnter
	CustomMouseLeave
)

// IMEEventType 是输入法组合的阶段，也用于 Document.DispatchIME。
type IMEEventType int32

const (
	IMEStart IMEEventType = iota
	IMEUpdate
	IMECommit
	IMECancel
)

type DecorationStyle int32

const (
	DecorationSolid DecorationStyle = iota
	DecorationDouble
	DecorationDotted
	DecorationDashed
	DecorationWavy
)

type CustomKeyEvent struct {
	Type      CustomKeyType
	Key, Code string
	Modifiers int
	Repeat    bool
}

type CustomMouseEvent struct {
	Type       CustomMouseType
	X, Y       float32
	Button     int // 0 左、1 中、2 右；移动 / 进入 / 离开为 -1
	Buttons    int
	Modifiers  int
	ClickCount int
}

type CustomWheelEvent struct {
	X, Y, DeltaX, DeltaY float32
	Modifiers            int
}

type CustomIMEEvent struct {
	Type          IMEEventType
	Text          string // 组合中的文字（Update）
	CaretPos      int    // Text 里的字节偏移（Update）
	CommittedText string // 最终文字（Commit）
}

// TextStyle 是绘制 / 测量文字的字体；零值是 sans-serif 14px 常规。
type TextStyle struct {
	FontFamily    string
	FontSize      float32
	FontWeight    uint16
	Italic        bool
	LetterSpacing float32
	WordSpacing   float32
}

type TextMetrics struct {
	Width, Ascent, Descent, LineHeight float32
}

// CornerRadius 是一个角的横向 / 纵向半径。
type CornerRadius struct{ X, Y float32 }

type CornerRadii struct{ TopLeft, TopRight, BottomRight, BottomLeft CornerRadius }

// UniformCornerRadii 是四个角相同的圆角。
func UniformCornerRadii(radius float32) CornerRadii {
	r := CornerRadius{radius, radius}
	return CornerRadii{r, r, r, r}
}

func (r CornerRadii) c() C.KuiCCornerRadii {
	return C.KuiCCornerRadii{
		top_left_x: C.float(r.TopLeft.X), top_left_y: C.float(r.TopLeft.Y),
		top_right_x: C.float(r.TopRight.X), top_right_y: C.float(r.TopRight.Y),
		bottom_right_x: C.float(r.BottomRight.X), bottom_right_y: C.float(r.BottomRight.Y),
		bottom_left_x: C.float(r.BottomLeft.X), bottom_left_y: C.float(r.BottomLeft.Y),
	}
}

// cTextStyle 转成 C 结构；字体名放在 C 内存里，release 释放。
func cTextStyle(style *TextStyle) (*C.KuiCTextStyle, func()) {
	if style == nil {
		return nil, func() {}
	}
	value := &C.KuiCTextStyle{font_size: C.float(style.FontSize), font_weight: C.uint16_t(style.FontWeight),
		italic: boolInt(style.Italic), letter_spacing: C.float(style.LetterSpacing), word_spacing: C.float(style.WordSpacing)}
	if style.FontFamily == "" {
		return value, func() {}
	}
	family := C.CString(style.FontFamily)
	value.font_family = family
	return value, func() { C.free(unsafe.Pointer(family)) }
}

// PaintContext 只在 Paint 回调期间有效；坐标是内容框里的 CSS 像素，颜色是 0xRRGGBBAA。
// 回调里做的变换、裁剪与透明度在回调结束时自动撤销。
type PaintContext struct {
	raw                             *C.KuiPaintContext
	Width, Height, DevicePixelRatio float32
}

func (p *PaintContext) FillRect(x, y, width, height float32, rgba uint32) error {
	return statusError(C.kui_paint_fill_rect(p.raw, C.float(x), C.float(y), C.float(width), C.float(height), C.uint32_t(rgba)), "PaintContext.FillRect")
}
func (p *PaintContext) FillRoundedRect(x, y, width, height float32, radii CornerRadii, rgba uint32) error {
	r := radii.c()
	return statusError(C.kui_paint_fill_rounded_rect(p.raw, C.float(x), C.float(y), C.float(width), C.float(height), &r, C.uint32_t(rgba)), "PaintContext.FillRoundedRect")
}
func (p *PaintContext) StrokeRect(x, y, width, height, lineWidth float32, rgba uint32) error {
	return statusError(C.kui_paint_stroke_rect(p.raw, C.float(x), C.float(y), C.float(width), C.float(height), C.float(lineWidth), C.uint32_t(rgba)), "PaintContext.StrokeRect")
}
func (p *PaintContext) StrokeRoundedRect(x, y, width, height float32, radii CornerRadii, lineWidth float32, rgba uint32) error {
	r := radii.c()
	return statusError(C.kui_paint_stroke_rounded_rect(p.raw, C.float(x), C.float(y), C.float(width), C.float(height), &r, C.float(lineWidth), C.uint32_t(rgba)), "PaintContext.StrokeRoundedRect")
}

// DrawText 在 (x, y) 画文字：y 是文字行盒的顶部（不是基线），字形在 y 下方，基线在 y + ascent
// （ascent 取自同一 style 的 Element.MeasureText）；style 为 nil 用 sans-serif 14px。
func (p *PaintContext) DrawText(text string, x, y float32, style *TextStyle, rgba uint32) error {
	raw, size, release := inputString(text)
	defer release()
	cStyle, releaseStyle := cTextStyle(style)
	defer releaseStyle()
	return statusError(C.kui_paint_draw_text(p.raw, raw, size, C.float(x), C.float(y), cStyle, C.uint32_t(rgba)), "PaintContext.DrawText")
}

// DrawPixels 画 width×height 个预乘的 0xAARRGGBB 像素，缩放到目标矩形。
func (p *PaintContext) DrawPixels(pixels []uint32, width, height int, x, y, destWidth, destHeight float32) error {
	if width <= 0 || height <= 0 || len(pixels) < width*height {
		return &Error{Status: ErrInvalidArg, Context: "PaintContext.DrawPixels", Message: "pixels shorter than width*height"}
	}
	return statusError(C.kui_paint_draw_pixels(p.raw, (*C.uint32_t)(unsafe.Pointer(&pixels[0])), C.uint32_t(width), C.uint32_t(height), C.float(x), C.float(y), C.float(destWidth), C.float(destHeight)), "PaintContext.DrawPixels")
}

// DrawImage 画一个图片文件（与 <img> 同一套加载与缓存）；src 是图片像素坐标。
func (p *PaintContext) DrawImage(path string, src, dest Rect) error {
	raw, size, release := inputString(path)
	defer release()
	from := C.KuiRect{x: C.float(src.X), y: C.float(src.Y), width: C.float(src.Width), height: C.float(src.Height)}
	to := C.KuiRect{x: C.float(dest.X), y: C.float(dest.Y), width: C.float(dest.Width), height: C.float(dest.Height)}
	return statusError(C.kui_paint_draw_image(p.raw, raw, size, &from, &to), "PaintContext.DrawImage")
}
func (p *PaintContext) DrawDecorationLine(x, y, width, thickness float32, style DecorationStyle, rgba uint32) error {
	return statusError(C.kui_paint_draw_decoration_line(p.raw, C.float(x), C.float(y), C.float(width), C.float(thickness), C.KuiCDecorationStyle(style), C.uint32_t(rgba)), "PaintContext.DrawDecorationLine")
}
func (p *PaintContext) Save() error {
	return statusError(C.kui_paint_save(p.raw), "PaintContext.Save")
}

// Restore 撤销最近一次 Save；没有对应的 Save 时返回 ErrInvalidArg。
func (p *PaintContext) Restore() error {
	return statusError(C.kui_paint_restore(p.raw), "PaintContext.Restore")
}
func (p *PaintContext) Translate(dx, dy float32) error {
	return statusError(C.kui_paint_translate(p.raw, C.float(dx), C.float(dy)), "PaintContext.Translate")
}
func (p *PaintContext) Scale(sx, sy float32) error {
	return statusError(C.kui_paint_scale(p.raw, C.float(sx), C.float(sy)), "PaintContext.Scale")
}
func (p *PaintContext) Rotate(degrees float32) error {
	return statusError(C.kui_paint_rotate(p.raw, C.float(degrees)), "PaintContext.Rotate")
}

// SetOpacity 让之后画的内容乘以 opacity（0..1），到对应的 Restore 为止。
func (p *PaintContext) SetOpacity(opacity float32) error {
	return statusError(C.kui_paint_set_opacity(p.raw, C.float(opacity)), "PaintContext.SetOpacity")
}
func (p *PaintContext) ClipRect(x, y, width, height float32) error {
	return statusError(C.kui_paint_clip_rect(p.raw, C.float(x), C.float(y), C.float(width), C.float(height)), "PaintContext.ClipRect")
}
func (p *PaintContext) ClipRoundedRect(x, y, width, height float32, radii CornerRadii) error {
	r := radii.c()
	return statusError(C.kui_paint_clip_rounded_rect(p.raw, C.float(x), C.float(y), C.float(width), C.float(height), &r), "PaintContext.ClipRoundedRect")
}

// MeasureText 用与 KUI 自己相同的排版测量文字；窗口创建之前返回 ErrNotSupported。
func (e *Element) MeasureText(text string, style *TextStyle) (TextMetrics, error) {
	raw, size, release := inputString(text)
	defer release()
	cStyle, releaseStyle := cTextStyle(style)
	defer releaseStyle()
	var out C.KuiCTextMetrics
	err := statusError(C.kui_element_measure_text(e.raw(), raw, size, cStyle, &out), "Element.MeasureText")
	return TextMetrics{Width: float32(out.width), Ascent: float32(out.ascent), Descent: float32(out.descent), LineHeight: float32(out.line_height)}, err
}

// RequestRepaint 让元素在下一帧重绘（自绘元素的状态变了之后调用）。
func (e *Element) RequestRepaint() error {
	return statusError(C.kui_element_request_repaint(e.raw()), "Element.RequestRepaint")
}

// InvalidateRect 只重绘内容框里的这块矩形。
func (e *Element) InvalidateRect(x, y, width, height float32) error {
	return statusError(C.kui_element_invalidate_rect(e.raw(), C.float(x), C.float(y), C.float(width), C.float(height)), "Element.InvalidateRect")
}

// DispatchIME 向获得焦点的元素注入一步输入法组合（与系统输入法走同一条路径）。
func (d *Document) DispatchIME(kind IMEEventType, text string) error {
	raw, size, release := inputString(text)
	defer release()
	return statusError(C.kui_devtools_dispatch_ime_event(d.raw(), C.KuiCImeEventType(kind), raw, size), "Document.DispatchIME")
}

// ---------------------------------------------------------------------------
// 回调
// ---------------------------------------------------------------------------

func customState(handle C.uintptr_t) (*customElementState, bool) {
	state, ok := cgo.Handle(handle).Value().(*customElementState)
	return state, ok
}

func (s *customElementState) element(raw *C.KuiElement) *Element {
	return &Element{handle: unsafe.Pointer(raw), document: s.document, borrowed: true}
}

//export goKuiCustomPaintCallback
func goKuiCustomPaintCallback(raw *C.KuiElement, paint *C.KuiPaintContext, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := customState(handle)
	if !ok || state.definition.Paint == nil {
		return
	}
	var width, height, dpr C.float
	C.kui_paint_get_info(paint, &width, &height, &dpr)
	state.definition.Paint(state.element(raw), &PaintContext{raw: paint, Width: float32(width), Height: float32(height), DevicePixelRatio: float32(dpr)})
}

//export goKuiCustomTickCallback
func goKuiCustomTickCallback(raw *C.KuiElement, delta C.double, handle C.uintptr_t) {
	defer containCallbackPanic()
	if state, ok := customState(handle); ok && state.definition.Tick != nil {
		state.definition.Tick(state.element(raw), float64(delta))
	}
}

//export goKuiCustomNeedsTickCallback
func goKuiCustomNeedsTickCallback(raw *C.KuiElement, handle C.uintptr_t) (result C.int) {
	defer containCallbackPanic()
	if state, ok := customState(handle); ok && state.definition.NeedsTick != nil {
		result = boolInt(state.definition.NeedsTick(state.element(raw)))
	}
	return result
}

//export goKuiCustomKeyCallback
func goKuiCustomKeyCallback(raw *C.KuiElement, event *C.KuiCCustomKeyEvent, handle C.uintptr_t) (result C.int) {
	defer containCallbackPanic()
	if state, ok := customState(handle); ok && state.definition.Key != nil {
		result = boolInt(state.definition.Key(state.element(raw), CustomKeyEvent{
			Type: CustomKeyType(event._type), Key: goStringN(event.key, event.key_len), Code: goStringN(event.code, event.code_len),
			Modifiers: int(event.modifiers), Repeat: event.is_repeat != 0,
		}))
	}
	return result
}

//export goKuiCustomMouseCallback
func goKuiCustomMouseCallback(raw *C.KuiElement, event *C.KuiCCustomMouseEvent, handle C.uintptr_t) (result C.int) {
	defer containCallbackPanic()
	if state, ok := customState(handle); ok && state.definition.Mouse != nil {
		result = boolInt(state.definition.Mouse(state.element(raw), CustomMouseEvent{
			Type: CustomMouseType(event._type), X: float32(event.x), Y: float32(event.y), Button: int(event.button),
			Buttons: int(event.buttons), Modifiers: int(event.modifiers), ClickCount: int(event.click_count),
		}))
	}
	return result
}

//export goKuiCustomWheelCallback
func goKuiCustomWheelCallback(raw *C.KuiElement, event *C.KuiCCustomWheelEvent, handle C.uintptr_t) (result C.int) {
	defer containCallbackPanic()
	if state, ok := customState(handle); ok && state.definition.Wheel != nil {
		result = boolInt(state.definition.Wheel(state.element(raw), CustomWheelEvent{
			X: float32(event.x), Y: float32(event.y), DeltaX: float32(event.delta_x), DeltaY: float32(event.delta_y),
			Modifiers: int(event.modifiers),
		}))
	}
	return result
}

//export goKuiCustomIMECallback
func goKuiCustomIMECallback(raw *C.KuiElement, event *C.KuiCCustomImeEvent, handle C.uintptr_t) (result C.int) {
	defer containCallbackPanic()
	if state, ok := customState(handle); ok && state.definition.IME != nil {
		result = boolInt(state.definition.IME(state.element(raw), CustomIMEEvent{
			Type: IMEEventType(event._type), Text: goStringN(event.text, event.text_len), CaretPos: int(event.caret_pos),
			CommittedText: goStringN(event.committed_text, event.committed_len),
		}))
	}
	return result
}

// customHooks 填第 17 节的回调指针（只填设置了的）。
func customHooks(definition CustomElementDefinition, raw *C.KuiCCustomElementDefinition) {
	if definition.Paint != nil {
		raw.paint = C.kui_go_custom_paint_trampoline()
	}
	if definition.Tick != nil {
		raw.tick = C.kui_go_custom_tick_trampoline()
	}
	if definition.NeedsTick != nil {
		raw.needs_tick = C.kui_go_custom_needs_tick_trampoline()
	}
	if definition.Key != nil {
		raw.key = C.kui_go_custom_key_trampoline()
	}
	if definition.Mouse != nil {
		raw.mouse = C.kui_go_custom_mouse_trampoline()
	}
	if definition.Wheel != nil {
		raw.wheel = C.kui_go_custom_wheel_trampoline()
	}
	if definition.IME != nil {
		raw.ime = C.kui_go_custom_ime_trampoline()
	}
}
