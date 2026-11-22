//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

/*
#cgo CFLAGS: -std=c11 -I${SRCDIR}
#include <stdlib.h>
#include "loader.h"
#include "callbacks.h"
*/
import "C"

import "unsafe"

// 3.61.8 起的弹出层、<input> 选择器与游戏音频（C ABI 第 12 节）。全部只能在 UI 线程调用。

// ---------------------------------------------------------------------------
// 弹出层
// ---------------------------------------------------------------------------

// PopupType 是弹出层的种类。
type PopupType int

const (
	PopupTooltip PopupType = iota
	PopupContextMenu
	PopupCustom
)

// PopupPosition 是提示框相对锚点的方向。
type PopupPosition int

const (
	PopupTop PopupPosition = iota
	PopupBottom
	PopupLeft
	PopupRight
)

// PopupMode 选择弹出层画在页面上（默认）还是用原生弹出窗口。
type PopupMode int

const (
	PopupModeOverlay PopupMode = iota
	PopupModeNativeWindow
)

// TooltipTheme 是提示框配色，颜色为 0xRRGGBBAA。
type TooltipTheme struct {
	Background, Foreground, Border uint32
	Radius                         float32
}

// Popups 是文档的弹出层（提示框、右键菜单）。
type Popups struct{ document *Document }

// Popups 返回文档的弹出层管理。
func (d *Document) Popups() *Popups { return &Popups{document: d} }

func (p *Popups) raw() *C.KuiDocument { return p.document.raw() }

// ShowTooltip 在锚点矩形（视口坐标）旁显示提示框，返回 id。
func (p *Popups) ShowTooltip(text string, anchor Rect, position PopupPosition) (uint32, error) {
	raw, size, release := inputString(text)
	defer release()
	rect := C.KuiRect{x: C.float(anchor.X), y: C.float(anchor.Y), width: C.float(anchor.Width), height: C.float(anchor.Height)}
	var id C.uint32_t
	err := statusError(C.kui_document_show_tooltip(p.raw(), raw, size, &rect, C.KuiCPopupPosition(position), &id), "Popups.ShowTooltip")
	return uint32(id), err
}

// ShowContextMenu 在 (x, y) 显示右键菜单（只用 MenuItem 的 Text / ID / Separator / Disabled），
// 选中某项时以它的 ID 调用 onActivate；菜单关闭后句柄自动释放。
func (p *Popups) ShowContextMenu(x, y float32, items []MenuItem, onActivate func(id uint32)) (uint32, error) {
	if onActivate == nil || len(items) == 0 {
		return 0, &Error{Status: ErrInvalidArg, Context: "Popups.ShowContextMenu", Message: "nil callback or no items"}
	}
	raw := (*C.KuiCMenuItem)(C.calloc(C.size_t(len(items)), C.size_t(unsafe.Sizeof(C.KuiCMenuItem{}))))
	defer C.free(unsafe.Pointer(raw))
	slots := unsafe.Slice(raw, len(items))
	var texts []*C.char
	defer func() {
		for _, text := range texts {
			C.free(unsafe.Pointer(text))
		}
	}()
	for i, item := range items {
		slots[i].id = C.uint32_t(item.ID)
		slots[i].separator = boolInt(item.Separator)
		slots[i].disabled = boolInt(item.Disabled)
		if !item.Separator {
			text := C.CString(item.Text)
			texts = append(texts, text)
			slots[i].text = text
			slots[i].text_len = C.size_t(len(item.Text))
		}
	}
	h, user := callbackHandle(onActivate)
	var id C.uint32_t
	status := C.kui_document_show_context_menu(p.raw(), C.float(x), C.float(y), raw, C.size_t(len(items)),
		C.kui_go_menu_activate_trampoline(), user, C.kui_go_release_trampoline(), &id)
	if err := statusError(status, "Popups.ShowContextMenu"); err != nil {
		h.Delete()
		return 0, err
	}
	return uint32(id), nil
}

