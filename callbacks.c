//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

/* syscall() 在严格 -std=c11 下要 _GNU_SOURCE 才有声明（Linux）。 */
#if defined(__linux__) && !defined(_GNU_SOURCE)
#define _GNU_SOURCE
#endif
#include "callbacks.h"

#if defined(_WIN32)
#define WIN32_LEAN_AND_MEAN 1
#include <windows.h>
uint64_t kui_go_os_thread_id(void) { return (uint64_t)GetCurrentThreadId(); }
#elif defined(__APPLE__)
#include <pthread.h>
uint64_t kui_go_os_thread_id(void) {
    uint64_t id = 0;
    pthread_threadid_np(NULL, &id);
    return id;
}
#else
#include <sys/syscall.h>
#include <unistd.h>
uint64_t kui_go_os_thread_id(void) { return (uint64_t)syscall(SYS_gettid); }
#endif

extern void goKuiTaskCallback(uintptr_t handle);
extern void goKuiReleaseCallback(uintptr_t handle);
extern KuiElement* goKuiRouteViewCallback(KuiDocument* doc, const KuiCRouteInfo* route,
                                          uintptr_t handle);
extern int goKuiRouteGuardCallback(const KuiCRouteInfo* from, const KuiCRouteInfo* to,
                                   uintptr_t handle);
extern void goKuiRouteInfoCallback(const KuiCRouteInfo* route, uintptr_t handle);
extern void goKuiMutationCallback(const KuiCMutationRecord* records, size_t count,
                                  uintptr_t handle);
extern void goKuiResizeObserverCallback(const KuiCResizeEntry* entries, size_t count,
                                        uintptr_t handle);
extern void goKuiIntersectionCallback(const KuiCIntersectionEntry* entries, size_t count,
                                      uintptr_t handle);
extern void goKuiTaskDiscardCallback(uintptr_t handle);
extern KuiStatus goKuiScriptFunctionCallback(KuiCScriptValue* args, size_t count,
                                             KuiCScriptValue* out, uintptr_t handle);
extern void goKuiScriptFunctionRelease(uintptr_t handle);
extern void goKuiScriptPrintCallback(char* message, size_t size, uintptr_t handle);
extern void goKuiCustomPaintCallback(KuiElement* element, KuiPaintContext* paint, uintptr_t handle);
extern void goKuiCustomTickCallback(KuiElement* element, double delta, uintptr_t handle);
extern int goKuiCustomNeedsTickCallback(KuiElement* element, uintptr_t handle);
extern int goKuiCustomKeyCallback(KuiElement* element, KuiCCustomKeyEvent* event, uintptr_t handle);
extern int goKuiCustomMouseCallback(KuiElement* element, KuiCCustomMouseEvent* event,
                                    uintptr_t handle);
extern int goKuiCustomWheelCallback(KuiElement* element, KuiCCustomWheelEvent* event,
                                    uintptr_t handle);
extern int goKuiCustomIMECallback(KuiElement* element, KuiCCustomImeEvent* event, uintptr_t handle);
extern void goKuiComponentRelease(uintptr_t handle);
extern void goKuiComponentSelect(uint32_t index, uintptr_t handle);
extern void goKuiComponentActivate(uint32_t index, uintptr_t handle);
extern KuiElement* goKuiVirtualListRender(uint32_t index, uintptr_t handle);
extern float goKuiVirtualListHeight(uint32_t index, uintptr_t handle);
extern void goKuiVirtualListFocus(int32_t old_index, int32_t new_index, uintptr_t handle);
extern void goKuiListViewItem(uint32_t index, KuiCListViewItem* out, uintptr_t handle);
extern void goKuiTreeToggle(uint32_t node, int expanded, uintptr_t handle);
extern void goKuiTreeSelection(uint32_t* selected, size_t count, uintptr_t handle);
extern void goKuiMenuActivateCallback(uint32_t id, uintptr_t handle);
extern void goKuiTrayEventCallback(int event, uintptr_t handle);
extern void goKuiResizeCallback(uint32_t width, uint32_t height, uintptr_t handle);
extern void goKuiEventCallback(KuiEvent* event, uintptr_t handle);
extern void goKuiWindowEventCallback(KuiEvent* event, uintptr_t handle);
extern void goKuiScriptErrorCallback(const char* error, size_t error_len,
                                     const char* source, size_t source_len,
                                     int line, uintptr_t handle);
