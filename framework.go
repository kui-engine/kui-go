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

// 3.61.8 起的应用框架（C ABI 第 11 节）：路由、数据绑定、动画状态、Mutation / Resize /
// Intersection Observer。全部只能在 UI 线程调用，回调也在 UI 线程执行。回调句柄由 KUI 在
// 不再需要时通知释放（移除、替换、文档销毁或最后一个引用释放），不必手动管理。

//export goKuiReleaseCallback
func goKuiReleaseCallback(handle C.uintptr_t) {
	defer containCallbackPanic()
	cgo.Handle(handle).Delete()
}

// callbackHandle 包装一个回调；C 调用失败时由调用方 Delete（KUI 不会再调用 release）。
func callbackHandle(value any) (cgo.Handle, unsafe.Pointer) {
	h := cgo.NewHandle(value)
	return h, C.kui_go_handle_pointer(C.uintptr_t(h))
}

func goStringN(raw *C.char, size C.size_t) string {
	if raw == nil || size == 0 {
		return ""
	}
	return C.GoStringN(raw, C.int(size))
}

// borrowedElement 包装回调期间借来的元素；回调返回后不要再用它（需要保留就 Clone）。
func borrowedElement(raw *C.KuiElement, document *Document) *Element {
	if raw == nil {
		return nil
	}
	return &Element{handle: unsafe.Pointer(raw), document: document, borrowed: true}
}

// ---------------------------------------------------------------------------
// 路由
// ---------------------------------------------------------------------------

// RouteInfo 是一次导航的目标或当前路由。
type RouteInfo struct {
	Path        string            // 完整路径（含查询串），例如 "/users/42?tab=1"
	Pattern     string            // 匹配到的模式；没有匹配时为空
	Query       string            // '?' 之后的部分
	Params      map[string]string // ":name" 段
	QueryParams map[string]string
}

func routeInfo(raw *C.KuiCRouteInfo) RouteInfo {
	info := RouteInfo{Params: map[string]string{}, QueryParams: map[string]string{}}
	if raw == nil {
		return info
	}
	info.Path = goStringN(raw.path, raw.path_len)
	info.Pattern = goStringN(raw.pattern, raw.pattern_len)
	info.Query = goStringN(raw.query, raw.query_len)
	fill := func(items *C.KuiCKeyValue, count C.size_t, out map[string]string) {
		if items == nil || count == 0 {
			return
		}
		for _, item := range unsafe.Slice(items, int(count)) {
			out[goStringN(item.key, item.key_len)] = goStringN(item.value, item.value_len)
		}
	}
	fill(raw.params, raw.param_count, info.Params)
	fill(raw.query_params, raw.query_param_count, info.QueryParams)
	return info
}

// RouteViewFunc 为当前路由生成视图。返回的元素交给 KUI（之后不要再 Close 它）；返回 nil 表示空视图。
type RouteViewFunc func(doc *Document, route RouteInfo) *Element

// RouteGuardFunc 返回 false 取消这次导航。不要在守卫里增删守卫。
type RouteGuardFunc func(from, to RouteInfo) bool

// RouteGuardKind 区分进入前与离开前守卫。
type RouteGuardKind int

const (
	RouteBeforeEnter RouteGuardKind = iota
	RouteBeforeLeave
)

type routeViewState struct {
	fn       RouteViewFunc
	document *Document
}

//export goKuiRouteViewCallback
func goKuiRouteViewCallback(doc *C.KuiDocument, route *C.KuiCRouteInfo, handle C.uintptr_t) (result *C.KuiElement) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*routeViewState)
	if !ok || state.fn == nil {
		return nil
	}
	element := state.fn(state.document, routeInfo(route))
	if element == nil || element.handle == nil {
		return nil
	}
	raw := element.raw()
	if element.borrowed {
		// 借来的元素没有可以转交的引用，补一个。
		if C.kui_element_retain(raw) != C.KUI_OK {
			return nil
		}
	} else {
		element.handle = nil // 引用已转交 KUI
	}
	return raw
}

//export goKuiRouteGuardCallback
func goKuiRouteGuardCallback(from, to *C.KuiCRouteInfo, handle C.uintptr_t) (allow C.int) {
	defer containCallbackPanic() // 守卫 panic 视为取消
	fn, ok := cgo.Handle(handle).Value().(RouteGuardFunc)
	if !ok || fn == nil {
		return 1
	}
	if fn(routeInfo(from), routeInfo(to)) {
		return 1
	}
	return 0
}