// Dismiss 关闭一个弹出层（未知 id 什么也不做）。
func (p *Popups) Dismiss(id uint32) error {
	return statusError(C.kui_document_dismiss_popup(p.raw(), C.uint32_t(id)), "Popups.Dismiss")
}

// DismissAll 关闭全部弹出层。
func (p *Popups) DismissAll() error {
	return statusError(C.kui_document_dismiss_all_popups(p.raw()), "Popups.DismissAll")
}

// DismissType 关闭某一种弹出层。
func (p *Popups) DismissType(kind PopupType) error {
	return statusError(C.kui_document_dismiss_popups_of_type(p.raw(), C.KuiCPopupType(kind)), "Popups.DismissType")
}

// Any 报告是否有弹出层在显示。
func (p *Popups) Any() (bool, error) {
	var value C.int
	err := statusError(C.kui_document_has_popups(p.raw(), &value), "Popups.Any")
	return value != 0, err
}

// Has 报告某一种弹出层是否在显示。
func (p *Popups) Has(kind PopupType) (bool, error) {
	var value C.int
	err := statusError(C.kui_document_has_popup_of_type(p.raw(), C.KuiCPopupType(kind), &value), "Popups.Has")
	return value != 0, err
}

// SetTooltipDelay 设置悬停多久后显示 title 提示（毫秒，默认 500）。
func (p *Popups) SetTooltipDelay(ms float64) error {
	return statusError(C.kui_document_set_tooltip_delay(p.raw(), C.double(ms)), "Popups.SetTooltipDelay")
}

// TooltipDelay 返回提示延迟（毫秒）。
func (p *Popups) TooltipDelay() (float64, error) {
	var value C.double
	err := statusError(C.kui_document_get_tooltip_delay(p.raw(), &value), "Popups.TooltipDelay")
	return float64(value), err
}

// SetTooltipTheme 设置提示框配色。
func (p *Popups) SetTooltipTheme(theme TooltipTheme) error {
	raw := C.KuiCTooltipTheme{background: C.uint32_t(theme.Background), foreground: C.uint32_t(theme.Foreground),
		border: C.uint32_t(theme.Border), radius: C.float(theme.Radius)}
	return statusError(C.kui_document_set_tooltip_theme(p.raw(), &raw), "Popups.SetTooltipTheme")
}

// TooltipTheme 返回提示框配色。
func (p *Popups) TooltipTheme() (TooltipTheme, error) {
	var raw C.KuiCTooltipTheme
	err := statusError(C.kui_document_get_tooltip_theme(p.raw(), &raw), "Popups.TooltipTheme")
	return TooltipTheme{Background: uint32(raw.background), Foreground: uint32(raw.foreground),
		Border: uint32(raw.border), Radius: float32(raw.radius)}, err
}

// SetMode 选择弹出层的呈现方式。
func (p *Popups) SetMode(mode PopupMode) error {
	return statusError(C.kui_document_set_popup_mode(p.raw(), C.KuiCPopupMode(mode)), "Popups.SetMode")
}

// SetDefaultContextMenuEnabled 开关内置的默认右键菜单（输入框：剪切 / 复制 / 粘贴 / 全选；其它地方：复制 / 全选）。
// 默认关闭：右键和菜单键照常派发 contextmenu 事件，自己弹的菜单不受影响。对整个文档生效，换页后保持；页面也可以
// 写 <html context-menu> 开启。开了检查器而默认菜单关闭时，右键只出「检查元素」。
func (p *Popups) SetDefaultContextMenuEnabled(enabled bool) error {
	return statusError(C.kui_document_set_default_context_menu_enabled(p.raw(), boolInt(enabled)),
		"Popups.SetDefaultContextMenuEnabled")
}