extern void goKuiHttpResponseCallback(KuiHttpResponse* response, uintptr_t handle);
extern void goKuiCustomConnectedCallback(KuiElement* element, uintptr_t handle);
extern void goKuiCustomDisconnectedCallback(KuiElement* element, uintptr_t handle);
extern void goKuiCustomAttributeCallback(KuiElement* element,
                                         const char* name, size_t name_len,
                                         const char* old_value, size_t old_len,
                                         const char* new_value, size_t new_len,
                                         uintptr_t handle);

static void task_callback(void* user_data) {
    goKuiTaskCallback((uintptr_t)user_data);
}
static void release_callback(void* user_data) {
    goKuiReleaseCallback((uintptr_t)user_data);
}
static KuiElement* route_view_callback(KuiDocument* doc, const KuiCRouteInfo* route,
                                       void* user_data) {
    return goKuiRouteViewCallback(doc, route, (uintptr_t)user_data);
}
static int route_guard_callback(const KuiCRouteInfo* from, const KuiCRouteInfo* to,
                                void* user_data) {
    return goKuiRouteGuardCallback(from, to, (uintptr_t)user_data);
}
static void route_info_callback(const KuiCRouteInfo* route, void* user_data) {
    goKuiRouteInfoCallback(route, (uintptr_t)user_data);
}
static void mutation_callback(const KuiCMutationRecord* records, size_t count,
                              void* user_data) {
    goKuiMutationCallback(records, count, (uintptr_t)user_data);
}
static void resize_observer_callback(const KuiCResizeEntry* entries, size_t count,
                                     void* user_data) {
    goKuiResizeObserverCallback(entries, count, (uintptr_t)user_data);
}
static void intersection_callback(const KuiCIntersectionEntry* entries, size_t count,
                                  void* user_data) {
    goKuiIntersectionCallback(entries, count, (uintptr_t)user_data);
}
static KuiStatus script_function_callback(const KuiCScriptValue* args, size_t count,
                                          KuiCScriptValue* out, void* user_data) {
    return goKuiScriptFunctionCallback((KuiCScriptValue*)args, count, out,
                                       (uintptr_t)user_data);
}
static void script_function_release(void* user_data) {
    goKuiScriptFunctionRelease((uintptr_t)user_data);
}
static void script_print_callback(const char* message, size_t size, void* user_data) {
    goKuiScriptPrintCallback((char*)message, size, (uintptr_t)user_data);
}
static void custom_paint_callback(KuiElement* element, KuiPaintContext* paint, void* user_data) {
    goKuiCustomPaintCallback(element, paint, (uintptr_t)user_data);
}
static void custom_tick_callback(KuiElement* element, double delta, void* user_data) {
    goKuiCustomTickCallback(element, delta, (uintptr_t)user_data);
}
static int custom_needs_tick_callback(KuiElement* element, void* user_data) {
    return goKuiCustomNeedsTickCallback(element, (uintptr_t)user_data);
}
static int custom_key_callback(KuiElement* element, const KuiCCustomKeyEvent* event,
                               void* user_data) {
    return goKuiCustomKeyCallback(element, (KuiCCustomKeyEvent*)event, (uintptr_t)user_data);
}
static int custom_mouse_callback(KuiElement* element, const KuiCCustomMouseEvent* event,
                                 void* user_data) {
    return goKuiCustomMouseCallback(element, (KuiCCustomMouseEvent*)event, (uintptr_t)user_data);
}
static int custom_wheel_callback(KuiElement* element, const KuiCCustomWheelEvent* event,
                                 void* user_data) {
    return goKuiCustomWheelCallback(element, (KuiCCustomWheelEvent*)event, (uintptr_t)user_data);
}
static int custom_ime_callback(KuiElement* element, const KuiCCustomImeEvent* event,
                               void* user_data) {
    return goKuiCustomIMECallback(element, (KuiCCustomImeEvent*)event, (uintptr_t)user_data);
}
static void component_release(void* user_data) { goKuiComponentRelease((uintptr_t)user_data); }
static void component_select(uint32_t index, void* user_data) {
    goKuiComponentSelect(index, (uintptr_t)user_data);
}
static void component_activate(uint32_t index, void* user_data) {
    goKuiComponentActivate(index, (uintptr_t)user_data);
}
static KuiElement* virtual_list_render(KuiDocument* document, uint32_t index, void* user_data) {
    (void)document;
    return goKuiVirtualListRender(index, (uintptr_t)user_data);
}
static float virtual_list_height(uint32_t index, void* user_data) {
    return goKuiVirtualListHeight(index, (uintptr_t)user_data);
}
static void virtual_list_focus(int32_t old_index, int32_t new_index, void* user_data) {
    goKuiVirtualListFocus(old_index, new_index, (uintptr_t)user_data);
}
static void list_view_item(uint32_t index, KuiCListViewItem* out, void* user_data) {
    goKuiListViewItem(index, out, (uintptr_t)user_data);
}
static void tree_toggle(uint32_t node, int expanded, void* user_data) {
    goKuiTreeToggle(node, expanded, (uintptr_t)user_data);
}
static void tree_selection(const uint32_t* selected, size_t count, void* user_data) {
    goKuiTreeSelection((uint32_t*)selected, count, (uintptr_t)user_data);
}
static void task_discard_callback(void* user_data) {
    goKuiTaskDiscardCallback((uintptr_t)user_data);
}
static void menu_activate_callback(uint32_t id, void* user_data) {
    goKuiMenuActivateCallback(id, (uintptr_t)user_data);
}
static void tray_event_callback(KuiCTrayEvent event, void* user_data) {
    goKuiTrayEventCallback((int)event, (uintptr_t)user_data);
}
static void resize_callback(uint32_t width, uint32_t height, void* user_data) {
    goKuiResizeCallback(width, height, (uintptr_t)user_data);
}
static void event_callback(KuiEvent* event, void* user_data) {
    goKuiEventCallback(event, (uintptr_t)user_data);
}
static void window_event_callback(KuiEvent* event, void* user_data) {
    goKuiWindowEventCallback(event, (uintptr_t)user_data);
}
static void script_error_callback(const char* error, size_t error_len,
                                  const char* source, size_t source_len,
                                  int line, void* user_data) {
    goKuiScriptErrorCallback(error, error_len, source, source_len, line,
                             (uintptr_t)user_data);
}
static void http_response_callback(KuiHttpResponse* response, void* user_data) {
    goKuiHttpResponseCallback(response, (uintptr_t)user_data);
}
static void custom_connected_callback(KuiElement* element, void* user_data) {
    goKuiCustomConnectedCallback(element, (uintptr_t)user_data);
}
static void custom_disconnected_callback(KuiElement* element, void* user_data) {
    goKuiCustomDisconnectedCallback(element, (uintptr_t)user_data);
}
static void custom_attribute_callback(KuiElement* element,
                                      const char* name, size_t name_len,
                                      const char* old_value, size_t old_len,
                                      const char* new_value, size_t new_len,
                                      void* user_data) {
    goKuiCustomAttributeCallback(element, name, name_len, old_value, old_len,
                                 new_value, new_len, (uintptr_t)user_data);
}