//export goKuiRouteInfoCallback
func goKuiRouteInfoCallback(route *C.KuiCRouteInfo, handle C.uintptr_t) {
	defer containCallbackPanic()
	if fn, ok := cgo.Handle(handle).Value().(func(RouteInfo)); ok && fn != nil {
		fn(routeInfo(route))
	}
}

// Router 是文档的路由器（每个文档一个）。
type Router struct{ document *Document }

// Router 返回文档的路由器。
func (d *Document) Router() *Router { return &Router{document: d} }

func (r *Router) raw() *C.KuiDocument { return r.document.raw() }

// AddRoute 登记一条路由。模式里用 ":name" 表示参数，例如 "/users/:id"。
func (r *Router) AddRoute(pattern, name string, view RouteViewFunc) error {
	if view == nil {
		return &Error{Status: ErrInvalidArg, Context: "Router.AddRoute", Message: "nil view"}
	}
	rawPattern, patternLen, releasePattern := inputString(pattern)
	defer releasePattern()
	rawName, nameLen, releaseName := inputString(name)
	defer releaseName()
	h, user := callbackHandle(&routeViewState{fn: view, document: r.document})
	status := C.kui_document_router_add_route(r.raw(), rawPattern, patternLen, rawName, nameLen,
		C.kui_go_route_view_trampoline(), user, C.kui_go_release_trampoline())
	if err := statusError(status, "Router.AddRoute"); err != nil {
		h.Delete()
		return err
	}
	return nil
}

// RemoveRoute 移除一条路由；没有这条时返回 ErrNotFound。
func (r *Router) RemoveRoute(pattern string) error {
	raw, size, release := inputString(pattern)
	defer release()
	return statusError(C.kui_document_router_remove_route(r.raw(), raw, size), "Router.RemoveRoute")
}

// Clear 移除全部路由。
func (r *Router) Clear() error {
	return statusError(C.kui_document_router_clear_routes(r.raw()), "Router.Clear")
}

// HasRoute 报告是否登记过这个模式。
func (r *Router) HasRoute(pattern string) (bool, error) {
	raw, size, release := inputString(pattern)
	defer release()
	var value C.int
	err := statusError(C.kui_document_router_has_route(r.raw(), raw, size, &value), "Router.HasRoute")
	return value != 0, err
}

// RouteCount 返回登记的路由数。
func (r *Router) RouteCount() (int, error) {
	var value C.uint32_t
	err := statusError(C.kui_document_router_get_route_count(r.raw(), &value), "Router.RouteCount")
	return int(value), err
}

// RegisterView 让 view 显示当前路由的视图（每次导航替换它的子节点）。页面里的 <kui-router-view>
// 会自动登记，其他容器用本方法；view 必须属于本文档，换页后要为新页面的容器重新登记。
func (r *Router) RegisterView(view *Element) error {
	return statusError(C.kui_document_router_register_view(r.raw(), view.raw()), "Router.RegisterView")
}

// UnregisterView 取消登记；没登记过时返回 ErrNotFound。
func (r *Router) UnregisterView(view *Element) error {
	return statusError(C.kui_document_router_unregister_view(r.raw(), view.raw()), "Router.UnregisterView")
}

func (r *Router) navigate(path string, replace bool, context string) (bool, error) {
	raw, size, release := inputString(path)
	defer release()
	var navigated C.int
	var status C.KuiStatus
	if replace {
		status = C.kui_document_router_replace(r.raw(), raw, size, &navigated)
	} else {
		status = C.kui_document_router_push(r.raw(), raw, size, &navigated)
	}
	return navigated != 0, statusError(status, context)
}

// Push 导航到 path 并压入历史；守卫取消时返回 false。没有路由匹配也会导航（视图清空）。
func (r *Router) Push(path string) (bool, error) { return r.navigate(path, false, "Router.Push") }

// Replace 导航到 path 并替换当前历史项。
func (r *Router) Replace(path string) (bool, error) { return r.navigate(path, true, "Router.Replace") }

// Go 在历史里移动：-1 后退、1 前进；越界或被守卫取消时返回 false。
func (r *Router) Go(delta int) (bool, error) {
	var navigated C.int
	err := statusError(C.kui_document_router_go(r.raw(), C.int(delta), &navigated), "Router.Go")
	return navigated != 0, err
}