// DefaultContextMenuEnabled 报告是否用代码开启了默认右键菜单（不含页面上的 <html context-menu>）。
func (p *Popups) DefaultContextMenuEnabled() (bool, error) {
	var value C.int
	err := statusError(C.kui_document_get_default_context_menu_enabled(p.raw(), &value), "Popups.DefaultContextMenuEnabled")
	return value != 0, err
}

// ---------------------------------------------------------------------------
// <input> 选择器
// ---------------------------------------------------------------------------

// PickerMode 是选择器面板的种类。
type PickerMode int

const (
	PickerNone PickerMode = iota
	PickerDate
	PickerTime
	PickerDateTimeLocal
	PickerColor
)

// InputPickerState 是选择器当前状态。
type InputPickerState struct {
	Open, Enabled             bool
	Mode                      PickerMode
	VisibleYear, VisibleMonth int // 日历页（月份 1-12）
	ColorH, ColorS, ColorV    float32
	Hour, Minute              int
}

// OpenInputPicker 为 <input> 打开选择器；PickerNone 按 input 的 type 选面板。没有面板的类型或选择器被禁用时
// 返回 ErrNotSupported，非 <input> 元素返回 ErrInvalidArg。
func (e *Element) OpenInputPicker(mode PickerMode) error {
	return statusError(C.kui_element_open_input_picker(e.raw(), C.KuiCPickerMode(mode)), "Element.OpenInputPicker")
}

// InputPicker 是文档的 <input> 日期 / 时间 / 取色面板。
type InputPicker struct{ document *Document }

// InputPicker 返回文档的选择器。
func (d *Document) InputPicker() *InputPicker { return &InputPicker{document: d} }

func (p *InputPicker) raw() *C.KuiDocument { return p.document.raw() }

// Close 关闭选择器。
func (p *InputPicker) Close() error {
	return statusError(C.kui_document_close_input_picker(p.raw()), "InputPicker.Close")
}

// Commit 把面板的值写回 input（触发 input / change）并关闭。
func (p *InputPicker) Commit() error {
	return statusError(C.kui_document_commit_input_picker(p.raw()), "InputPicker.Commit")
}

// State 返回选择器当前状态。
func (p *InputPicker) State() (InputPickerState, error) {
	var raw C.KuiCInputPickerState
	err := statusError(C.kui_document_get_input_picker_state(p.raw(), &raw), "InputPicker.State")
	return InputPickerState{Open: raw.open != 0, Enabled: raw.enabled != 0, Mode: PickerMode(raw.mode),
		VisibleYear: int(raw.visible_year), VisibleMonth: int(raw.visible_month),
		ColorH: float32(raw.color_h), ColorS: float32(raw.color_s), ColorV: float32(raw.color_v),
		Hour: int(raw.hour), Minute: int(raw.minute)}, err
}

// Target 返回选择器所属的 input；关闭时返回 nil。
func (p *InputPicker) Target() (*Element, error) {
	var raw *C.KuiElement
	status := C.kui_document_get_input_picker_target(p.raw(), &raw)
	return p.document.wrapElement(raw, status, "InputPicker.Target")
}

// SetEnabled 启用或禁用选择器（禁用后 input 点击不再弹出面板）。
func (p *InputPicker) SetEnabled(enabled bool) error {
	return statusError(C.kui_document_set_input_picker_enabled(p.raw(), boolInt(enabled)), "InputPicker.SetEnabled")
}

// SetMonth 切换日历页（月份 1-12）。
func (p *InputPicker) SetMonth(year, month int) error {
	return statusError(C.kui_document_set_input_picker_month(p.raw(), C.int(year), C.int(month)), "InputPicker.SetMonth")
}

// SetColorHSV 设置取色面板的颜色（色相 0-360，饱和度 / 明度 0-1）。
func (p *InputPicker) SetColorHSV(h, s, v float32) error {
	return statusError(C.kui_document_set_input_picker_color_hsv(p.raw(), C.float(h), C.float(s), C.float(v)), "InputPicker.SetColorHSV")
}

