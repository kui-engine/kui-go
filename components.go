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

// 3.61.8 起的组件（C ABI 第 18 节）：内置自定义元素、主题、VirtualList、ListView、TreeView。
// 全部只能在 UI 线程调用，回调也在 UI 线程执行；组件句柄用完要 Close（会移除组件挂在容器上的监听，
// 已渲染的行留在容器里）。

// RegisterBuiltinComponents 注册 <kui-dialog>、<kui-tabs>、<kui-tooltip>、<kui-badge>、<kui-tree>、
// <kui-table>、<kui-menu> 等内置元素。
func (d *Document) RegisterBuiltinComponents() error {
	return statusError(C.kui_document_register_builtin_components(d.raw()), "Document.RegisterBuiltinComponents")
}

// ---------------------------------------------------------------------------
// 主题
// ---------------------------------------------------------------------------

// ThemeManager 管理具名的一组 CSS 自定义属性；内置 "light" 与 "dark"。
type ThemeManager struct{ handle unsafe.Pointer }

func NewThemeManager() (*ThemeManager, error) {
	var raw *C.KuiThemeManager
	if err := statusError(C.kui_theme_manager_create(&raw), "NewThemeManager"); err != nil {
		return nil, err
	}
	return &ThemeManager{handle: unsafe.Pointer(raw)}, nil
}
func (m *ThemeManager) raw() *C.KuiThemeManager {
	if m == nil {
		return nil
	}
	return (*C.KuiThemeManager)(m.handle)
}
func (m *ThemeManager) Close() error {
	if m == nil || m.handle == nil {
		return nil
	}
	err := statusError(C.kui_theme_manager_release(m.raw()), "ThemeManager.Close")
	m.handle = nil
	return err
}

// RegisterTheme 新增或替换一个主题；键是自定义属性名（如 "--bg"）。
func (m *ThemeManager) RegisterTheme(name string, variables map[string]string) error {
	rawName, size, release := inputString(name)
	defer release()
	pairs := make([]C.KuiCKeyValue, 0, len(variables))
	var strings []*C.char
	defer func() {
		for _, s := range strings {
			C.free(unsafe.Pointer(s))
		}
	}()
	for key, value := range variables {
		k, v := C.CString(key), C.CString(value)
		strings = append(strings, k, v)
		pairs = append(pairs, C.KuiCKeyValue{key: k, key_len: C.size_t(len(key)), value: v, value_len: C.size_t(len(value))})
	}
	var first *C.KuiCKeyValue
	if len(pairs) > 0 {
		first = &pairs[0]
	}
	return statusError(C.kui_theme_manager_register(m.raw(), rawName, size, first, C.size_t(len(pairs))), "ThemeManager.RegisterTheme")
}

// Apply 把主题变量设到 <html> 的内联样式上；未知主题返回 ErrNotFound。
func (m *ThemeManager) Apply(name string, document *Document) error {
	rawName, size, release := inputString(name)
	defer release()
	return statusError(C.kui_theme_manager_apply(m.raw(), rawName, size, document.raw()), "ThemeManager.Apply")
}

