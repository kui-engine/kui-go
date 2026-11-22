/* Go/cgo 回调 trampoline；C ABI 函数指针不能直接指向 Go closure。 */
#ifndef KUI_GO_CALLBACKS_H
#define KUI_GO_CALLBACKS_H

#include <stdint.h>
#include "loader.h"

KuiTaskFn kui_go_task_trampoline(void);
KuiTaskFn kui_go_task_discard_trampoline(void);
KuiMenuActivateFn kui_go_menu_activate_trampoline(void);
KuiTrayEventFn kui_go_tray_event_trampoline(void);
KuiWindowResizeFn kui_go_resize_trampoline(void);
KuiEventListenerFn kui_go_event_trampoline(void);
KuiWindowEventFn kui_go_window_event_trampoline(void);
KuiScriptErrorFn kui_go_script_error_trampoline(void);
KuiHttpResponseFn kui_go_http_response_trampoline(void);
KuiCustomElementLifecycleFn kui_go_custom_connected_trampoline(void);
KuiCustomElementLifecycleFn kui_go_custom_disconnected_trampoline(void);
KuiCustomElementAttributeFn kui_go_custom_attribute_trampoline(void);
KuiTaskFn kui_go_release_trampoline(void);
KuiRouteViewFn kui_go_route_view_trampoline(void);
KuiRouteGuardFn kui_go_route_guard_trampoline(void);
KuiRouteInfoFn kui_go_route_info_trampoline(void);
KuiMutationFn kui_go_mutation_trampoline(void);
KuiResizeFn kui_go_resize_observer_trampoline(void);
KuiIntersectionFn kui_go_intersection_trampoline(void);
KuiScriptFunctionFn kui_go_script_function_trampoline(void);
KuiTaskFn kui_go_script_function_release_trampoline(void);
KuiScriptPrintFn kui_go_script_print_trampoline(void);
KuiCustomElementPaintFn kui_go_custom_paint_trampoline(void);
KuiCustomElementTickFn kui_go_custom_tick_trampoline(void);
KuiCustomElementNeedsTickFn kui_go_custom_needs_tick_trampoline(void);
KuiCustomElementKeyFn kui_go_custom_key_trampoline(void);
KuiCustomElementMouseFn kui_go_custom_mouse_trampoline(void);
KuiCustomElementWheelFn kui_go_custom_wheel_trampoline(void);
KuiCustomElementImeFn kui_go_custom_ime_trampoline(void);
KuiTaskFn kui_go_component_release_trampoline(void);
KuiIndexFn kui_go_component_select_trampoline(void);
KuiIndexFn kui_go_component_activate_trampoline(void);
KuiVirtualListRenderFn kui_go_virtual_list_render_trampoline(void);
KuiVirtualListHeightFn kui_go_virtual_list_height_trampoline(void);
KuiVirtualListFocusFn kui_go_virtual_list_focus_trampoline(void);
KuiListViewItemFn kui_go_list_view_item_trampoline(void);
KuiTreeToggleFn kui_go_tree_toggle_trampoline(void);
KuiTreeSelectionFn kui_go_tree_selection_trampoline(void);
void* kui_go_handle_pointer(uintptr_t handle);
/* 当前 OS 线程的 ID（测试核对回调在哪个线程上执行）。 */
uint64_t kui_go_os_thread_id(void);

#endif