// SetTime 设置时间面板（0-23 时、0-59 分）。
func (p *InputPicker) SetTime(hour, minute int) error {
	return statusError(C.kui_document_set_input_picker_time(p.raw(), C.int(hour), C.int(minute)), "InputPicker.SetTime")
}

// ---------------------------------------------------------------------------
// 游戏音频
// ---------------------------------------------------------------------------

// AudioGroup 是音频分组；AudioAllGroups 用于「全部分组」。
type AudioGroup int

const (
	AudioMaster    AudioGroup = 0
	AudioMusic     AudioGroup = 1
	AudioSFX       AudioGroup = 2
	AudioUI        AudioGroup = 3
	AudioAllGroups AudioGroup = -1
)

// AudioPlayOptions 是播放选项；nil 表示 SFX、音量 1、声像 0、音高 1。
type AudioPlayOptions struct {
	Group       AudioGroup
	Volume      float32 // 0-1，分组 / 主音量缩放之前
	Pan         float32 // -1（左）到 1（右）
	Pitch       float32 // 播放速度倍率
	Loop        bool
	PreloadOnly bool
}

// SoundInfo 是一个声音的状态。
type SoundInfo struct {
	Group              AudioGroup
	Volume, Pan, Pitch float32
	Playing, Looping   bool
	CurrentTime        float64 // 秒
	Duration           float64
}

// Audio 是文档的 AudioManager（需要音频媒体插件）。
type Audio struct{ document *Document }

// Audio 返回文档的音频管理。
func (d *Document) Audio() *Audio { return &Audio{document: d} }

func (a *Audio) raw() *C.KuiDocument { return a.document.raw() }

// Available 报告是否有音频设备与媒体插件。
func (a *Audio) Available() (bool, error) {
	var value C.int
	err := statusError(C.kui_document_audio_is_available(a.raw(), &value), "Audio.Available")
	return value != 0, err
}

// Play 播放文件并返回句柄。音频不可用返回 ErrNotSupported，文件放不了返回 ErrIO。
func (a *Audio) Play(path string, options *AudioPlayOptions) (uint32, error) {
	raw, size, release := inputString(path)
	defer release()
	var rawOptions *C.KuiCAudioPlayOptions
	if options != nil {
		rawOptions = &C.KuiCAudioPlayOptions{group: C.KuiCAudioGroup(options.Group), volume: C.float(options.Volume),
			pan: C.float(options.Pan), pitch: C.float(options.Pitch), loop: boolInt(options.Loop),
			preload_only: boolInt(options.PreloadOnly)}
	}
	var handle C.uint32_t
	err := statusError(C.kui_document_audio_play(a.raw(), raw, size, rawOptions, &handle), "Audio.Play")
	return uint32(handle), err
}

// Stop 停止一个声音（未知句柄什么也不做）。
func (a *Audio) Stop(handle uint32) error {
	return statusError(C.kui_document_audio_stop(a.raw(), C.uint32_t(handle)), "Audio.Stop")
}

// SetPaused 暂停或继续一个声音。
func (a *Audio) SetPaused(handle uint32, paused bool) error {
	return statusError(C.kui_document_audio_set_paused(a.raw(), C.uint32_t(handle), boolInt(paused)), "Audio.SetPaused")
}

// SetVolume 设置一个声音的音量。
func (a *Audio) SetVolume(handle uint32, volume float32) error {
	return statusError(C.kui_document_audio_set_volume(a.raw(), C.uint32_t(handle), C.float(volume)), "Audio.SetVolume")
}

// SetPan 设置一个声音的声像。
func (a *Audio) SetPan(handle uint32, pan float32) error {
	return statusError(C.kui_document_audio_set_pan(a.raw(), C.uint32_t(handle), C.float(pan)), "Audio.SetPan")
}

// SetPitch 设置一个声音的音高（播放速度倍率）。
func (a *Audio) SetPitch(handle uint32, pitch float32) error {
	return statusError(C.kui_document_audio_set_pitch(a.raw(), C.uint32_t(handle), C.float(pitch)), "Audio.SetPitch")
}