func (m *ThemeManager) readString(call func(*C.char, C.size_t, *C.size_t) C.KuiStatus, context string) (string, error) {
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

// ActiveTheme 是最近一次应用的主题；还没应用过为空。
func (m *ThemeManager) ActiveTheme() (string, error) {
	return m.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_theme_manager_get_active(m.raw(), buf, capacity, out)
	}, "ThemeManager.ActiveTheme")
}
func (m *ThemeManager) ThemeNames() ([]string, error) {
	var count C.size_t
	if err := statusError(C.kui_theme_manager_get_theme_count(m.raw(), &count), "ThemeManager.ThemeNames"); err != nil {
		return nil, err
	}
	names := make([]string, 0, int(count))
	for index := C.size_t(0); index < count; index++ {
		name, err := m.readString(func(buf *C.char, capacity C.size_t, out *C.size_t) C.KuiStatus {
			return C.kui_theme_manager_get_theme_name(m.raw(), index, buf, capacity, out)
		}, "ThemeManager.ThemeNames")
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

// Theme 返回主题的变量；未知主题返回 ErrNotFound。
func (m *ThemeManager) Theme(name string) (map[string]string, error) {
	rawName, size, release := inputString(name)
	defer release()
	var count C.size_t
	if err := statusError(C.kui_theme_manager_get_variable_count(m.raw(), rawName, size, &count), "ThemeManager.Theme"); err != nil {
		return nil, err
	}
	result := make(map[string]string, int(count))
	for index := C.size_t(0); index < count; index++ {
		var keyLen, valueLen C.size_t
		C.kui_theme_manager_get_variable(m.raw(), rawName, size, index, nil, 0, &keyLen, nil, 0, &valueLen)
		key := make([]byte, int(keyLen)+1)
		value := make([]byte, int(valueLen)+1)
		if err := statusError(C.kui_theme_manager_get_variable(m.raw(), rawName, size, index,
			(*C.char)(unsafe.Pointer(&key[0])), C.size_t(len(key)), &keyLen,
			(*C.char)(unsafe.Pointer(&value[0])), C.size_t(len(value)), &valueLen), "ThemeManager.Theme"); err != nil {
			return nil, err
		}
		result[string(key[:int(keyLen)])] = string(value[:int(valueLen)])
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// 共用
// ---------------------------------------------------------------------------

type ScrollUnit int32

const (
	ScrollPixel ScrollUnit = iota
	ScrollLine
)

// VirtualListState 是虚拟列表的当前状态：已渲染的行是 [VisibleStart, VisibleEnd)。
type VirtualListState struct {
	VisibleStart, VisibleEnd    uint32
	TotalItems, RecyclePoolSize uint32
	FocusedIndex                int32 // -1 表示没有
	TotalHeight, ItemHeight     float32
	ScrollUnit                  ScrollUnit
}

func virtualListState(raw C.KuiCVirtualListState) VirtualListState {
	return VirtualListState{
		VisibleStart: uint32(raw.visible_start), VisibleEnd: uint32(raw.visible_end),
		TotalItems: uint32(raw.total_items), RecyclePoolSize: uint32(raw.recycle_pool_size),
		FocusedIndex: int32(raw.focused_index), TotalHeight: float32(raw.total_height),
		ItemHeight: float32(raw.item_height), ScrollUnit: ScrollUnit(raw.scroll_unit),
	}
}

// componentState 是组件回调拿到的状态；字符串结果放在 C 内存里，下次回调或释放时 free。
type componentState struct {
	document *Document
	vlist    *VirtualListConfig
	list     *ListViewConfig
	tree     *TreeViewConfig
	cStrings []*C.char
}

func (s *componentState) freeStrings() {
	for _, value := range s.cStrings {
		C.free(unsafe.Pointer(value))
	}
	s.cStrings = s.cStrings[:0]
}
func (s *componentState) keep(value string) (*C.char, C.size_t) {
	raw := C.CString(value)
	s.cStrings = append(s.cStrings, raw)
	return raw, C.size_t(len(value))
}

func componentFromHandle(handle C.uintptr_t) (*componentState, bool) {
	state, ok := cgo.Handle(handle).Value().(*componentState)
	return state, ok
}

//export goKuiComponentRelease
func goKuiComponentRelease(handle C.uintptr_t) {
	defer containCallbackPanic()
	h := cgo.Handle(handle)
	if state, ok := h.Value().(*componentState); ok {
		state.freeStrings()
	}
	h.Delete()
}

//export goKuiComponentSelect
func goKuiComponentSelect(index C.uint32_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := componentFromHandle(handle)
	switch {
	case !ok:
	case state.list != nil && state.list.OnSelect != nil:
		state.list.OnSelect(uint32(index))
	case state.tree != nil && state.tree.OnSelect != nil:
		state.tree.OnSelect(uint32(index))
	}
}

//export goKuiComponentActivate
func goKuiComponentActivate(index C.uint32_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := componentFromHandle(handle)
	switch {
	case !ok:
	case state.vlist != nil && state.vlist.OnActivate != nil:
		state.vlist.OnActivate(uint32(index))
	case state.list != nil && state.list.OnActivate != nil:
		state.list.OnActivate(uint32(index))
	}
}

// ---------------------------------------------------------------------------
// VirtualList
// ---------------------------------------------------------------------------

// VirtualListConfig 配置虚拟列表。RenderItem 返回一行的新元素，KUI 接管它（之后不要再用或 Close）。
type VirtualListConfig struct {
	TotalItems     uint32
	ItemHeight     float32 // 0 为 32
	Overscan       uint32
	RecyclePoolMax uint32
	ScrollUnit     ScrollUnit
	EnableKeyboard bool
	RenderItem     func(document *Document, index uint32) *Element
	HeightForItem  func(index uint32) float32 // nil 时每行 ItemHeight
	OnFocusChange  func(oldIndex, newIndex int32)
	OnActivate     func(index uint32)
}

//export goKuiVirtualListRender
func goKuiVirtualListRender(index C.uint32_t, handle C.uintptr_t) (result *C.KuiElement) {
	defer containCallbackPanic()
	state, ok := componentFromHandle(handle)
	if !ok || state.vlist == nil || state.vlist.RenderItem == nil {
		return nil
	}
	element := state.vlist.RenderItem(state.document, uint32(index))
	if element == nil || element.handle == nil {
		return nil
	}
	raw := element.raw()
	if element.borrowed {
		if C.kui_element_retain(raw) != C.KUI_OK {
			return nil
		}
		return raw
	}
	element.handle = nil // KUI 接管这个引用
	return raw
}

//export goKuiVirtualListHeight
func goKuiVirtualListHeight(index C.uint32_t, handle C.uintptr_t) (result C.float) {
	defer containCallbackPanic()
	if state, ok := componentFromHandle(handle); ok && state.vlist != nil && state.vlist.HeightForItem != nil {
		result = C.float(state.vlist.HeightForItem(uint32(index)))
	}
	return result
}

//export goKuiVirtualListFocus
func goKuiVirtualListFocus(oldIndex, newIndex C.int32_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	if state, ok := componentFromHandle(handle); ok && state.vlist != nil && state.vlist.OnFocusChange != nil {
		state.vlist.OnFocusChange(int32(oldIndex), int32(newIndex))
	}
}

type VirtualList struct{ handle unsafe.Pointer }

func (v *VirtualList) raw() *C.KuiVirtualList {
	if v == nil {
		return nil
	}
	return (*C.KuiVirtualList)(v.handle)
}

// NewVirtualList 把 container 设为虚拟滚动容器（会改它的 position / overflow），返回前先渲染首屏。
func NewVirtualList(container *Element, config VirtualListConfig) (*VirtualList, error) {
	if config.RenderItem == nil {
		return nil, &Error{Status: ErrInvalidArg, Context: "NewVirtualList", Message: "RenderItem is required"}
	}
	stored := config
	handle, user := callbackHandle(&componentState{document: container.document, vlist: &stored})
	raw := C.KuiCVirtualListConfig{
		total_items: C.uint32_t(config.TotalItems), item_height: C.float(config.ItemHeight),
		overscan: C.uint32_t(config.Overscan), recycle_pool_max: C.uint32_t(config.RecyclePoolMax),
		scroll_unit: C.KuiCScrollUnit(config.ScrollUnit), enable_keyboard: boolInt(config.EnableKeyboard),
		render_item: C.kui_go_virtual_list_render_trampoline(), user_data: user,
		release_fn: C.kui_go_component_release_trampoline(),
	}
	if config.HeightForItem != nil {
		raw.height_for_item = C.kui_go_virtual_list_height_trampoline()
	}
	if config.OnFocusChange != nil {
		raw.on_focus_change = C.kui_go_virtual_list_focus_trampoline()
	}
	if config.OnActivate != nil {
		raw.on_activate = C.kui_go_component_activate_trampoline()
	}
	var list *C.KuiVirtualList
	if err := statusError(C.kui_virtual_list_create(container.raw(), &raw, &list), "NewVirtualList"); err != nil {
		handle.Delete()
		return nil, err
	}
	return &VirtualList{handle: unsafe.Pointer(list)}, nil
}
func (v *VirtualList) Close() error {
	if v == nil || v.handle == nil {
		return nil
	}
	err := statusError(C.kui_virtual_list_release(v.raw()), "VirtualList.Close")
	v.handle = nil
	return err
}
func (v *VirtualList) State() (VirtualListState, error) {
	var raw C.KuiCVirtualListState
	err := statusError(C.kui_virtual_list_get_state(v.raw(), &raw), "VirtualList.State")
	return virtualListState(raw), err
}
func (v *VirtualList) SetTotalItems(count uint32) error {
	return statusError(C.kui_virtual_list_set_total_items(v.raw(), C.uint32_t(count)), "VirtualList.SetTotalItems")
}
func (v *VirtualList) InsertItems(index, count uint32) error {
	return statusError(C.kui_virtual_list_insert_items(v.raw(), C.uint32_t(index), C.uint32_t(count)), "VirtualList.InsertItems")
}
func (v *VirtualList) RemoveItems(index, count uint32) error {
	return statusError(C.kui_virtual_list_remove_items(v.raw(), C.uint32_t(index), C.uint32_t(count)), "VirtualList.RemoveItems")
}

// Refresh 重新渲染可见的行（数据变了之后）。
func (v *VirtualList) Refresh() error {
	return statusError(C.kui_virtual_list_refresh(v.raw()), "VirtualList.Refresh")
}
func (v *VirtualList) ScrollToIndex(index uint32) error {
	return statusError(C.kui_virtual_list_scroll_to_index(v.raw(), C.uint32_t(index)), "VirtualList.ScrollToIndex")
}
func (v *VirtualList) ScrollToTop() error {
	return statusError(C.kui_virtual_list_scroll_to_top(v.raw()), "VirtualList.ScrollToTop")
}
func (v *VirtualList) ScrollToBottom() error {
	return statusError(C.kui_virtual_list_scroll_to_bottom(v.raw()), "VirtualList.ScrollToBottom")
}
func (v *VirtualList) SetScrollUnit(unit ScrollUnit) error {
	return statusError(C.kui_virtual_list_set_scroll_unit(v.raw(), C.KuiCScrollUnit(unit)), "VirtualList.SetScrollUnit")
}

// SetFocusedIndex 设置键盘焦点行（越界会被夹住，-1 清除），变化时调用 OnFocusChange。
func (v *VirtualList) SetFocusedIndex(index int32) error {
	return statusError(C.kui_virtual_list_set_focused_index(v.raw(), C.int32_t(index)), "VirtualList.SetFocusedIndex")
}
func (v *VirtualList) MoveFocus(delta int32) error {
	return statusError(C.kui_virtual_list_move_focus(v.raw(), C.int32_t(delta)), "VirtualList.MoveFocus")
}

// ---------------------------------------------------------------------------
// ListView
// ---------------------------------------------------------------------------

type SelectionMode int32

const (
	SelectionNone SelectionMode = iota
	SelectionSingle
	SelectionMulti
)

type ListViewItem struct {
	Text, Subtitle, Icon string
	Disabled             bool
}

// ListViewConfig 配置列表；颜色为空时用默认（深色主题）。
type ListViewConfig struct {
	ItemCount      uint32
	ItemHeight     float32 // 0 为 36
	Overscan       uint32
	ScrollUnit     ScrollUnit
	SelectionMode  SelectionMode
	ItemBg         string
	ItemHoverBg    string
	ItemSelectedBg string
	ItemTextColor  string
	ItemSubColor   string
	ItemProvider   func(index uint32) ListViewItem
	OnSelect       func(index uint32) // 点击 / 键盘选择
	OnActivate     func(index uint32) // 双击 / Enter
}

//export goKuiListViewItem
func goKuiListViewItem(index C.uint32_t, out *C.KuiCListViewItem, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := componentFromHandle(handle)
	if !ok || state.list == nil || state.list.ItemProvider == nil {
		return
	}
	item := state.list.ItemProvider(uint32(index))
	state.freeStrings() // 上一次的结果 KUI 已复制
	out.text, out.text_len = state.keep(item.Text)
	out.subtitle, out.subtitle_len = state.keep(item.Subtitle)
	out.icon, out.icon_len = state.keep(item.Icon)
	out.disabled = boolInt(item.Disabled)
}

type ListView struct{ handle unsafe.Pointer }

func (l *ListView) raw() *C.KuiListView {
	if l == nil {
		return nil
	}
	return (*C.KuiListView)(l.handle)
}

// colorArg 返回 C 字符串（空串为 NULL，用默认）；release 释放。
func colorArg(value string, keep *[]*C.char) *C.char {
	if value == "" {
		return nil
	}
	raw := C.CString(value)
	*keep = append(*keep, raw)
	return raw
}

func NewListView(container *Element, config ListViewConfig) (*ListView, error) {
	if config.ItemProvider == nil {
		return nil, &Error{Status: ErrInvalidArg, Context: "NewListView", Message: "ItemProvider is required"}
	}
	stored := config
	handle, user := callbackHandle(&componentState{document: container.document, list: &stored})
	var colors []*C.char
	defer func() {
		for _, value := range colors {
			C.free(unsafe.Pointer(value))
		}
	}()
	raw := C.KuiCListViewConfig{
		item_count: C.uint32_t(config.ItemCount), item_height: C.float(config.ItemHeight),
		overscan: C.uint32_t(config.Overscan), scroll_unit: C.KuiCScrollUnit(config.ScrollUnit),
		selection_mode: C.KuiCSelectionMode(config.SelectionMode),
		item_bg:        colorArg(config.ItemBg, &colors), item_hover_bg: colorArg(config.ItemHoverBg, &colors),
		item_selected_bg: colorArg(config.ItemSelectedBg, &colors), item_text_color: colorArg(config.ItemTextColor, &colors),
		item_sub_color: colorArg(config.ItemSubColor, &colors),
		item_provider:  C.kui_go_list_view_item_trampoline(), user_data: user,
		release_fn: C.kui_go_component_release_trampoline(),
	}
	if config.OnSelect != nil {
		raw.on_select = C.kui_go_component_select_trampoline()
	}
	if config.OnActivate != nil {
		raw.on_activate = C.kui_go_component_activate_trampoline()
	}
	var list *C.KuiListView
	if err := statusError(C.kui_list_view_create(container.raw(), &raw, &list), "NewListView"); err != nil {
		handle.Delete()
		return nil, err
	}
	return &ListView{handle: unsafe.Pointer(list)}, nil
}
func (l *ListView) Close() error {
	if l == nil || l.handle == nil {
		return nil
	}
	err := statusError(C.kui_list_view_release(l.raw()), "ListView.Close")
	l.handle = nil
	return err
}
func (l *ListView) Select(index uint32) error {
	return statusError(C.kui_list_view_select(l.raw(), C.uint32_t(index)), "ListView.Select")
}
func (l *ListView) Deselect(index uint32) error {
	return statusError(C.kui_list_view_deselect(l.raw(), C.uint32_t(index)), "ListView.Deselect")
}
func (l *ListView) ClearSelection() error {
	return statusError(C.kui_list_view_clear_selection(l.raw()), "ListView.ClearSelection")
}

// SelectRange 把 [start, end) 加进选择。
func (l *ListView) SelectRange(start, end uint32) error {
	return statusError(C.kui_list_view_select_range(l.raw(), C.uint32_t(start), C.uint32_t(end)), "ListView.SelectRange")
}
func (l *ListView) IsSelected(index uint32) (bool, error) {
	var value C.int
	err := statusError(C.kui_list_view_is_selected(l.raw(), C.uint32_t(index), &value), "ListView.IsSelected")
	return value != 0, err
}

func readIndices(call func(*C.uint32_t, C.size_t, *C.size_t) C.KuiStatus, context string) ([]uint32, error) {
	var count C.size_t
	if err := statusError(call(nil, 0, &count), context); err != nil {
		return nil, err
	}
	if count == 0 {
		return []uint32{}, nil
	}
	values := make([]uint32, int(count))
	if err := statusError(call((*C.uint32_t)(unsafe.Pointer(&values[0])), count, &count), context); err != nil {
		return nil, err
	}
	return values[:int(count)], nil
}

// Selected 是选中的行，升序。
func (l *ListView) Selected() ([]uint32, error) {
	return readIndices(func(buf *C.uint32_t, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_list_view_get_selected(l.raw(), buf, capacity, out)
	}, "ListView.Selected")
}
func (l *ListView) FocusedIndex() (int32, error) {
	var value C.int32_t
	err := statusError(C.kui_list_view_get_focused_index(l.raw(), &value), "ListView.FocusedIndex")
	return int32(value), err
}
func (l *ListView) SetItemCount(count uint32) error {
	return statusError(C.kui_list_view_set_item_count(l.raw(), C.uint32_t(count)), "ListView.SetItemCount")
}
func (l *ListView) Refresh() error {
	return statusError(C.kui_list_view_refresh(l.raw()), "ListView.Refresh")
}
func (l *ListView) ScrollToIndex(index uint32) error {
	return statusError(C.kui_list_view_scroll_to_index(l.raw(), C.uint32_t(index)), "ListView.ScrollToIndex")
}
func (l *ListView) ScrollToTop() error {
	return statusError(C.kui_list_view_scroll_to_top(l.raw()), "ListView.ScrollToTop")
}
func (l *ListView) ScrollToBottom() error {
	return statusError(C.kui_list_view_scroll_to_bottom(l.raw()), "ListView.ScrollToBottom")
}
func (l *ListView) SetScrollUnit(unit ScrollUnit) error {
	return statusError(C.kui_list_view_set_scroll_unit(l.raw(), C.KuiCScrollUnit(unit)), "ListView.SetScrollUnit")
}

// State 是列表底下 VirtualList 的状态。
func (l *ListView) State() (VirtualListState, error) {
	var raw C.KuiCVirtualListState
	err := statusError(C.kui_list_view_get_state(l.raw(), &raw), "ListView.State")
	return virtualListState(raw), err
}

// ---------------------------------------------------------------------------
// TreeView
// ---------------------------------------------------------------------------

// TreeViewConfig 配置树；颜色为空时用默认。
type TreeViewConfig struct {
	ItemHeight         float32 // 0 为 28
	IndentWidth        float32 // 0 为 20
	Overscan           uint32
	ScrollUnit         ScrollUnit
	ItemBg             string
	ItemHoverBg        string
	ItemSelectedBg     string
	ItemTextColor      string
	MultiSelect        bool // Ctrl+单击切换，Shift+单击选区间
	OnSelect           func(node uint32)
	OnToggle           func(node uint32, expanded bool)
	OnSelectionChanged func(selected []uint32) // 只在 MultiSelect 时
}

type TreeNode struct {
	Text, Icon         string
	Expanded, Disabled bool
	Parent             int32 // -1 为根
	Children           []uint32
}

//export goKuiTreeToggle
func goKuiTreeToggle(node C.uint32_t, expanded C.int, handle C.uintptr_t) {
	defer containCallbackPanic()
	if state, ok := componentFromHandle(handle); ok && state.tree != nil && state.tree.OnToggle != nil {
		state.tree.OnToggle(uint32(node), expanded != 0)
	}
}

//export goKuiTreeSelection
func goKuiTreeSelection(selected *C.uint32_t, count C.size_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := componentFromHandle(handle)
	if !ok || state.tree == nil || state.tree.OnSelectionChanged == nil {
		return
	}
	values := make([]uint32, 0, int(count))
	if count > 0 {
		for _, value := range unsafe.Slice(selected, int(count)) {
			values = append(values, uint32(value))
		}
	}
	state.tree.OnSelectionChanged(values)
}

type TreeView struct{ handle unsafe.Pointer }

func (t *TreeView) raw() *C.KuiTreeView {
	if t == nil {
		return nil
	}
	return (*C.KuiTreeView)(t.handle)
}

// NewTreeView 返回空树：先加节点再 Setup（Setup 之后加的节点在 Refresh 后出现）。
func NewTreeView() (*TreeView, error) {
	var raw *C.KuiTreeView
	if err := statusError(C.kui_tree_view_create(&raw), "NewTreeView"); err != nil {
		return nil, err
	}
	return &TreeView{handle: unsafe.Pointer(raw)}, nil
}
func (t *TreeView) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	err := statusError(C.kui_tree_view_release(t.raw()), "TreeView.Close")
	t.handle = nil
	return err
}
func (t *TreeView) AddNode(text, icon string) (uint32, error) {
	rawText, textLen, releaseText := inputString(text)
	defer releaseText()
	rawIcon, iconLen, releaseIcon := inputString(icon)
	defer releaseIcon()
	var index C.uint32_t
	err := statusError(C.kui_tree_view_add_node(t.raw(), rawText, textLen, rawIcon, iconLen, &index), "TreeView.AddNode")
	return uint32(index), err
}
func (t *TreeView) AddChild(parent uint32, text, icon string) (uint32, error) {
	rawText, textLen, releaseText := inputString(text)
	defer releaseText()
	rawIcon, iconLen, releaseIcon := inputString(icon)
	defer releaseIcon()
	var index C.uint32_t
	err := statusError(C.kui_tree_view_add_child(t.raw(), C.uint32_t(parent), rawText, textLen, rawIcon, iconLen, &index), "TreeView.AddChild")
	return uint32(index), err
}
func (t *TreeView) NodeCount() (uint32, error) {
	var count C.uint32_t
	err := statusError(C.kui_tree_view_get_node_count(t.raw(), &count), "TreeView.NodeCount")
	return uint32(count), err
}
func (t *TreeView) Node(index uint32) (TreeNode, error) {
	var raw C.KuiCTreeNode
	if err := statusError(C.kui_tree_view_get_node(t.raw(), C.uint32_t(index), &raw), "TreeView.Node"); err != nil {
		return TreeNode{}, err
	}
	node := TreeNode{
		Text: goStringN(raw.text, raw.text_len), Icon: goStringN(raw.icon, raw.icon_len),
		Expanded: raw.expanded != 0, Disabled: raw.disabled != 0, Parent: int32(raw.parent),
		Children: make([]uint32, 0, int(raw.child_count)),
	}
	for position := C.size_t(0); position < raw.child_count; position++ {
		var child C.uint32_t
		if err := statusError(C.kui_tree_view_get_node_child(t.raw(), C.uint32_t(index), position, &child), "TreeView.Node"); err != nil {
			return TreeNode{}, err
		}
		node.Children = append(node.Children, uint32(child))
	}
	return node, nil
}

// TreeNodeUpdate 是 UpdateNode 的改动：为 nil 的字段保持原值。
type TreeNodeUpdate struct {
	Text, Icon *string
	Disabled   *bool
}

// UpdateNode 只改 update 里设了的字段（nil 保持原值）；Refresh 后显示。
func (t *TreeView) UpdateNode(index uint32, update TreeNodeUpdate) error {
	var rawText, rawIcon *C.char
	var textLen, iconLen C.size_t
	if update.Text != nil {
		var release func()
		rawText, textLen, release = inputString(*update.Text)
		defer release()
	}
	if update.Icon != nil {
		var release func()
		rawIcon, iconLen, release = inputString(*update.Icon)
		defer release()
	}
	var disabled C.int
	if update.Disabled != nil {
		disabled = boolInt(*update.Disabled)
	} else {
		current, err := t.Node(index)
		if err != nil {
			return err
		}
		disabled = boolInt(current.Disabled)
	}
	return statusError(C.kui_tree_view_set_node(t.raw(), C.uint32_t(index), rawText, textLen, rawIcon, iconLen, disabled), "TreeView.UpdateNode")
}

// SetNode 改节点的文字、图标与禁用状态；Refresh 后显示。要保持其中某项原值用 UpdateNode。
func (t *TreeView) SetNode(index uint32, text, icon string, disabled bool) error {
	rawText, textLen, releaseText := inputString(text)
	defer releaseText()
	rawIcon, iconLen, releaseIcon := inputString(icon)
	defer releaseIcon()
	return statusError(C.kui_tree_view_set_node(t.raw(), C.uint32_t(index), rawText, textLen, rawIcon, iconLen, boolInt(disabled)), "TreeView.SetNode")
}
func (t *TreeView) Setup(container *Element, config TreeViewConfig) error {
	stored := config
	handle, user := callbackHandle(&componentState{document: container.document, tree: &stored})
	var colors []*C.char
	defer func() {
		for _, value := range colors {
			C.free(unsafe.Pointer(value))
		}
	}()
	raw := C.KuiCTreeViewConfig{
		item_height: C.float(config.ItemHeight), indent_width: C.float(config.IndentWidth),
		overscan: C.uint32_t(config.Overscan), scroll_unit: C.KuiCScrollUnit(config.ScrollUnit),
		item_bg: colorArg(config.ItemBg, &colors), item_hover_bg: colorArg(config.ItemHoverBg, &colors),
		item_selected_bg: colorArg(config.ItemSelectedBg, &colors), item_text_color: colorArg(config.ItemTextColor, &colors),
		multi_select: boolInt(config.MultiSelect), user_data: user,
		release_fn: C.kui_go_component_release_trampoline(),
	}
	if config.OnSelect != nil {
		raw.on_select = C.kui_go_component_select_trampoline()
	}
	if config.OnToggle != nil {
		raw.on_toggle = C.kui_go_tree_toggle_trampoline()
	}
	if config.OnSelectionChanged != nil {
		raw.on_selection_changed = C.kui_go_tree_selection_trampoline()
	}
	if err := statusError(C.kui_tree_view_setup(t.raw(), container.raw(), &raw), "TreeView.Setup"); err != nil {
		handle.Delete()
		return err
	}
	return nil
}
func (t *TreeView) Expand(index uint32) error {
	return statusError(C.kui_tree_view_expand(t.raw(), C.uint32_t(index)), "TreeView.Expand")
}
func (t *TreeView) Collapse(index uint32) error {
	return statusError(C.kui_tree_view_collapse(t.raw(), C.uint32_t(index)), "TreeView.Collapse")
}
func (t *TreeView) Toggle(index uint32) error {
	return statusError(C.kui_tree_view_toggle(t.raw(), C.uint32_t(index)), "TreeView.Toggle")
}
func (t *TreeView) ExpandAll() error {
	return statusError(C.kui_tree_view_expand_all(t.raw()), "TreeView.ExpandAll")
}
func (t *TreeView) CollapseAll() error {
	return statusError(C.kui_tree_view_collapse_all(t.raw()), "TreeView.CollapseAll")
}
func (t *TreeView) Select(index uint32) error {
	return statusError(C.kui_tree_view_select(t.raw(), C.uint32_t(index)), "TreeView.Select")
}
func (t *TreeView) ClearSelection() error {
	return statusError(C.kui_tree_view_clear_selection(t.raw()), "TreeView.ClearSelection")
}
func (t *TreeView) ToggleSelect(index uint32) error {
	return statusError(C.kui_tree_view_toggle_select(t.raw(), C.uint32_t(index)), "TreeView.ToggleSelect")
}
func (t *TreeView) SelectRange(index uint32) error {
	return statusError(C.kui_tree_view_select_range(t.raw(), C.uint32_t(index)), "TreeView.SelectRange")
}
func (t *TreeView) SelectAll() error {
	return statusError(C.kui_tree_view_select_all(t.raw()), "TreeView.SelectAll")
}
func (t *TreeView) IsSelected(index uint32) (bool, error) {
	var value C.int
	err := statusError(C.kui_tree_view_is_selected(t.raw(), C.uint32_t(index), &value), "TreeView.IsSelected")
	return value != 0, err
}

// SelectedIndex 是最近选中的节点；没有为 -1。
func (t *TreeView) SelectedIndex() (int32, error) {
	var value C.int32_t
	err := statusError(C.kui_tree_view_get_selected_index(t.raw(), &value), "TreeView.SelectedIndex")
	return int32(value), err
}

// Selected 与 IsSelected 口径相同：多选树是选中的集合，单选树是那一个节点；升序。
func (t *TreeView) Selected() ([]uint32, error) {
	return readIndices(func(buf *C.uint32_t, capacity C.size_t, out *C.size_t) C.KuiStatus {
		return C.kui_tree_view_get_selected(t.raw(), buf, capacity, out)
	}, "TreeView.Selected")
}
func (t *TreeView) Refresh() error {
	return statusError(C.kui_tree_view_refresh(t.raw()), "TreeView.Refresh")
}
func (t *TreeView) ScrollToNode(index uint32) error {
	return statusError(C.kui_tree_view_scroll_to_node(t.raw(), C.uint32_t(index)), "TreeView.ScrollToNode")
}
func (t *TreeView) SetScrollUnit(unit ScrollUnit) error {
	return statusError(C.kui_tree_view_set_scroll_unit(t.raw(), C.KuiCScrollUnit(unit)), "TreeView.SetScrollUnit")
}
func (t *TreeView) VisibleRowCount() (uint32, error) {
	var count C.uint32_t
	err := statusError(C.kui_tree_view_get_visible_row_count(t.raw(), &count), "TreeView.VisibleRowCount")
	return uint32(count), err
}

// ---------------------------------------------------------------------------
// 文档样式配置（C ABI 第 19 节）
// ---------------------------------------------------------------------------

// StyleProfile 是文档级的呈现策略：StyleNativeGUI（默认）保持 KUI 的 GUI 默认值，
// StyleBrowserCompat 用浏览器的文档默认值（默认字体为 serif）；作者样式照常覆盖。
type StyleProfile int32

const (
	StyleNativeGUI StyleProfile = iota
	StyleBrowserCompat
)

// SetStyleProfile 设置样式配置；在加载页面之前设，改动会重载 UA 样式表。
func (d *Document) SetStyleProfile(profile StyleProfile) error {
	return statusError(C.kui_document_set_style_profile(d.raw(), C.KuiCStyleProfile(profile)), "Document.SetStyleProfile")
}
func (d *Document) StyleProfile() (StyleProfile, error) {
	var value C.KuiCStyleProfile
	err := statusError(C.kui_document_get_style_profile(d.raw(), &value), "Document.StyleProfile")
	return StyleProfile(value), err
}