KuiTaskFn kui_go_task_trampoline(void) { return task_callback; }
KuiTaskFn kui_go_task_discard_trampoline(void) { return task_discard_callback; }
KuiTaskFn kui_go_release_trampoline(void) { return release_callback; }
KuiRouteViewFn kui_go_route_view_trampoline(void) { return route_view_callback; }
KuiRouteGuardFn kui_go_route_guard_trampoline(void) { return route_guard_callback; }
KuiRouteInfoFn kui_go_route_info_trampoline(void) { return route_info_callback; }
KuiMutationFn kui_go_mutation_trampoline(void) { return mutation_callback; }
KuiResizeFn kui_go_resize_observer_trampoline(void) { return resize_observer_callback; }
KuiIntersectionFn kui_go_intersection_trampoline(void) { return intersection_callback; }
KuiScriptFunctionFn kui_go_script_function_trampoline(void) { return script_function_callback; }
KuiTaskFn kui_go_script_function_release_trampoline(void) { return script_function_release; }
KuiScriptPrintFn kui_go_script_print_trampoline(void) { return script_print_callback; }
KuiCustomElementPaintFn kui_go_custom_paint_trampoline(void) { return custom_paint_callback; }
KuiCustomElementTickFn kui_go_custom_tick_trampoline(void) { return custom_tick_callback; }
KuiCustomElementNeedsTickFn kui_go_custom_needs_tick_trampoline(void) {
    return custom_needs_tick_callback;
}
KuiCustomElementKeyFn kui_go_custom_key_trampoline(void) { return custom_key_callback; }
KuiCustomElementMouseFn kui_go_custom_mouse_trampoline(void) { return custom_mouse_callback; }
KuiCustomElementWheelFn kui_go_custom_wheel_trampoline(void) { return custom_wheel_callback; }
KuiCustomElementImeFn kui_go_custom_ime_trampoline(void) { return custom_ime_callback; }
KuiTaskFn kui_go_component_release_trampoline(void) { return component_release; }
KuiIndexFn kui_go_component_select_trampoline(void) { return component_select; }
KuiIndexFn kui_go_component_activate_trampoline(void) { return component_activate; }
KuiVirtualListRenderFn kui_go_virtual_list_render_trampoline(void) { return virtual_list_render; }
KuiVirtualListHeightFn kui_go_virtual_list_height_trampoline(void) { return virtual_list_height; }
KuiVirtualListFocusFn kui_go_virtual_list_focus_trampoline(void) { return virtual_list_focus; }
KuiListViewItemFn kui_go_list_view_item_trampoline(void) { return list_view_item; }
KuiTreeToggleFn kui_go_tree_toggle_trampoline(void) { return tree_toggle; }
KuiTreeSelectionFn kui_go_tree_selection_trampoline(void) { return tree_selection; }
KuiMenuActivateFn kui_go_menu_activate_trampoline(void) { return menu_activate_callback; }
KuiTrayEventFn kui_go_tray_event_trampoline(void) { return tray_event_callback; }
KuiWindowResizeFn kui_go_resize_trampoline(void) { return resize_callback; }
KuiEventListenerFn kui_go_event_trampoline(void) { return event_callback; }
KuiWindowEventFn kui_go_window_event_trampoline(void) { return window_event_callback; }
KuiScriptErrorFn kui_go_script_error_trampoline(void) { return script_error_callback; }
KuiHttpResponseFn kui_go_http_response_trampoline(void) { return http_response_callback; }
KuiCustomElementLifecycleFn kui_go_custom_connected_trampoline(void) {
    return custom_connected_callback;
}
KuiCustomElementLifecycleFn kui_go_custom_disconnected_trampoline(void) {
    return custom_disconnected_callback;
}
KuiCustomElementAttributeFn kui_go_custom_attribute_trampoline(void) {
    return custom_attribute_callback;
}
void* kui_go_handle_pointer(uintptr_t handle) { return (void*)handle; }