// Seek 跳到某个时刻（秒）。
func (a *Audio) Seek(handle uint32, seconds float64) error {
	return statusError(C.kui_document_audio_seek(a.raw(), C.uint32_t(handle), C.double(seconds)), "Audio.Seek")
}

// Info 返回一个声音的状态；未知或已结束的句柄 ok 为 false。
func (a *Audio) Info(handle uint32) (info SoundInfo, ok bool, err error) {
	var raw C.KuiCSoundInfo
	status := C.kui_document_audio_get_info(a.raw(), C.uint32_t(handle), &raw)
	if status == C.KUI_ERR_NOT_FOUND {
		return SoundInfo{}, false, nil
	}
	if err := statusError(status, "Audio.Info"); err != nil {
		return SoundInfo{}, false, err
	}
	return SoundInfo{Group: AudioGroup(raw.group), Volume: float32(raw.volume), Pan: float32(raw.pan),
		Pitch: float32(raw.pitch), Playing: raw.playing != 0, Looping: raw.looping != 0,
		CurrentTime: float64(raw.current_time), Duration: float64(raw.duration)}, true, nil
}

// SetGroup 设置一个分组的音量（0-1）与静音。
func (a *Audio) SetGroup(group AudioGroup, volume float32, muted bool) error {
	return statusError(C.kui_document_audio_set_group(a.raw(), C.KuiCAudioGroup(group), C.float(volume), boolInt(muted)), "Audio.SetGroup")
}

// Group 返回一个分组的音量与静音状态。
func (a *Audio) Group(group AudioGroup) (volume float32, muted bool, err error) {
	var rawVolume C.float
	var rawMuted C.int
	err = statusError(C.kui_document_audio_get_group(a.raw(), C.KuiCAudioGroup(group), &rawVolume, &rawMuted), "Audio.Group")
	return float32(rawVolume), rawMuted != 0, err
}

// StopAll 停止一个分组（或 AudioAllGroups）的全部声音。
func (a *Audio) StopAll(group AudioGroup) error {
	return statusError(C.kui_document_audio_stop_all(a.raw(), C.int(group)), "Audio.StopAll")
}

// SetAllPaused 暂停或继续一个分组（或 AudioAllGroups）的全部声音。
func (a *Audio) SetAllPaused(group AudioGroup, paused bool) error {
	return statusError(C.kui_document_audio_set_all_paused(a.raw(), C.int(group), boolInt(paused)), "Audio.SetAllPaused")
}

// Preload 预加载 count 个实例，返回成功的个数。
func (a *Audio) Preload(path string, count uint32) (uint32, error) {
	raw, size, release := inputString(path)
	defer release()
	var loaded C.uint32_t
	err := statusError(C.kui_document_audio_preload(a.raw(), raw, size, C.uint32_t(count), &loaded), "Audio.Preload")
	return uint32(loaded), err
}

// Unload 卸载一个文件的全部实例。
func (a *Audio) Unload(path string) error {
	raw, size, release := inputString(path)
	defer release()
	return statusError(C.kui_document_audio_unload(a.raw(), raw, size), "Audio.Unload")
}

// UnloadAll 卸载全部。
func (a *Audio) UnloadAll() error {
	return statusError(C.kui_document_audio_unload(a.raw(), nil, 0), "Audio.UnloadAll")
}

// Counts 返回一个分组（或 AudioAllGroups）正在播放的数量，以及整个实例池的大小。
func (a *Audio) Counts(group AudioGroup) (active, pooled uint32, err error) {
	var rawActive, rawPooled C.uint32_t
	err = statusError(C.kui_document_audio_get_counts(a.raw(), C.int(group), &rawActive, &rawPooled), "Audio.Counts")
	return uint32(rawActive), uint32(rawPooled), err
}
