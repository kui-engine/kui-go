// Package kui 是 KUI Full Profile 的 Go 门面（Windows x64、Linux x64、macOS arm64）。
//
// 所有对象最终委托给稳定的 kui_c C ABI。GUI 调用必须发生在通过
// runtime.LockOSThread 锁定的 UI goroutine；跨 goroutine 工作应通过 PostTask。
package kui

import "fmt"

// Status 是稳定 C ABI 的状态码。
type Status int32

const (
	OK                Status = 0
	ErrInvalidArg     Status = 1
	ErrOutOfMemory    Status = 2
	ErrNotFound       Status = 3
	ErrNotSupported   Status = 4
	ErrParse          Status = 5
	ErrIO             Status = 6
	ErrPluginMissing  Status = 7
	ErrBufferTooSmall Status = 8
	ErrInternal       Status = 99
)

// Error 保留原始 C ABI 状态和调用上下文。
type Error struct {
	Status  Status
	Context string
	Message string
}

func (e *Error) Error() string {
	if e.Context == "" {
		return fmt.Sprintf("kui: %s (%d)", e.Message, e.Status)
	}
	return fmt.Sprintf("kui: %s: %s (%d)", e.Context, e.Message, e.Status)
}

// Version 是已加载运行时版本。
type Version struct{ Major, Minor, Patch int }

type Point struct{ X, Y float32 }
type Size struct{ Width, Height float32 }
type Rect struct{ X, Y, Width, Height float32 }

type MonitorInfo struct {
	Name     string
	Bounds   Rect
	WorkArea Rect
	Scale    float32
	Primary  bool
}

// FileDialogOptions 配置 Windows 原生打开/保存文件对话框。
type FileDialogOptions struct {
	Title, DefaultPath string
	Filters            []string
	AllowMultiple      bool
}

// WindowProps 的零值产生普通、可调整大小、带系统装饰的窗口。
type WindowProps struct {
	Title                      string
	Width, Height              int
	MinWidth, MinHeight        int
	MaxWidth, MaxHeight        int
	NoResize, NoDecorations    bool
	Transparent, AlwaysOnTop   bool
	Maximized, Centered        bool
	Frameless, FramelessShadow bool
	Opacity                    float32
	// 创建时就生效的选项（走 kui_window_create_ex，没有「先显示再隐藏」的闪一下）：
	Sizeable    bool     // 无边框窗口可拖边缘调整大小
	Hidden      bool     // 创建后不显示，之后调 Window.Show
	FrameRadius *float32 // 顶层窗口圆角（CSS px / DIP）；nil 用平台默认
}

type PositioningOptions struct {
	Flip, ClampToViewport  bool
	Offset, ViewportMargin float32
}

type MouseEventType int32

const (
	MouseMoved MouseEventType = iota
	MousePressed
	MouseReleased
	MouseWheel
)

type MouseButton int32

const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
)

type KeyEventType int32

const (
	KeyDown KeyEventType = iota
	KeyUp
	KeyChar
)

type EventTypeTag int32

const (
	EventBase EventTypeTag = iota
	EventMouse
	EventWheel
	EventPointer
	EventKeyboard
	EventFocus
	EventInput
	EventDrag
)

// DropEffect 是拖放效果（DataTransfer.dropEffect）。
type DropEffect int32

const (
	DropNone DropEffect = iota
	DropCopy
	DropMove
	DropLink
)

// EffectAllowed 是拖动源允许的效果（DataTransfer.effectAllowed）。
type EffectAllowed int32

const (
	EffectNone EffectAllowed = iota
	EffectCopy
	EffectMove
	EffectLink
	EffectCopyMove
	EffectCopyLink
	EffectMoveLink
	EffectAll
	EffectUninitialized
)

type PDFOptions struct {
	Title, Author, Subject, Keywords, Creator, CreationDate string
	PageWidthPT, PageHeightPT                               float32
	RasterizeFonts, GenerateOutline                         bool
}

type ExtensionValueType int32

const (
	ExtensionNil ExtensionValueType = iota
	ExtensionInt
	ExtensionFloat
	ExtensionString
	ExtensionBool
)

type ExtensionValue struct {
	Type   ExtensionValueType
	Int    int64
	Float  float64
	String string
	Bool   bool
}

type ExtensionPerfStats struct {
	LastTickMS, LastPaintMS, MaxTickMS, MaxPaintMS float64
	TickCount, PaintCount                          uint64
	TotalTickMS, TotalPaintMS                      float64
}

type HTTPResponse struct {
	Status      int
	Body        []byte
	ContentType string
	Error       string
}