// Back 等于 Go(-1)。
func (r *Router) Back() (bool, error) { return r.Go(-1) }

// Forward 等于 Go(1)。
func (r *Router) Forward() (bool, error) { return r.Go(1) }

// Current 返回当前路由。
func (r *Router) Current() (RouteInfo, error) {
	var info RouteInfo
	h, user := callbackHandle(func(route RouteInfo) { info = route })
	defer h.Delete()
	err := statusError(C.kui_document_router_get_current(r.raw(), C.kui_go_route_info_trampoline(), user), "Router.Current")
	return info, err
}

// CurrentPath 返回当前路径；第一次导航前为空串。
func (r *Router) CurrentPath() (string, error) {
	var length C.size_t
	status := C.kui_document_router_get_current_path(r.raw(), nil, 0, &length)
	if status == C.KUI_ERR_NOT_FOUND {
		return "", nil
	}
	if err := statusError(status, "Router.CurrentPath"); err != nil {
		return "", err
	}
	buffer := make([]byte, int(length)+1)
	err := statusError(C.kui_document_router_get_current_path(r.raw(),
		(*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &length), "Router.CurrentPath")
	return string(buffer[:int(length)]), err
}

// Param 返回当前路由的 ":name" 参数；没有时 ok 为 false。
func (r *Router) Param(name string) (value string, ok bool, err error) {
	raw, size, release := inputString(name)
	defer release()
	var length C.size_t
	status := C.kui_document_router_get_param(r.raw(), raw, size, nil, 0, &length)
	if status == C.KUI_ERR_NOT_FOUND {
		return "", false, nil
	}
	if err := statusError(status, "Router.Param"); err != nil {
		return "", false, err
	}
	buffer := make([]byte, int(length)+1)
	err = statusError(C.kui_document_router_get_param(r.raw(), raw, size,
		(*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &length), "Router.Param")
	return string(buffer[:int(length)]), err == nil, err
}

// History 返回当前历史位置与历史长度；第一次导航前 index 为 -1。
func (r *Router) History() (index, size int, err error) {
	var rawIndex, rawSize C.int
	err = statusError(C.kui_document_router_get_history(r.raw(), &rawIndex, &rawSize), "Router.History")
	return int(rawIndex), int(rawSize), err
}

// AddGuard 登记全局守卫，返回用于移除的 id。
func (r *Router) AddGuard(kind RouteGuardKind, fn RouteGuardFunc) (uint32, error) {
	if fn == nil {
		return 0, &Error{Status: ErrInvalidArg, Context: "Router.AddGuard", Message: "nil guard"}
	}
	h, user := callbackHandle(fn)
	var id C.uint32_t
	status := C.kui_document_router_add_guard(r.raw(), C.KuiCRouteGuardKind(kind),
		C.kui_go_route_guard_trampoline(), user, C.kui_go_release_trampoline(), &id)
	if err := statusError(status, "Router.AddGuard"); err != nil {
		h.Delete()
		return 0, err
	}
	return uint32(id), nil
}

// RemoveGuard 移除守卫（未知 id 什么也不做）。
func (r *Router) RemoveGuard(id uint32) error {
	return statusError(C.kui_document_router_remove_guard(r.raw(), C.uint32_t(id)), "Router.RemoveGuard")
}

// OnNavigated 在每次导航成功、视图更新之后调用 fn，返回用于移除的 id。
func (r *Router) OnNavigated(fn func(RouteInfo)) (uint32, error) {
	if fn == nil {
		return 0, &Error{Status: ErrInvalidArg, Context: "Router.OnNavigated", Message: "nil callback"}
	}
	h, user := callbackHandle(fn)
	var id C.uint32_t
	status := C.kui_document_router_add_navigated_callback(r.raw(), C.kui_go_route_info_trampoline(), user,
		C.kui_go_release_trampoline(), &id)
	if err := statusError(status, "Router.OnNavigated"); err != nil {
		h.Delete()
		return 0, err
	}
	return uint32(id), nil
}

// RemoveNavigated 移除 OnNavigated 登记的回调（未知 id 什么也不做）。
func (r *Router) RemoveNavigated(id uint32) error {
	return statusError(C.kui_document_router_remove_navigated_callback(r.raw(), C.uint32_t(id)), "Router.RemoveNavigated")
}

// ---------------------------------------------------------------------------
// 数据绑定
// ---------------------------------------------------------------------------

// Observable 保存一个字符串值，可绑定到元素属性（"textContent"、"innerHTML"、"className"、
// "hidden"、"disabled"、"value"、"checked"、"valueAsNumber"，其他名字按属性处理）。
// 绑定时立即写入；之后的变化在下一帧（或 Document.FlushBindings）统一写入。
// 用完必须 Close（在 UI 线程）；它会让绑定的元素和文档一直存活到解绑。
type Observable struct{ handle unsafe.Pointer }

// NewObservable 创建初值为 initial 的 Observable。
func NewObservable(initial string) (*Observable, error) {
	raw, size, release := inputString(initial)
	defer release()
	var out *C.KuiObservable
	if err := statusError(C.kui_observable_create(raw, size, &out), "NewObservable"); err != nil {
		return nil, err
	}
	return &Observable{handle: unsafe.Pointer(out)}, nil
}

func (o *Observable) raw() *C.KuiObservable {
	if o == nil {
		return nil
	}
	return (*C.KuiObservable)(o.handle)
}

// Clone 再取一个引用（kui_observable_retain）：每个句柄各自 Close，最后一个关闭时才解除全部绑定与订阅。
func (o *Observable) Clone() (*Observable, error) {
	if o == nil || o.handle == nil {
		return nil, &Error{Status: ErrInvalidArg, Context: "Observable.Clone", Message: "observable is closed"}
	}
	if err := statusError(C.kui_observable_retain(o.raw()), "Observable.Clone"); err != nil {
		return nil, err
	}
	return &Observable{handle: o.handle}, nil
}

// Close 释放引用：解除全部绑定与订阅。
func (o *Observable) Close() error {
	if o == nil || o.handle == nil {
		return nil
	}
	err := statusError(C.kui_observable_release(o.raw()), "Observable.Close")
	o.handle = nil
	return err
}

// Set 改值；值变了才同步通知订阅者。
func (o *Observable) Set(value string) error {
	raw, size, release := inputString(value)
	defer release()
	return statusError(C.kui_observable_set(o.raw(), raw, size), "Observable.Set")
}

// Get 返回当前值。
func (o *Observable) Get() (string, error) {
	var size C.size_t
	if err := bufferStatus(C.kui_observable_get(o.raw(), nil, 0, &size), "Observable.Get"); err != nil {
		return "", err
	}
	buffer := make([]byte, int(size)+1)
	err := statusError(C.kui_observable_get(o.raw(), (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "Observable.Get")
	return string(buffer[:int(size)]), err
}

// Subscribe 在每次变化后调用 fn，返回用于退订的 id。
func (o *Observable) Subscribe(fn func()) (uint32, error) {
	if fn == nil {
		return 0, &Error{Status: ErrInvalidArg, Context: "Observable.Subscribe", Message: "nil callback"}
	}
	h, user := callbackHandle(&timerState{fn: fn})
	var id C.uint32_t
	status := C.kui_observable_subscribe(o.raw(), C.kui_go_task_trampoline(), user, C.kui_go_release_trampoline(), &id)
	if err := statusError(status, "Observable.Subscribe"); err != nil {
		h.Delete()
		return 0, err
	}
	return uint32(id), nil
}

// Unsubscribe 退订；未知 id 返回 ErrNotFound。
func (o *Observable) Unsubscribe(id uint32) error {
	return statusError(C.kui_observable_unsubscribe(o.raw(), C.uint32_t(id)), "Observable.Unsubscribe")
}

// Bind 把值绑定到元素属性；twoWay 时用户输入也会写回（"value" / "checked"）。返回绑定 id。
func (o *Observable) Bind(element *Element, property string, twoWay bool) (uint32, error) {
	raw, size, release := inputString(property)
	defer release()
	var id C.uint32_t
	err := statusError(C.kui_observable_bind(o.raw(), element.raw(), raw, size, boolInt(twoWay), &id), "Observable.Bind")
	return uint32(id), err
}

// Unbind 解除一个绑定；未知 id 返回 ErrNotFound。
func (o *Observable) Unbind(id uint32) error {
	return statusError(C.kui_observable_unbind(o.raw(), C.uint32_t(id)), "Observable.Unbind")
}

// UnbindAll 解除全部绑定。
func (o *Observable) UnbindAll() error {
	return statusError(C.kui_observable_unbind_all(o.raw()), "Observable.UnbindAll")
}

// FlushBindings 立即写入待提交的绑定更新，不等下一帧。
func (d *Document) FlushBindings() error {
	return statusError(C.kui_document_flush_bindings(d.raw()), "Document.FlushBindings")
}

// ---------------------------------------------------------------------------
// 动画状态
// ---------------------------------------------------------------------------

// AnimationState 报告元素上是否有正在运行的 CSS 动画 / 过渡。控制动画请改 CSS（class、内联样式、
// animation-play-state）；animationstart / animationend / transitionend 等事件用 AddEventListener 监听。
func (e *Element) AnimationState() (animating, transitioning bool, err error) {
	var a, t C.int
	err = statusError(C.kui_element_get_animation_state(e.raw(), &a, &t), "Element.AnimationState")
	return a != 0, t != 0, err
}

// AnimationCounts 返回文档里正在运行的 CSS 动画与过渡数。
func (d *Document) AnimationCounts() (animations, transitions uint32, err error) {
	var a, t C.uint32_t
	err = statusError(C.kui_document_get_animation_counts(d.raw(), &a, &t), "Document.AnimationCounts")
	return uint32(a), uint32(t), err
}

// ---------------------------------------------------------------------------
// Observer
// ---------------------------------------------------------------------------

// MutationType 是 DOM 变化的种类。
type MutationType int

const (
	MutationChildList MutationType = iota
	MutationAttributes
	MutationCharacterData
)

// MutationRecord 描述一次 DOM 变化。元素只在回调期间有效（需要保留就 Clone）；
// *NodeType 是 DOM nodeType（0 无、1 元素、3 文本、8 注释），文本节点对应的元素字段为 nil。
// CharacterData 的 Target 是文本节点的父元素。
type MutationRecord struct {
	Type            MutationType
	Target          *Element
	TargetNodeType  int
	Added           *Element
	AddedNodeType   int
	Removed         *Element
	RemovedNodeType int
	PreviousSibling *Element
	NextSibling     *Element
	AttributeName   string
	OldValue        string
}

// MutationOptions 选择要观察的变化；ChildList / Attributes / CharacterData 至少一个。
type MutationOptions struct {
	ChildList             bool
	Attributes            bool
	CharacterData         bool
	Subtree               bool
	AttributeOldValue     bool
	CharacterDataOldValue bool
	AttributeFilter       []string // 空 = 全部属性
}

// ResizeEntry 是一次尺寸变化（CSS px）。ContentX / ContentY 是内容盒在边框盒里的偏移。
type ResizeEntry struct {
	Target                      *Element
	ContentWidth, ContentHeight float32
	BorderWidth, BorderHeight   float32
	ContentX, ContentY          float32
}

// IntersectionOptions 配置 IntersectionObserver；Root 为 nil 表示视口，Thresholds 为空等于 {0}。
type IntersectionOptions struct {
	Root                                             *Element
	MarginTop, MarginRight, MarginBottom, MarginLeft float32
	Thresholds                                       []float32
}

// IntersectionEntry 是一次相交状态变化。
type IntersectionEntry struct {
	Target             *Element
	BoundingClientRect Rect
	IntersectionRect   Rect
	RootBounds         Rect
	Ratio              float32
	Intersecting       bool
}

type observerState struct {
	mutation     func([]MutationRecord)
	resize       func([]ResizeEntry)
	intersection func([]IntersectionEntry)
}

func toRect(raw C.KuiRect) Rect {
	return Rect{X: float32(raw.x), Y: float32(raw.y), Width: float32(raw.width), Height: float32(raw.height)}
}

//export goKuiMutationCallback
func goKuiMutationCallback(records *C.KuiCMutationRecord, count C.size_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*observerState)
	if !ok || state.mutation == nil || records == nil {
		return
	}
	out := make([]MutationRecord, 0, int(count))
	for _, r := range unsafe.Slice(records, int(count)) {
		out = append(out, MutationRecord{
			Type:            MutationType(r._type),
			Target:          borrowedElement(r.target, nil),
			TargetNodeType:  int(r.target_node_type),
			Added:           borrowedElement(r.added, nil),
			AddedNodeType:   int(r.added_node_type),
			Removed:         borrowedElement(r.removed, nil),
			RemovedNodeType: int(r.removed_node_type),
			PreviousSibling: borrowedElement(r.previous_sibling, nil),
			NextSibling:     borrowedElement(r.next_sibling, nil),
			AttributeName:   goStringN(r.attribute_name, r.attribute_name_len),
			OldValue:        goStringN(r.old_value, r.old_value_len),
		})
	}
	state.mutation(out)
}

//export goKuiResizeObserverCallback
func goKuiResizeObserverCallback(entries *C.KuiCResizeEntry, count C.size_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*observerState)
	if !ok || state.resize == nil || entries == nil {
		return
	}
	out := make([]ResizeEntry, 0, int(count))
	for _, e := range unsafe.Slice(entries, int(count)) {
		out = append(out, ResizeEntry{
			Target:        borrowedElement(e.target, nil),
			ContentWidth:  float32(e.content_width),
			ContentHeight: float32(e.content_height),
			BorderWidth:   float32(e.border_width),
			BorderHeight:  float32(e.border_height),
			ContentX:      float32(e.content_x),
			ContentY:      float32(e.content_y),
		})
	}
	state.resize(out)
}

//export goKuiIntersectionCallback
func goKuiIntersectionCallback(entries *C.KuiCIntersectionEntry, count C.size_t, handle C.uintptr_t) {
	defer containCallbackPanic()
	state, ok := cgo.Handle(handle).Value().(*observerState)
	if !ok || state.intersection == nil || entries == nil {
		return
	}
	out := make([]IntersectionEntry, 0, int(count))
	for _, e := range unsafe.Slice(entries, int(count)) {
		out = append(out, IntersectionEntry{
			Target:             borrowedElement(e.target, nil),
			BoundingClientRect: toRect(e.bounding_client_rect),
			IntersectionRect:   toRect(e.intersection_rect),
			RootBounds:         toRect(e.root_bounds),
			Ratio:              float32(e.intersection_ratio),
			Intersecting:       e.is_intersecting != 0,
		})
	}
	state.intersection(out)
}

// Observer 是 Mutation / Resize / Intersection Observer。被观察的元素在观察期间保持存活；
// 用完必须 Close（在 UI 线程，回调里也可以）。
type Observer struct{ handle unsafe.Pointer }

func (o *Observer) raw() *C.KuiObserver {
	if o == nil {
		return nil
	}
	return (*C.KuiObserver)(o.handle)
}

func newObserver(state *observerState, create func(unsafe.Pointer, **C.KuiObserver) C.KuiStatus, context string) (*Observer, error) {
	h, user := callbackHandle(state)
	var out *C.KuiObserver
	if err := statusError(create(user, &out), context); err != nil {
		h.Delete()
		return nil, err
	}
	return &Observer{handle: unsafe.Pointer(out)}, nil
}

// NewMutationObserver 创建 MutationObserver；变化发生后同步回调。
func NewMutationObserver(fn func([]MutationRecord)) (*Observer, error) {
	if fn == nil {
		return nil, &Error{Status: ErrInvalidArg, Context: "NewMutationObserver", Message: "nil callback"}
	}
	return newObserver(&observerState{mutation: fn}, func(user unsafe.Pointer, out **C.KuiObserver) C.KuiStatus {
		return C.kui_mutation_observer_create(C.kui_go_mutation_trampoline(), user, C.kui_go_release_trampoline(), out)
	}, "NewMutationObserver")
}

// NewResizeObserver 创建 ResizeObserver；在渲染帧排版之后回调（首次观察总会报告一次）。
func NewResizeObserver(fn func([]ResizeEntry)) (*Observer, error) {
	if fn == nil {
		return nil, &Error{Status: ErrInvalidArg, Context: "NewResizeObserver", Message: "nil callback"}
	}
	return newObserver(&observerState{resize: fn}, func(user unsafe.Pointer, out **C.KuiObserver) C.KuiStatus {
		return C.kui_resize_observer_create(C.kui_go_resize_observer_trampoline(), user, C.kui_go_release_trampoline(), out)
	}, "NewResizeObserver")
}

// NewIntersectionObserver 创建 IntersectionObserver；options 为 nil 表示视口、阈值 {0}。
func NewIntersectionObserver(options *IntersectionOptions, fn func([]IntersectionEntry)) (*Observer, error) {
	if fn == nil {
		return nil, &Error{Status: ErrInvalidArg, Context: "NewIntersectionObserver", Message: "nil callback"}
	}
	var raw *C.KuiCIntersectionOptions
	if options != nil {
		raw = (*C.KuiCIntersectionOptions)(C.calloc(1, C.size_t(unsafe.Sizeof(C.KuiCIntersectionOptions{}))))
		defer C.free(unsafe.Pointer(raw))
		raw.root = options.Root.raw()
		raw.margin_top = C.float(options.MarginTop)
		raw.margin_right = C.float(options.MarginRight)
		raw.margin_bottom = C.float(options.MarginBottom)
		raw.margin_left = C.float(options.MarginLeft)
		if len(options.Thresholds) > 0 {
			thresholds := (*C.float)(C.malloc(C.size_t(len(options.Thresholds)) * C.size_t(unsafe.Sizeof(C.float(0)))))
			defer C.free(unsafe.Pointer(thresholds))
			for i, value := range options.Thresholds {
				unsafe.Slice(thresholds, len(options.Thresholds))[i] = C.float(value)
			}
			raw.thresholds = thresholds
			raw.threshold_count = C.size_t(len(options.Thresholds))
		}
	}
	return newObserver(&observerState{intersection: fn}, func(user unsafe.Pointer, out **C.KuiObserver) C.KuiStatus {
		return C.kui_intersection_observer_create(raw, C.kui_go_intersection_trampoline(), user, C.kui_go_release_trampoline(), out)
	}, "NewIntersectionObserver")
}

// ObserveMutations 让 MutationObserver 观察 target；再次观察同一目标会替换选项。
func (o *Observer) ObserveMutations(target *Element, options MutationOptions) error {
	raw := (*C.KuiCMutationOptions)(C.calloc(1, C.size_t(unsafe.Sizeof(C.KuiCMutationOptions{}))))
	defer C.free(unsafe.Pointer(raw))
	raw.child_list = boolInt(options.ChildList)
	raw.attributes = boolInt(options.Attributes)
	raw.character_data = boolInt(options.CharacterData)
	raw.subtree = boolInt(options.Subtree)
	raw.attribute_old_value = boolInt(options.AttributeOldValue)
	raw.character_data_old_value = boolInt(options.CharacterDataOldValue)
	if count := len(options.AttributeFilter); count > 0 {
		names := (**C.char)(C.calloc(C.size_t(count), C.size_t(unsafe.Sizeof((*C.char)(nil)))))
		defer C.free(unsafe.Pointer(names))
		slots := unsafe.Slice(names, count)
		for i, name := range options.AttributeFilter {
			slots[i] = C.CString(name)
		}
		defer func() {
			for _, name := range slots {
				C.free(unsafe.Pointer(name))
			}
		}()
		raw.attribute_filter = names
		raw.attribute_filter_count = C.size_t(count)
	}
	return statusError(C.kui_mutation_observer_observe(o.raw(), target.raw(), raw), "Observer.ObserveMutations")
}

// Observe 让 Resize / Intersection Observer 观察 target（MutationObserver 请用 ObserveMutations）。
func (o *Observer) Observe(target *Element) error {
	return statusError(C.kui_observer_observe(o.raw(), target.raw()), "Observer.Observe")
}

// Unobserve 停止观察一个目标（MutationObserver 返回 ErrNotSupported）。
func (o *Observer) Unobserve(target *Element) error {
	return statusError(C.kui_observer_unobserve(o.raw(), target.raw()), "Observer.Unobserve")
}

// Disconnect 停止观察全部目标并丢弃未投递的记录。
func (o *Observer) Disconnect() error {
	return statusError(C.kui_observer_disconnect(o.raw()), "Observer.Disconnect")
}

// Clone 再取一个引用（kui_observer_retain）：每个句柄各自 Close，最后一个关闭时才断开观察。
func (o *Observer) Clone() (*Observer, error) {
	if o == nil || o.handle == nil {
		return nil, &Error{Status: ErrInvalidArg, Context: "Observer.Clone", Message: "observer is closed"}
	}
	if err := statusError(C.kui_observer_retain(o.raw()), "Observer.Clone"); err != nil {
		return nil, err
	}
	return &Observer{handle: o.handle}, nil
}

// Close 释放引用（最后一个引用会断开观察）。
func (o *Observer) Close() error {
	if o == nil || o.handle == nil {
		return nil
	}
	err := statusError(C.kui_observer_release(o.raw()), "Observer.Close")
	o.handle = nil
	return err
}
