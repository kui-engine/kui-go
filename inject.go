//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
)

// 结构化的注入、盒模型与 DevTools 调用（T-DEBT-SDK-GO-LOWLEVEL-SIGNATURES）。原来的
// DispatchMouse / DispatchKey / BoxModel / Event.Flags / Event.Modifiers 照旧可用，保留一个版本后移除。

// DevTools 注入事件的修饰键位（与自定义元素输入事件的 InputMod* 位不同）。
const (
	InjectAlt   = 1
	InjectCtrl  = 2
	InjectMeta  = 4
	InjectShift = 8
)

// MouseInjection 是一次注入的鼠标事件（视口坐标）；滚轮事件用 DeltaX / DeltaY。
type MouseInjection struct {
	Kind       MouseEventType
	X, Y       float32
	Button     MouseButton
	Modifiers  int // Inject* 位
	ClickCount int
	DeltaX     float32
	DeltaY     float32
}

// DispatchMouseEvent 经 DevTools 注入鼠标事件，与真实输入走同一条派发路径。
func (d *Document) DispatchMouseEvent(event MouseInjection) error {
	return d.DispatchMouse(event.Kind, event.X, event.Y, event.Button, event.Modifiers, event.ClickCount, event.DeltaX, event.DeltaY)
}

// KeyInjection 是一次注入的键盘事件；Text 是 KeyChar 输入的文字。
type KeyInjection struct {
	Kind       KeyEventType
	Key, Code  string // 例如 "a" / "KeyA"、"Enter" / "Enter"
	Modifiers  int    // Inject* 位
	Text       string
	AutoRepeat bool
}

// DispatchKeyEvent 经 DevTools 注入键盘事件（发给获得焦点的元素）。
func (d *Document) DispatchKeyEvent(event KeyInjection) error {
	return d.DispatchKey(event.Kind, event.Key, event.Code, event.Modifiers, event.Text, event.AutoRepeat)
}

// BoxModelInfo 是元素四层盒的顶点（视口坐标，左上起顺时针）与内容尺寸。
type BoxModelInfo struct {
	Content, Padding, Border, Margin [4]Point
	Width, Height                    float32
}

// ElementBoxModel 返回第一个匹配 selector 的元素的盒模型；没有匹配返回 ErrNotFound。
func (d *Document) ElementBoxModel(selector string) (BoxModelInfo, error) {
	quads, size, err := d.BoxModel(selector)
	if err != nil {
		return BoxModelInfo{}, err
	}
	info := BoxModelInfo{Width: size.Width, Height: size.Height}
	for box, target := range []*[4]Point{&info.Content, &info.Padding, &info.Border, &info.Margin} {
		for corner := 0; corner < 4; corner++ {
			target[corner] = Point{X: quads[box*8+corner*2], Y: quads[box*8+corner*2+1]}
		}
	}
	return info, nil
}

// EventFlags 是事件的三个标志。
type EventFlags struct {
	Bubbles, Cancelable, DefaultPrevented bool
}

// FlagValues 返回事件的标志。
func (e *Event) FlagValues() (EventFlags, error) {
	bubbles, cancelable, prevented, err := e.Flags()
	return EventFlags{Bubbles: bubbles, Cancelable: cancelable, DefaultPrevented: prevented}, err
}

// KeyModifiers 是键盘事件按下的修饰键。
type KeyModifiers struct {
	Ctrl, Shift, Alt, Meta bool
}

// ModifierKeys 返回键盘事件的修饰键；不是键盘事件时返回 ErrNotSupported。
func (e *Event) ModifierKeys() (KeyModifiers, error) {
	ctrl, shift, alt, meta, err := e.Modifiers()
	return KeyModifiers{Ctrl: ctrl, Shift: shift, Alt: alt, Meta: meta}, err
}

// DevToolsError 是 DevTools 命令返回的错误。
type DevToolsError struct {
	Code    int
	Message string
}

func (e *DevToolsError) Error() string {
	return fmt.Sprintf("DevTools error %d: %s", e.Code, e.Message)
}

var devToolsCallID atomic.Int64

// DevToolsCall 发送一条 DevTools 命令（CDP 风格的 method 与 params，params 可为 nil），把结果解到 result
// （可为 nil）。命令本身失败时返回 *DevToolsError。
func (d *Document) DevToolsCall(method string, params any, result any) error {
	request := struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params any    `json:"params,omitempty"`
	}{ID: devToolsCallID.Add(1), Method: method, Params: params}
	command, err := json.Marshal(request)
	if err != nil {
		return err
	}
	response, err := d.DispatchDevToolsCommand(string(command))
	if err != nil {
		return err
	}
	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(response), &envelope); err != nil {
		return fmt.Errorf("DevTools %s: invalid response: %w", method, err)
	}
	if envelope.Error != nil {
		return &DevToolsError{Code: envelope.Error.Code, Message: envelope.Error.Message}
	}
	if result == nil || len(envelope.Result) == 0 {
		return nil
	}
	return json.Unmarshal(envelope.Result, result)
}
