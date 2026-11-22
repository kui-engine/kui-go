/* 由 scripts/generate_go_loader.py 生成；禁止手改。 */
#define WIN32_LEAN_AND_MEAN 1
#include <windows.h>
#include <stdio.h>
#include <string.h>
#include "loader.h"

static HMODULE g_kui_module = NULL;
static char g_kui_loader_error[512] = {0};

typedef KuiVersion (__cdecl *kui_go_pfn_kui_get_version)(void);
static kui_go_pfn_kui_get_version g_kui_get_version = NULL;
typedef const char * (__cdecl *kui_go_pfn_kui_status_message)(KuiStatus s);
static kui_go_pfn_kui_status_message g_kui_status_message = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_init)(void);
static kui_go_pfn_kui_platform_init g_kui_platform_init = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_shutdown)(void);
static kui_go_pfn_kui_platform_shutdown g_kui_platform_shutdown = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_run)(int * out_exit_code);
static kui_go_pfn_kui_platform_run g_kui_platform_run = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_quit)(int exit_code);
static kui_go_pfn_kui_platform_quit g_kui_platform_quit = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_poll_events)(void);
static kui_go_pfn_kui_platform_poll_events g_kui_platform_poll_events = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_wait_events)(double timeout_seconds);
static kui_go_pfn_kui_platform_wait_events g_kui_platform_wait_events = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_set_timeout)(KuiTaskFn fn, void * user_data, uint32_t delay_ms, KuiTimerId * out_id);
static kui_go_pfn_kui_platform_set_timeout g_kui_platform_set_timeout = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_set_interval)(KuiTaskFn fn, void * user_data, uint32_t interval_ms, KuiTimerId * out_id);
static kui_go_pfn_kui_platform_set_interval g_kui_platform_set_interval = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_clear_timer)(KuiTimerId id);
static kui_go_pfn_kui_platform_clear_timer g_kui_platform_clear_timer = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_request_idle)(KuiTaskFn fn, void * user_data);
static kui_go_pfn_kui_platform_request_idle g_kui_platform_request_idle = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_post_to_main_thread)(KuiTaskFn fn, KuiTaskFn discard_fn, void * user_data);
static kui_go_pfn_kui_platform_post_to_main_thread g_kui_platform_post_to_main_thread = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_set_f12_devtools_enabled)(int enabled);
static kui_go_pfn_kui_platform_set_f12_devtools_enabled g_kui_platform_set_f12_devtools_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_primary_monitor)(KuiCMonitorInfo * out_monitor);
static kui_go_pfn_kui_platform_get_primary_monitor g_kui_platform_get_primary_monitor = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_create)(const KuiWindowProps * props, KuiWindow ** out);
static kui_go_pfn_kui_window_create g_kui_window_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_release)(KuiWindow * w);
static kui_go_pfn_kui_window_release g_kui_window_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_show)(KuiWindow * w);
static kui_go_pfn_kui_window_show g_kui_window_show = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_hide)(KuiWindow * w);
static kui_go_pfn_kui_window_hide g_kui_window_hide = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_is_visible)(KuiWindow * w, int * out_visible);
static kui_go_pfn_kui_window_is_visible g_kui_window_is_visible = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_document)(KuiWindow * w, KuiDocument ** out);
static kui_go_pfn_kui_window_get_document g_kui_window_get_document = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_title)(KuiWindow * w, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_window_get_title g_kui_window_get_title = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_title)(KuiWindow * w, const char * title, size_t len);
static kui_go_pfn_kui_window_set_title g_kui_window_set_title = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_size)(KuiWindow * w, int * out_width, int * out_height);
static kui_go_pfn_kui_window_get_size g_kui_window_get_size = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_resize)(KuiWindow * w, int width, int height);
static kui_go_pfn_kui_window_resize g_kui_window_resize = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_position)(KuiWindow * w, int * out_x, int * out_y);
static kui_go_pfn_kui_window_get_position g_kui_window_get_position = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_position)(KuiWindow * w, int x, int y);
static kui_go_pfn_kui_window_set_position g_kui_window_set_position = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_minimize)(KuiWindow * w);
static kui_go_pfn_kui_window_minimize g_kui_window_minimize = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_maximize)(KuiWindow * w);
static kui_go_pfn_kui_window_maximize g_kui_window_maximize = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_restore)(KuiWindow * w);
static kui_go_pfn_kui_window_restore g_kui_window_restore = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_always_on_top)(KuiWindow * w, int on_top);
static kui_go_pfn_kui_window_set_always_on_top g_kui_window_set_always_on_top = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_opacity)(KuiWindow * w, float opacity);
static kui_go_pfn_kui_window_set_opacity g_kui_window_set_opacity = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_device_pixel_ratio)(KuiWindow * w, float * out_ratio);
static kui_go_pfn_kui_window_get_device_pixel_ratio g_kui_window_get_device_pixel_ratio = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_is_minimized)(KuiWindow * w, int * out_minimized);
static kui_go_pfn_kui_window_is_minimized g_kui_window_is_minimized = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_is_maximized)(KuiWindow * w, int * out_maximized);
static kui_go_pfn_kui_window_is_maximized g_kui_window_is_maximized = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_close)(KuiWindow * w);
static kui_go_pfn_kui_window_close g_kui_window_close = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_should_close)(KuiWindow * w, int * out_should_close);
static kui_go_pfn_kui_window_should_close g_kui_window_should_close = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_on_resize)(KuiWindow * w, KuiWindowResizeFn fn, void * user_data);
static kui_go_pfn_kui_window_on_resize g_kui_window_on_resize = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_start_devtools)(KuiWindow * w, uint16_t port, int * out_started);
static kui_go_pfn_kui_window_start_devtools g_kui_window_start_devtools = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_devtools_url)(KuiWindow * w, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_window_get_devtools_url g_kui_window_get_devtools_url = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_native_handle)(KuiWindow * w, void ** out_handle);
static kui_go_pfn_kui_window_get_native_handle g_kui_window_get_native_handle = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_create)(KuiDocument ** out);
static kui_go_pfn_kui_document_create g_kui_document_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_retain)(KuiDocument * doc);
static kui_go_pfn_kui_document_retain g_kui_document_retain = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_release)(KuiDocument * doc);
static kui_go_pfn_kui_document_release g_kui_document_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_load_html)(KuiDocument * doc, const char * html, size_t html_len);
static kui_go_pfn_kui_document_load_html g_kui_document_load_html = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_load_html_file)(KuiDocument * doc, const char * path, size_t path_len);
static kui_go_pfn_kui_document_load_html_file g_kui_document_load_html_file = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_add_stylesheet)(KuiDocument * doc, const char * css, size_t css_len);
static kui_go_pfn_kui_document_add_stylesheet g_kui_document_add_stylesheet = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_clear_stylesheets)(KuiDocument * doc);
static kui_go_pfn_kui_document_clear_stylesheets g_kui_document_clear_stylesheets = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_root)(KuiDocument * doc, KuiElement ** out);
static kui_go_pfn_kui_document_root g_kui_document_root = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_body)(KuiDocument * doc, KuiElement ** out);
static kui_go_pfn_kui_document_body g_kui_document_body = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_head)(KuiDocument * doc, KuiElement ** out);
static kui_go_pfn_kui_document_head g_kui_document_head = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_element_by_id)(KuiDocument * doc, const char * id, size_t id_len, KuiElement ** out);
static kui_go_pfn_kui_document_get_element_by_id g_kui_document_get_element_by_id = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_query_selector)(KuiDocument * doc, const char * sel, size_t sel_len, KuiElement ** out);
static kui_go_pfn_kui_document_query_selector g_kui_document_query_selector = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_query_selector_all)(KuiDocument * doc, const char * sel, size_t sel_len, KuiElement ** out_buf, size_t buf_capacity, size_t * out_count);
static kui_go_pfn_kui_document_query_selector_all g_kui_document_query_selector_all = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_create_element)(KuiDocument * doc, const char * tag, size_t tag_len, KuiElement ** out);
static kui_go_pfn_kui_document_create_element g_kui_document_create_element = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_retain)(KuiElement * el);
static kui_go_pfn_kui_element_retain g_kui_element_retain = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_release)(KuiElement * el);
static kui_go_pfn_kui_element_release g_kui_element_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_tag_name)(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_tag_name g_kui_element_get_tag_name = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_id)(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_id g_kui_element_get_id = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_class_name)(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_class_name g_kui_element_get_class_name = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_class_name)(KuiElement * el, const char * value, size_t value_len);
static kui_go_pfn_kui_element_set_class_name g_kui_element_set_class_name = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_attribute)(KuiElement * el, const char * name, size_t name_len, const char * value, size_t value_len);
static kui_go_pfn_kui_element_set_attribute g_kui_element_set_attribute = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_attribute)(KuiElement * el, const char * name, size_t name_len, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_attribute g_kui_element_get_attribute = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_has_attribute)(KuiElement * el, const char * name, size_t name_len, int * out_has);
static kui_go_pfn_kui_element_has_attribute g_kui_element_has_attribute = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_remove_attribute)(KuiElement * el, const char * name, size_t name_len);
static kui_go_pfn_kui_element_remove_attribute g_kui_element_remove_attribute = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_value)(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_value g_kui_element_get_value = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_value)(KuiElement * el, const char * value, size_t value_len);
static kui_go_pfn_kui_element_set_value g_kui_element_set_value = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_checked)(KuiElement * el, int * out_checked);
static kui_go_pfn_kui_element_get_checked g_kui_element_get_checked = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_checked)(KuiElement * el, int checked);
static kui_go_pfn_kui_element_set_checked g_kui_element_set_checked = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_selected_index)(KuiElement * el, int32_t * out_index);
static kui_go_pfn_kui_element_get_selected_index g_kui_element_get_selected_index = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_selected_index)(KuiElement * el, int32_t index);
static kui_go_pfn_kui_element_set_selected_index g_kui_element_set_selected_index = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_text_content)(KuiElement * el, const char * text, size_t text_len);
static kui_go_pfn_kui_element_set_text_content g_kui_element_set_text_content = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_text_content)(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_text_content g_kui_element_get_text_content = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_inner_html)(KuiElement * el, const char * html, size_t html_len);
static kui_go_pfn_kui_element_set_inner_html g_kui_element_set_inner_html = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_inner_html)(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_inner_html g_kui_element_get_inner_html = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_class_list_add)(KuiElement * el, const char * token, size_t len);
static kui_go_pfn_kui_element_class_list_add g_kui_element_class_list_add = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_class_list_remove)(KuiElement * el, const char * token, size_t len);
static kui_go_pfn_kui_element_class_list_remove g_kui_element_class_list_remove = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_class_list_toggle)(KuiElement * el, const char * token, size_t len);
static kui_go_pfn_kui_element_class_list_toggle g_kui_element_class_list_toggle = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_class_list_contains)(KuiElement * el, const char * token, size_t len, int * out_contains);
static kui_go_pfn_kui_element_class_list_contains g_kui_element_class_list_contains = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_query_selector)(KuiElement * el, const char * sel, size_t sel_len, KuiElement ** out);
static kui_go_pfn_kui_element_query_selector g_kui_element_query_selector = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_query_selector_all)(KuiElement * el, const char * sel, size_t sel_len, KuiElement ** out_buf, size_t buf_capacity, size_t * out_count);
static kui_go_pfn_kui_element_query_selector_all g_kui_element_query_selector_all = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_parent)(KuiElement * el, KuiElement ** out);
static kui_go_pfn_kui_element_parent g_kui_element_parent = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_first_child)(KuiElement * el, KuiElement ** out);
static kui_go_pfn_kui_element_first_child g_kui_element_first_child = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_last_child)(KuiElement * el, KuiElement ** out);
static kui_go_pfn_kui_element_last_child g_kui_element_last_child = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_next_sibling)(KuiElement * el, KuiElement ** out);
static kui_go_pfn_kui_element_next_sibling g_kui_element_next_sibling = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_previous_sibling)(KuiElement * el, KuiElement ** out);
static kui_go_pfn_kui_element_previous_sibling g_kui_element_previous_sibling = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_child_count)(KuiElement * el, size_t * out_count);
static kui_go_pfn_kui_element_child_count g_kui_element_child_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_append_child)(KuiElement * parent, KuiElement * child);
static kui_go_pfn_kui_element_append_child g_kui_element_append_child = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_remove_child)(KuiElement * parent, KuiElement * child);
static kui_go_pfn_kui_element_remove_child g_kui_element_remove_child = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_remove)(KuiElement * el);
static kui_go_pfn_kui_element_remove g_kui_element_remove = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_scroll_position)(KuiElement * el, float left, float top);
static kui_go_pfn_kui_element_set_scroll_position g_kui_element_set_scroll_position = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_scroll_by)(KuiElement * el, float dx, float dy);
static kui_go_pfn_kui_element_scroll_by g_kui_element_scroll_by = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_focus)(KuiElement * el);
static kui_go_pfn_kui_element_focus g_kui_element_focus = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_blur)(KuiElement * el);
static kui_go_pfn_kui_element_blur g_kui_element_blur = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_add_event_listener)(KuiElement * el, const char * event_name, size_t len, KuiEventListenerFn fn, void * user_data, KuiListenerId * out_id);
static kui_go_pfn_kui_element_add_event_listener g_kui_element_add_event_listener = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_add_event_listener_ex)(KuiElement * el, const char * event_name, size_t len, KuiEventListenerFn fn, void * user_data, int capture, KuiListenerId * out_id);
static kui_go_pfn_kui_element_add_event_listener_ex g_kui_element_add_event_listener_ex = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_remove_event_listener)(KuiElement * el, const char * event_name, size_t len, KuiListenerId id);
static kui_go_pfn_kui_element_remove_event_listener g_kui_element_remove_event_listener = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_remove_event_listeners)(KuiElement * el, const char * event_name, size_t len);
static kui_go_pfn_kui_element_remove_event_listeners g_kui_element_remove_event_listeners = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_dispatch_event)(KuiElement * el, const char * event_name, size_t len, int bubbles, int cancelable, int * out_not_canceled);
static kui_go_pfn_kui_element_dispatch_event g_kui_element_dispatch_event = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_get_type)(KuiEvent * evt, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_event_get_type g_kui_event_get_type = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_get_target)(KuiEvent * evt, KuiElement ** out);
static kui_go_pfn_kui_event_get_target g_kui_event_get_target = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_get_current_target)(KuiEvent * evt, KuiElement ** out);
static kui_go_pfn_kui_event_get_current_target g_kui_event_get_current_target = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_get_related_target)(KuiEvent * evt, KuiElement ** out);
static kui_go_pfn_kui_event_get_related_target g_kui_event_get_related_target = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_prevent_default)(KuiEvent * evt);
static kui_go_pfn_kui_event_prevent_default g_kui_event_prevent_default = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_stop_propagation)(KuiEvent * evt);
static kui_go_pfn_kui_event_stop_propagation g_kui_event_stop_propagation = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_get_flags)(KuiEvent * evt, int * out_bubbles, int * out_cancelable, int * out_default_prevented);
static kui_go_pfn_kui_event_get_flags g_kui_event_get_flags = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_mouse_get_client_xy)(KuiEvent * evt, float * x, float * y);
static kui_go_pfn_kui_event_mouse_get_client_xy g_kui_event_mouse_get_client_xy = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_mouse_get_page_xy)(KuiEvent * evt, float * x, float * y);
static kui_go_pfn_kui_event_mouse_get_page_xy g_kui_event_mouse_get_page_xy = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_mouse_get_offset_xy)(KuiEvent * evt, float * x, float * y);
static kui_go_pfn_kui_event_mouse_get_offset_xy g_kui_event_mouse_get_offset_xy = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_mouse_get_button)(KuiEvent * evt, int * out_button);
static kui_go_pfn_kui_event_mouse_get_button g_kui_event_mouse_get_button = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_key_get_key)(KuiEvent * evt, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_event_key_get_key g_kui_event_key_get_key = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_key_get_code)(KuiEvent * evt, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_event_key_get_code g_kui_event_key_get_code = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_key_get_modifiers)(KuiEvent * evt, int * out_ctrl, int * out_shift, int * out_alt, int * out_meta);
static kui_go_pfn_kui_event_key_get_modifiers g_kui_event_key_get_modifiers = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_style_property)(KuiElement * el, const char * prop, size_t prop_len, const char * value, size_t value_len, int important);
static kui_go_pfn_kui_element_set_style_property g_kui_element_set_style_property = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_style_property)(KuiElement * el, const char * prop, size_t prop_len, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_style_property g_kui_element_get_style_property = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_remove_style_property)(KuiElement * el, const char * prop, size_t prop_len);
static kui_go_pfn_kui_element_remove_style_property g_kui_element_remove_style_property = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_style_property_is_important)(KuiElement * el, const char * prop, size_t prop_len, int * out_important);
static kui_go_pfn_kui_element_style_property_is_important g_kui_element_style_property_is_important = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_style_text)(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_style_text g_kui_element_get_style_text = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_style_text)(KuiElement * el, const char * css, size_t css_len);
static kui_go_pfn_kui_element_set_style_text g_kui_element_set_style_text = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_style_property_count)(KuiElement * el, size_t * out_count);
static kui_go_pfn_kui_element_get_style_property_count g_kui_element_get_style_property_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_style_property_name)(KuiElement * el, size_t index, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_style_property_name g_kui_element_get_style_property_name = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_bounding_rect)(KuiElement * el, KuiRect * out);
static kui_go_pfn_kui_element_get_bounding_rect g_kui_element_get_bounding_rect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_client_metrics)(KuiElement * el, float * out_left, float * out_top, float * out_width, float * out_height);
static kui_go_pfn_kui_element_get_client_metrics g_kui_element_get_client_metrics = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_offset_metrics)(KuiElement * el, float * out_left, float * out_top, float * out_width, float * out_height);
static kui_go_pfn_kui_element_get_offset_metrics g_kui_element_get_offset_metrics = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_scroll_metrics)(KuiElement * el, float * out_left, float * out_top, float * out_width, float * out_height);
static kui_go_pfn_kui_element_get_scroll_metrics g_kui_element_get_scroll_metrics = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_devtools_dispatch_mouse_event)(KuiDocument * doc, KuiCMouseEventType type, float x, float y, KuiCMouseButton button, int modifiers, int click_count, float delta_x, float delta_y);
static kui_go_pfn_kui_devtools_dispatch_mouse_event g_kui_devtools_dispatch_mouse_event = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_devtools_dispatch_key_event)(KuiDocument * doc, KuiCKeyEventType type, const char * key, size_t key_len, const char * code, size_t code_len, int modifiers, const char * text, size_t text_len, int is_auto_repeat);
static kui_go_pfn_kui_devtools_dispatch_key_event g_kui_devtools_dispatch_key_event = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_devtools_get_box_model)(KuiDocument * doc, const char * selector, float * out_quads, float * out_width, float * out_height);
static kui_go_pfn_kui_devtools_get_box_model g_kui_devtools_get_box_model = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_devtools_dispatch_command)(KuiDocument * doc, const char * json_command, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_devtools_dispatch_command g_kui_devtools_dispatch_command = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_monitor_count)(size_t * out_count);
static kui_go_pfn_kui_platform_get_monitor_count g_kui_platform_get_monitor_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_monitor)(size_t index, KuiCMonitorInfo * out_monitor, char * name_buf, size_t name_capacity, size_t * out_name_len);
static kui_go_pfn_kui_platform_get_monitor g_kui_platform_get_monitor = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_primary_monitor_physical_size)(float * out_width, float * out_height);
static kui_go_pfn_kui_platform_get_primary_monitor_physical_size g_kui_platform_get_primary_monitor_physical_size = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_clipboard_text)(char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_platform_get_clipboard_text g_kui_platform_get_clipboard_text = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_set_clipboard_text)(const char * text, size_t len);
static kui_go_pfn_kui_platform_set_clipboard_text g_kui_platform_set_clipboard_text = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_has_clipboard_text)(int * out_has_text);
static kui_go_pfn_kui_platform_has_clipboard_text g_kui_platform_has_clipboard_text = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_os_name)(char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_platform_get_os_name g_kui_platform_get_os_name = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_os_version)(char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_platform_get_os_version g_kui_platform_get_os_version = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_open_file_dialog)(const KuiCFileDialogOptions * options, char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_platform_open_file_dialog g_kui_platform_open_file_dialog = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_save_file_dialog)(const KuiCFileDialogOptions * options, char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_platform_save_file_dialog g_kui_platform_save_file_dialog = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_select_folder_dialog)(const char * title, size_t title_len, char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_platform_select_folder_dialog g_kui_platform_select_folder_dialog = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_frameless_sizeable)(KuiWindow * window, int enabled);
static kui_go_pfn_kui_window_set_frameless_sizeable g_kui_window_set_frameless_sizeable = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_frame_radius)(KuiWindow * window, float radius);
static kui_go_pfn_kui_window_set_frame_radius g_kui_window_set_frame_radius = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_clear_frame_radius)(KuiWindow * window);
static kui_go_pfn_kui_window_clear_frame_radius g_kui_window_clear_frame_radius = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_devtools_port)(KuiWindow * window, uint16_t * out_port);
static kui_go_pfn_kui_window_get_devtools_port g_kui_window_get_devtools_port = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_add_event_listener)(KuiWindow * window, const char * event_name, size_t len, KuiWindowEventFn fn, void * user_data, int once, KuiListenerId * out_id);
static kui_go_pfn_kui_window_add_event_listener g_kui_window_add_event_listener = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_remove_event_listener)(KuiWindow * window, const char * event_name, size_t len, KuiListenerId id);
static kui_go_pfn_kui_window_remove_event_listener g_kui_window_remove_event_listener = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_remove_event_listeners)(KuiWindow * window, const char * event_name, size_t len);
static kui_go_pfn_kui_window_remove_event_listeners g_kui_window_remove_event_listeners = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_get_type_tag)(KuiEvent * event, KuiCEventTypeTag * out_tag);
static kui_go_pfn_kui_event_get_type_tag g_kui_event_get_type_tag = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_get_file_count)(KuiEvent * event, size_t * out_count);
static kui_go_pfn_kui_event_drag_get_file_count g_kui_event_drag_get_file_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_get_file)(KuiEvent * event, size_t index, char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_event_drag_get_file g_kui_event_drag_get_file = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_position_relative_to)(KuiElement * element, KuiElement * anchor, const char * placement, size_t placement_len, const KuiCPositioningOptions * options);
static kui_go_pfn_kui_element_position_relative_to g_kui_element_position_relative_to = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_play)(KuiElement * element);
static kui_go_pfn_kui_media_play g_kui_media_play = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_pause)(KuiElement * element);
static kui_go_pfn_kui_media_pause g_kui_media_pause = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_set_src)(KuiElement * element, const char * src, size_t len);
static kui_go_pfn_kui_media_set_src g_kui_media_set_src = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_set_controls)(KuiElement * element, int enabled);
static kui_go_pfn_kui_media_set_controls g_kui_media_set_controls = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_set_current_time)(KuiElement * element, double seconds);
static kui_go_pfn_kui_media_set_current_time g_kui_media_set_current_time = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_get_current_time)(KuiElement * element, double * out_seconds);
static kui_go_pfn_kui_media_get_current_time g_kui_media_get_current_time = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_get_duration)(KuiElement * element, double * out_seconds);
static kui_go_pfn_kui_media_get_duration g_kui_media_get_duration = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_set_volume)(KuiElement * element, float volume);
static kui_go_pfn_kui_media_set_volume g_kui_media_set_volume = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_get_volume)(KuiElement * element, float * out_volume);
static kui_go_pfn_kui_media_get_volume g_kui_media_get_volume = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_set_muted)(KuiElement * element, int muted);
static kui_go_pfn_kui_media_set_muted g_kui_media_set_muted = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_get_muted)(KuiElement * element, int * out_muted);
static kui_go_pfn_kui_media_get_muted g_kui_media_get_muted = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_get_paused)(KuiElement * element, int * out_paused);
static kui_go_pfn_kui_media_get_paused g_kui_media_get_paused = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_get_ready_state)(KuiElement * element, int * out_state);
static kui_go_pfn_kui_media_get_ready_state g_kui_media_get_ready_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_get_network_state)(KuiElement * element, int * out_state);
static kui_go_pfn_kui_media_get_network_state g_kui_media_get_network_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_is_available)(KuiElement * element, int * out_available);
static kui_go_pfn_kui_media_is_available g_kui_media_is_available = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_set_src)(KuiElement * element, const char * src, size_t len);
static kui_go_pfn_kui_lottie_set_src g_kui_lottie_set_src = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_play)(KuiElement * element);
static kui_go_pfn_kui_lottie_play g_kui_lottie_play = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_pause)(KuiElement * element);
static kui_go_pfn_kui_lottie_pause g_kui_lottie_pause = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_get_complete)(KuiElement * element, int * out_complete);
static kui_go_pfn_kui_lottie_get_complete g_kui_lottie_get_complete = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_get_natural_size)(KuiElement * element, float * out_width, float * out_height);
static kui_go_pfn_kui_lottie_get_natural_size g_kui_lottie_get_natural_size = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_create)(KuiScriptEngine ** out_engine);
static kui_go_pfn_kui_script_engine_create g_kui_script_engine_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_release)(KuiScriptEngine * engine);
static kui_go_pfn_kui_script_engine_release g_kui_script_engine_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_bind_document)(KuiScriptEngine * engine, KuiDocument * document);
static kui_go_pfn_kui_script_engine_bind_document g_kui_script_engine_bind_document = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_unbind_document)(KuiScriptEngine * engine);
static kui_go_pfn_kui_script_engine_unbind_document g_kui_script_engine_unbind_document = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_eval)(KuiScriptEngine * engine, const char * code, size_t code_len, const char * source, size_t source_len, int * out_success, char * error_buf, size_t error_capacity, size_t * out_error_len);
static kui_go_pfn_kui_script_engine_eval g_kui_script_engine_eval = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_execute_scripts)(KuiScriptEngine * engine);
static kui_go_pfn_kui_script_engine_execute_scripts g_kui_script_engine_execute_scripts = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_set_error_handler)(KuiScriptEngine * engine, KuiScriptErrorFn fn, void * user_data);
static kui_go_pfn_kui_script_engine_set_error_handler g_kui_script_engine_set_error_handler = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_script_engine)(KuiWindow * window, KuiScriptEngine * engine);
static kui_go_pfn_kui_window_set_script_engine g_kui_window_set_script_engine = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_client_create)(KuiHttpClient ** out_client);
static kui_go_pfn_kui_http_client_create g_kui_http_client_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_client_release)(KuiHttpClient * client);
static kui_go_pfn_kui_http_client_release g_kui_http_client_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_client_set_user_agent)(KuiHttpClient * client, const char * user_agent, size_t len);
static kui_go_pfn_kui_http_client_set_user_agent g_kui_http_client_set_user_agent = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_client_fetch_sync)(KuiHttpClient * client, const char * url, size_t len, double timeout_seconds, KuiHttpResponse ** out_response);
static kui_go_pfn_kui_http_client_fetch_sync g_kui_http_client_fetch_sync = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_client_fetch_async)(KuiHttpClient * client, const char * url, size_t len, double timeout_seconds, KuiHttpResponseFn fn, void * user_data);
static kui_go_pfn_kui_http_client_fetch_async g_kui_http_client_fetch_async = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_client_fetch_async_on_main_thread)(KuiHttpClient * client, const char * url, size_t len, double timeout_seconds, KuiHttpResponseFn fn, void * user_data);
static kui_go_pfn_kui_http_client_fetch_async_on_main_thread g_kui_http_client_fetch_async_on_main_thread = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_client_cancel_all)(KuiHttpClient * client);
static kui_go_pfn_kui_http_client_cancel_all g_kui_http_client_cancel_all = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_response_release)(KuiHttpResponse * response);
static kui_go_pfn_kui_http_response_release g_kui_http_response_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_response_get_status)(KuiHttpResponse * response, int * out_status_code);
static kui_go_pfn_kui_http_response_get_status g_kui_http_response_get_status = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_response_get_body)(KuiHttpResponse * response, uint8_t * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_http_response_get_body g_kui_http_response_get_body = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_response_get_content_type)(KuiHttpResponse * response, char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_http_response_get_content_type g_kui_http_response_get_content_type = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_http_response_get_error)(KuiHttpResponse * response, char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_http_response_get_error g_kui_http_response_get_error = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_init_extensions)(KuiDocument * document, const char * directory, size_t len, size_t * out_loaded_count);
static kui_go_pfn_kui_document_init_extensions g_kui_document_init_extensions = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_extension_is_loaded)(KuiDocument * document, const char * tag_name, size_t len, int * out_loaded);
static kui_go_pfn_kui_document_extension_is_loaded g_kui_document_extension_is_loaded = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_extension_invoke)(KuiDocument * document, KuiElement * element, const char * method, size_t method_len, const KuiCExtensionValue * args, size_t arg_count, size_t max_results, KuiExtensionResult ** out_result);
static kui_go_pfn_kui_document_extension_invoke g_kui_document_extension_invoke = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_extension_result_release)(KuiExtensionResult * result);
static kui_go_pfn_kui_extension_result_release g_kui_extension_result_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_extension_result_get_status)(KuiExtensionResult * result, int * out_status);
static kui_go_pfn_kui_extension_result_get_status g_kui_extension_result_get_status = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_extension_result_get_count)(KuiExtensionResult * result, size_t * out_count);
static kui_go_pfn_kui_extension_result_get_count g_kui_extension_result_get_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_extension_result_get_value)(KuiExtensionResult * result, size_t index, KuiCExtensionValue * out_value);
static kui_go_pfn_kui_extension_result_get_value g_kui_extension_result_get_value = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_extension_result_get_string)(KuiExtensionResult * result, size_t index, char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_extension_result_get_string g_kui_extension_result_get_string = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_extension_result_get_error)(KuiExtensionResult * result, char * buf, size_t capacity, size_t * out_len);
static kui_go_pfn_kui_extension_result_get_error g_kui_extension_result_get_error = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_extension_get_perf_stats)(KuiDocument * document, const char * tag_name, size_t len, KuiCExtensionPerfStats * out_stats);
static kui_go_pfn_kui_document_extension_get_perf_stats g_kui_document_extension_get_perf_stats = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_extension_reset_perf_stats)(KuiDocument * document);
static kui_go_pfn_kui_document_extension_reset_perf_stats g_kui_document_extension_reset_perf_stats = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_define_custom_element)(KuiDocument * document, const char * tag_name, size_t len, const KuiCCustomElementCallbacks * callbacks);
static kui_go_pfn_kui_document_define_custom_element g_kui_document_define_custom_element = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_custom_element_is_defined)(KuiDocument * document, const char * tag_name, size_t len, int * out_defined);
static kui_go_pfn_kui_document_custom_element_is_defined g_kui_document_custom_element_is_defined = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_print_to_pdf)(KuiDocument * document, const char * path, size_t path_len, const KuiCPDFOptions * options, int * out_written);
static kui_go_pfn_kui_document_print_to_pdf g_kui_document_print_to_pdf = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_focus)(KuiWindow * window);
static kui_go_pfn_kui_window_focus g_kui_window_focus = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_has_focus)(KuiWindow * window, int * out_focused);
static kui_go_pfn_kui_window_has_focus g_kui_window_has_focus = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_cursor)(KuiWindow * window, KuiCCursor cursor);
static kui_go_pfn_kui_window_set_cursor g_kui_window_set_cursor = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_cursor_visible)(KuiWindow * window, int visible);
static kui_go_pfn_kui_window_set_cursor_visible g_kui_window_set_cursor_visible = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_fullscreen)(KuiWindow * window, int fullscreen);
static kui_go_pfn_kui_window_set_fullscreen g_kui_window_set_fullscreen = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_is_fullscreen)(KuiWindow * window, int * out_fullscreen);
static kui_go_pfn_kui_window_is_fullscreen g_kui_window_is_fullscreen = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_frameless)(KuiWindow * window, int frameless);
static kui_go_pfn_kui_window_set_frameless g_kui_window_set_frameless = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_frameless_shadow)(KuiWindow * window, int shadow);
static kui_go_pfn_kui_window_set_frameless_shadow g_kui_window_set_frameless_shadow = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_frame_state)(KuiWindow * window, int * out_frameless, int * out_shadow, int * out_sizeable);
static kui_go_pfn_kui_window_get_frame_state g_kui_window_get_frame_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_frame_radii)(KuiWindow * window, float top_left, float top_right, float bottom_right, float bottom_left);
static kui_go_pfn_kui_window_set_frame_radii g_kui_window_set_frame_radii = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_frame_radii)(KuiWindow * window, float * out_top_left, float * out_top_right, float * out_bottom_right, float * out_bottom_left);
static kui_go_pfn_kui_window_get_frame_radii g_kui_window_get_frame_radii = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_ime_candidate_pos)(KuiWindow * window, float x, float y);
static kui_go_pfn_kui_window_set_ime_candidate_pos g_kui_window_set_ime_candidate_pos = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_insert_before)(KuiElement * parent, KuiElement * child, KuiElement * ref_child);
static kui_go_pfn_kui_element_insert_before g_kui_element_insert_before = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_replace_child)(KuiElement * parent, KuiElement * new_child, KuiElement * old_child);
static kui_go_pfn_kui_element_replace_child g_kui_element_replace_child = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_clone)(KuiElement * el, int deep, KuiElement ** out);
static kui_go_pfn_kui_element_clone g_kui_element_clone = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_matches)(KuiElement * el, const char * sel, size_t sel_len, int * out_matches);
static kui_go_pfn_kui_element_matches g_kui_element_matches = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_closest)(KuiElement * el, const char * sel, size_t sel_len, KuiElement ** out);
static kui_go_pfn_kui_element_closest g_kui_element_closest = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_outer_html)(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_outer_html g_kui_element_get_outer_html = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_computed_style_value)(KuiElement * el, const char * name, size_t name_len, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_element_get_computed_style_value g_kui_element_get_computed_style_value = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_insert_adjacent_html)(KuiElement * el, const char * position, size_t position_len, const char * html, size_t html_len);
static kui_go_pfn_kui_element_insert_adjacent_html g_kui_element_insert_adjacent_html = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_append_text)(KuiElement * el, const char * text, size_t len);
static kui_go_pfn_kui_element_append_text g_kui_element_append_text = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_has_focus)(KuiElement * el, int * out_focused);
static kui_go_pfn_kui_element_has_focus g_kui_element_has_focus = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_set_pointer_capture)(KuiElement * el, int pointer_id);
static kui_go_pfn_kui_element_set_pointer_capture g_kui_element_set_pointer_capture = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_release_pointer_capture)(KuiElement * el, int pointer_id);
static kui_go_pfn_kui_element_release_pointer_capture g_kui_element_release_pointer_capture = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_has_pointer_capture)(KuiElement * el, int pointer_id, int * out_captured);
static kui_go_pfn_kui_element_has_pointer_capture g_kui_element_has_pointer_capture = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_element_from_point)(KuiDocument * doc, float x, float y, KuiElement ** out);
static kui_go_pfn_kui_document_element_from_point g_kui_document_element_from_point = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_elements_by_class_name)(KuiDocument * doc, const char * class_name, size_t len, KuiElement ** out_buf, size_t buf_capacity, size_t * out_count);
static kui_go_pfn_kui_document_get_elements_by_class_name g_kui_document_get_elements_by_class_name = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_elements_by_tag_name)(KuiDocument * doc, const char * tag_name, size_t len, KuiElement ** out_buf, size_t buf_capacity, size_t * out_count);
static kui_go_pfn_kui_document_get_elements_by_tag_name g_kui_document_get_elements_by_tag_name = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_message_box)(const char * title, size_t title_len, const char * message, size_t message_len, KuiCMessageBoxType type, KuiCMessageBoxButtons buttons, KuiCMessageBoxResult * out_result);
static kui_go_pfn_kui_platform_message_box g_kui_platform_message_box = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_open_url)(const char * url, size_t len);
static kui_go_pfn_kui_platform_open_url g_kui_platform_open_url = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_now)(double * out_ms);
static kui_go_pfn_kui_platform_now g_kui_platform_now = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_monitor_from_point)(float x, float y, KuiCMonitorInfo * out_monitor, char * name_buf, size_t name_capacity, size_t * out_name_len);
static kui_go_pfn_kui_platform_get_monitor_from_point g_kui_platform_get_monitor_from_point = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_get_monitor_from_window)(void * native_window, KuiCMonitorInfo * out_monitor, char * name_buf, size_t name_capacity, size_t * out_name_len);
static kui_go_pfn_kui_platform_get_monitor_from_window g_kui_platform_get_monitor_from_window = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_menu_bar)(KuiWindow * window, const KuiCMenu * menus, size_t menu_count, KuiMenuActivateFn fn, void * user_data);
static kui_go_pfn_kui_window_set_menu_bar g_kui_window_set_menu_bar = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_clear_menu_bar)(KuiWindow * window);
static kui_go_pfn_kui_window_clear_menu_bar g_kui_window_clear_menu_bar = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_menu_set_checked)(KuiWindow * window, uint32_t item_id, int checked);
static kui_go_pfn_kui_window_menu_set_checked g_kui_window_menu_set_checked = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_menu_set_enabled)(KuiWindow * window, uint32_t item_id, int enabled);
static kui_go_pfn_kui_window_menu_set_enabled g_kui_window_menu_set_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_create)(KuiWindow * window, const char * icon_path, size_t icon_path_len, const char * tooltip, size_t tooltip_len, int visible);
static kui_go_pfn_kui_window_tray_create g_kui_window_tray_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_destroy)(KuiWindow * window);
static kui_go_pfn_kui_window_tray_destroy g_kui_window_tray_destroy = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_is_active)(KuiWindow * window, int * out_active);
static kui_go_pfn_kui_window_tray_is_active g_kui_window_tray_is_active = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_set_visible)(KuiWindow * window, int visible);
static kui_go_pfn_kui_window_tray_set_visible g_kui_window_tray_set_visible = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_is_visible)(KuiWindow * window, int * out_visible);
static kui_go_pfn_kui_window_tray_is_visible g_kui_window_tray_is_visible = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_set_tooltip)(KuiWindow * window, const char * tooltip, size_t len);
static kui_go_pfn_kui_window_tray_set_tooltip g_kui_window_tray_set_tooltip = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_set_icon)(KuiWindow * window, const char * icon_path, size_t len);
static kui_go_pfn_kui_window_tray_set_icon g_kui_window_tray_set_icon = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_set_context_menu)(KuiWindow * window, const KuiCMenuItem * items, size_t item_count, KuiMenuActivateFn fn, void * user_data);
static kui_go_pfn_kui_window_tray_set_context_menu g_kui_window_tray_set_context_menu = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_popup_context_menu)(KuiWindow * window);
static kui_go_pfn_kui_window_tray_popup_context_menu g_kui_window_tray_popup_context_menu = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_show_balloon)(KuiWindow * window, const char * title, size_t title_len, const char * message, size_t message_len, int icon_type, int duration_ms);
static kui_go_pfn_kui_window_tray_show_balloon g_kui_window_tray_show_balloon = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_set_event_callback)(KuiWindow * window, KuiTrayEventFn fn, void * user_data);
static kui_go_pfn_kui_window_tray_set_event_callback g_kui_window_tray_set_event_callback = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_tray_simulate_event)(KuiWindow * window, KuiCTrayEvent event);
static kui_go_pfn_kui_window_tray_simulate_event g_kui_window_tray_simulate_event = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_bind_shortcut)(KuiWindow * window, const char * shortcut, size_t len, KuiTaskFn fn, void * user_data, uint32_t * out_id);
static kui_go_pfn_kui_window_bind_shortcut g_kui_window_bind_shortcut = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_unbind_shortcut)(KuiWindow * window, uint32_t id);
static kui_go_pfn_kui_window_unbind_shortcut g_kui_window_unbind_shortcut = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_unbind_all_shortcuts)(KuiWindow * window);
static kui_go_pfn_kui_window_unbind_all_shortcuts g_kui_window_unbind_all_shortcuts = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_shortcuts_enabled)(KuiWindow * window, int enabled);
static kui_go_pfn_kui_window_set_shortcuts_enabled g_kui_window_set_shortcuts_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_show_toast)(KuiDocument * doc, const KuiCToastOptions * options, uint32_t * out_id);
static kui_go_pfn_kui_document_show_toast g_kui_document_show_toast = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_dismiss_toast)(KuiDocument * doc, uint32_t id);
static kui_go_pfn_kui_document_dismiss_toast g_kui_document_dismiss_toast = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_dismiss_all_toasts)(KuiDocument * doc);
static kui_go_pfn_kui_document_dismiss_all_toasts g_kui_document_dismiss_all_toasts = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_toast_count)(KuiDocument * doc, int * out_count);
static kui_go_pfn_kui_document_get_toast_count g_kui_document_get_toast_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_save_screenshot)(KuiWindow * window, const char * path, size_t path_len, int * out_saved);
static kui_go_pfn_kui_window_save_screenshot g_kui_window_save_screenshot = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_add_route)(KuiDocument * doc, const char * pattern, size_t pattern_len, const char * name, size_t name_len, KuiRouteViewFn fn, void * user_data, KuiTaskFn release_fn);
static kui_go_pfn_kui_document_router_add_route g_kui_document_router_add_route = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_remove_route)(KuiDocument * doc, const char * pattern, size_t len);
static kui_go_pfn_kui_document_router_remove_route g_kui_document_router_remove_route = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_clear_routes)(KuiDocument * doc);
static kui_go_pfn_kui_document_router_clear_routes g_kui_document_router_clear_routes = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_register_view)(KuiDocument * doc, KuiElement * el);
static kui_go_pfn_kui_document_router_register_view g_kui_document_router_register_view = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_unregister_view)(KuiDocument * doc, KuiElement * el);
static kui_go_pfn_kui_document_router_unregister_view g_kui_document_router_unregister_view = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_has_route)(KuiDocument * doc, const char * pattern, size_t len, int * out_has);
static kui_go_pfn_kui_document_router_has_route g_kui_document_router_has_route = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_get_route_count)(KuiDocument * doc, uint32_t * out_count);
static kui_go_pfn_kui_document_router_get_route_count g_kui_document_router_get_route_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_push)(KuiDocument * doc, const char * path, size_t len, int * out_navigated);
static kui_go_pfn_kui_document_router_push g_kui_document_router_push = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_replace)(KuiDocument * doc, const char * path, size_t len, int * out_navigated);
static kui_go_pfn_kui_document_router_replace g_kui_document_router_replace = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_go)(KuiDocument * doc, int delta, int * out_navigated);
static kui_go_pfn_kui_document_router_go g_kui_document_router_go = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_get_current)(KuiDocument * doc, KuiRouteInfoFn fn, void * user_data);
static kui_go_pfn_kui_document_router_get_current g_kui_document_router_get_current = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_get_current_path)(KuiDocument * doc, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_document_router_get_current_path g_kui_document_router_get_current_path = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_get_param)(KuiDocument * doc, const char * name, size_t name_len, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_document_router_get_param g_kui_document_router_get_param = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_get_history)(KuiDocument * doc, int * out_index, int * out_size);
static kui_go_pfn_kui_document_router_get_history g_kui_document_router_get_history = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_add_guard)(KuiDocument * doc, KuiCRouteGuardKind kind, KuiRouteGuardFn fn, void * user_data, KuiTaskFn release_fn, uint32_t * out_id);
static kui_go_pfn_kui_document_router_add_guard g_kui_document_router_add_guard = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_remove_guard)(KuiDocument * doc, uint32_t id);
static kui_go_pfn_kui_document_router_remove_guard g_kui_document_router_remove_guard = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_add_navigated_callback)(KuiDocument * doc, KuiRouteInfoFn fn, void * user_data, KuiTaskFn release_fn, uint32_t * out_id);
static kui_go_pfn_kui_document_router_add_navigated_callback g_kui_document_router_add_navigated_callback = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_router_remove_navigated_callback)(KuiDocument * doc, uint32_t id);
static kui_go_pfn_kui_document_router_remove_navigated_callback g_kui_document_router_remove_navigated_callback = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_create)(const char * initial, size_t len, KuiObservable ** out);
static kui_go_pfn_kui_observable_create g_kui_observable_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_retain)(KuiObservable * obs);
static kui_go_pfn_kui_observable_retain g_kui_observable_retain = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_release)(KuiObservable * obs);
static kui_go_pfn_kui_observable_release g_kui_observable_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_set)(KuiObservable * obs, const char * value, size_t len);
static kui_go_pfn_kui_observable_set g_kui_observable_set = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_get)(KuiObservable * obs, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_observable_get g_kui_observable_get = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_subscribe)(KuiObservable * obs, KuiTaskFn fn, void * user_data, KuiTaskFn release_fn, uint32_t * out_id);
static kui_go_pfn_kui_observable_subscribe g_kui_observable_subscribe = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_unsubscribe)(KuiObservable * obs, uint32_t id);
static kui_go_pfn_kui_observable_unsubscribe g_kui_observable_unsubscribe = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_bind)(KuiObservable * obs, KuiElement * el, const char * property, size_t property_len, int two_way, uint32_t * out_binding_id);
static kui_go_pfn_kui_observable_bind g_kui_observable_bind = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_unbind)(KuiObservable * obs, uint32_t binding_id);
static kui_go_pfn_kui_observable_unbind g_kui_observable_unbind = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observable_unbind_all)(KuiObservable * obs);
static kui_go_pfn_kui_observable_unbind_all g_kui_observable_unbind_all = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_flush_bindings)(KuiDocument * doc);
static kui_go_pfn_kui_document_flush_bindings g_kui_document_flush_bindings = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_get_animation_state)(KuiElement * el, int * out_animating, int * out_transitioning);
static kui_go_pfn_kui_element_get_animation_state g_kui_element_get_animation_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_animation_counts)(KuiDocument * doc, uint32_t * out_animations, uint32_t * out_transitions);
static kui_go_pfn_kui_document_get_animation_counts g_kui_document_get_animation_counts = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_mutation_observer_create)(KuiMutationFn fn, void * user_data, KuiTaskFn release_fn, KuiObserver ** out);
static kui_go_pfn_kui_mutation_observer_create g_kui_mutation_observer_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_resize_observer_create)(KuiResizeFn fn, void * user_data, KuiTaskFn release_fn, KuiObserver ** out);
static kui_go_pfn_kui_resize_observer_create g_kui_resize_observer_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_intersection_observer_create)(const KuiCIntersectionOptions * options, KuiIntersectionFn fn, void * user_data, KuiTaskFn release_fn, KuiObserver ** out);
static kui_go_pfn_kui_intersection_observer_create g_kui_intersection_observer_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_mutation_observer_observe)(KuiObserver * obs, KuiElement * target, const KuiCMutationOptions * options);
static kui_go_pfn_kui_mutation_observer_observe g_kui_mutation_observer_observe = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observer_observe)(KuiObserver * obs, KuiElement * target);
static kui_go_pfn_kui_observer_observe g_kui_observer_observe = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observer_unobserve)(KuiObserver * obs, KuiElement * target);
static kui_go_pfn_kui_observer_unobserve g_kui_observer_unobserve = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observer_disconnect)(KuiObserver * obs);
static kui_go_pfn_kui_observer_disconnect g_kui_observer_disconnect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observer_retain)(KuiObserver * obs);
static kui_go_pfn_kui_observer_retain g_kui_observer_retain = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_observer_release)(KuiObserver * obs);
static kui_go_pfn_kui_observer_release g_kui_observer_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_show_tooltip)(KuiDocument * doc, const char * text, size_t len, const KuiRect * anchor, KuiCPopupPosition position, uint32_t * out_id);
static kui_go_pfn_kui_document_show_tooltip g_kui_document_show_tooltip = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_show_context_menu)(KuiDocument * doc, float x, float y, const KuiCMenuItem * items, size_t item_count, KuiMenuActivateFn fn, void * user_data, KuiTaskFn release_fn, uint32_t * out_id);
static kui_go_pfn_kui_document_show_context_menu g_kui_document_show_context_menu = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_dismiss_popup)(KuiDocument * doc, uint32_t id);
static kui_go_pfn_kui_document_dismiss_popup g_kui_document_dismiss_popup = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_dismiss_all_popups)(KuiDocument * doc);
static kui_go_pfn_kui_document_dismiss_all_popups g_kui_document_dismiss_all_popups = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_dismiss_popups_of_type)(KuiDocument * doc, KuiCPopupType type);
static kui_go_pfn_kui_document_dismiss_popups_of_type g_kui_document_dismiss_popups_of_type = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_has_popups)(KuiDocument * doc, int * out_any);
static kui_go_pfn_kui_document_has_popups g_kui_document_has_popups = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_has_popup_of_type)(KuiDocument * doc, KuiCPopupType type, int * out_has);
static kui_go_pfn_kui_document_has_popup_of_type g_kui_document_has_popup_of_type = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_tooltip_delay)(KuiDocument * doc, double ms);
static kui_go_pfn_kui_document_set_tooltip_delay g_kui_document_set_tooltip_delay = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_tooltip_delay)(KuiDocument * doc, double * out_ms);
static kui_go_pfn_kui_document_get_tooltip_delay g_kui_document_get_tooltip_delay = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_tooltip_theme)(KuiDocument * doc, const KuiCTooltipTheme * theme);
static kui_go_pfn_kui_document_set_tooltip_theme g_kui_document_set_tooltip_theme = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_tooltip_theme)(KuiDocument * doc, KuiCTooltipTheme * out_theme);
static kui_go_pfn_kui_document_get_tooltip_theme g_kui_document_get_tooltip_theme = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_popup_mode)(KuiDocument * doc, KuiCPopupMode mode);
static kui_go_pfn_kui_document_set_popup_mode g_kui_document_set_popup_mode = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_default_context_menu_enabled)(KuiDocument * doc, int enabled);
static kui_go_pfn_kui_document_set_default_context_menu_enabled g_kui_document_set_default_context_menu_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_default_context_menu_enabled)(KuiDocument * doc, int * out_enabled);
static kui_go_pfn_kui_document_get_default_context_menu_enabled g_kui_document_get_default_context_menu_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_open_input_picker)(KuiElement * el, KuiCPickerMode mode);
static kui_go_pfn_kui_element_open_input_picker g_kui_element_open_input_picker = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_close_input_picker)(KuiDocument * doc);
static kui_go_pfn_kui_document_close_input_picker g_kui_document_close_input_picker = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_commit_input_picker)(KuiDocument * doc);
static kui_go_pfn_kui_document_commit_input_picker g_kui_document_commit_input_picker = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_input_picker_state)(KuiDocument * doc, KuiCInputPickerState * out_state);
static kui_go_pfn_kui_document_get_input_picker_state g_kui_document_get_input_picker_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_input_picker_target)(KuiDocument * doc, KuiElement ** out);
static kui_go_pfn_kui_document_get_input_picker_target g_kui_document_get_input_picker_target = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_input_picker_enabled)(KuiDocument * doc, int enabled);
static kui_go_pfn_kui_document_set_input_picker_enabled g_kui_document_set_input_picker_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_input_picker_month)(KuiDocument * doc, int year, int month);
static kui_go_pfn_kui_document_set_input_picker_month g_kui_document_set_input_picker_month = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_input_picker_color_hsv)(KuiDocument * doc, float h, float s, float v);
static kui_go_pfn_kui_document_set_input_picker_color_hsv g_kui_document_set_input_picker_color_hsv = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_input_picker_time)(KuiDocument * doc, int hour, int minute);
static kui_go_pfn_kui_document_set_input_picker_time g_kui_document_set_input_picker_time = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_is_available)(KuiDocument * doc, int * out_available);
static kui_go_pfn_kui_document_audio_is_available g_kui_document_audio_is_available = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_play)(KuiDocument * doc, const char * path, size_t len, const KuiCAudioPlayOptions * options, uint32_t * out_handle);
static kui_go_pfn_kui_document_audio_play g_kui_document_audio_play = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_stop)(KuiDocument * doc, uint32_t handle);
static kui_go_pfn_kui_document_audio_stop g_kui_document_audio_stop = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_set_paused)(KuiDocument * doc, uint32_t handle, int paused);
static kui_go_pfn_kui_document_audio_set_paused g_kui_document_audio_set_paused = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_set_volume)(KuiDocument * doc, uint32_t handle, float volume);
static kui_go_pfn_kui_document_audio_set_volume g_kui_document_audio_set_volume = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_set_pan)(KuiDocument * doc, uint32_t handle, float pan);
static kui_go_pfn_kui_document_audio_set_pan g_kui_document_audio_set_pan = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_set_pitch)(KuiDocument * doc, uint32_t handle, float pitch);
static kui_go_pfn_kui_document_audio_set_pitch g_kui_document_audio_set_pitch = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_seek)(KuiDocument * doc, uint32_t handle, double seconds);
static kui_go_pfn_kui_document_audio_seek g_kui_document_audio_seek = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_get_info)(KuiDocument * doc, uint32_t handle, KuiCSoundInfo * out_info);
static kui_go_pfn_kui_document_audio_get_info g_kui_document_audio_get_info = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_set_group)(KuiDocument * doc, KuiCAudioGroup group, float volume, int muted);
static kui_go_pfn_kui_document_audio_set_group g_kui_document_audio_set_group = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_get_group)(KuiDocument * doc, KuiCAudioGroup group, float * out_volume, int * out_muted);
static kui_go_pfn_kui_document_audio_get_group g_kui_document_audio_get_group = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_stop_all)(KuiDocument * doc, int group);
static kui_go_pfn_kui_document_audio_stop_all g_kui_document_audio_stop_all = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_set_all_paused)(KuiDocument * doc, int group, int paused);
static kui_go_pfn_kui_document_audio_set_all_paused g_kui_document_audio_set_all_paused = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_preload)(KuiDocument * doc, const char * path, size_t len, uint32_t count, uint32_t * out_loaded);
static kui_go_pfn_kui_document_audio_preload g_kui_document_audio_preload = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_unload)(KuiDocument * doc, const char * path, size_t len);
static kui_go_pfn_kui_document_audio_unload g_kui_document_audio_unload = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_audio_get_counts)(KuiDocument * doc, int group, uint32_t * out_active, uint32_t * out_pooled);
static kui_go_pfn_kui_document_audio_get_counts g_kui_document_audio_get_counts = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_load)(KuiElement * element);
static kui_go_pfn_kui_media_load g_kui_media_load = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_set_loop)(KuiElement * element, int loop);
static kui_go_pfn_kui_media_set_loop g_kui_media_set_loop = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_set_autoplay)(KuiElement * element, int autoplay);
static kui_go_pfn_kui_media_set_autoplay g_kui_media_set_autoplay = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_media_get_playback_flags)(KuiElement * element, int * out_loop, int * out_autoplay, int * out_ended);
static kui_go_pfn_kui_media_get_playback_flags g_kui_media_get_playback_flags = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_stop)(KuiElement * element);
static kui_go_pfn_kui_lottie_stop g_kui_lottie_stop = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_seek)(KuiElement * element, double position);
static kui_go_pfn_kui_lottie_seek g_kui_lottie_seek = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_set_speed)(KuiElement * element, float speed);
static kui_go_pfn_kui_lottie_set_speed g_kui_lottie_set_speed = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_set_loop)(KuiElement * element, int loop);
static kui_go_pfn_kui_lottie_set_loop g_kui_lottie_set_loop = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_lottie_get_state)(KuiElement * element, KuiCLottieState * out_state);
static kui_go_pfn_kui_lottie_get_state g_kui_lottie_get_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_load_extension)(KuiDocument * doc, const char * path, size_t len, int * out_loaded);
static kui_go_pfn_kui_document_load_extension g_kui_document_load_extension = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_unload_extension)(KuiDocument * doc, const char * tag_name, size_t len);
static kui_go_pfn_kui_document_unload_extension g_kui_document_unload_extension = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_reload_extension)(KuiDocument * doc, const char * path, size_t len, int * out_reloaded);
static kui_go_pfn_kui_document_reload_extension g_kui_document_reload_extension = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_unload_all_extensions)(KuiDocument * doc);
static kui_go_pfn_kui_document_unload_all_extensions g_kui_document_unload_all_extensions = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_watch_extensions)(KuiDocument * doc, const char * directory, size_t len);
static kui_go_pfn_kui_document_watch_extensions g_kui_document_watch_extensions = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_extension_state)(KuiDocument * doc, size_t * out_count, int * out_watching);
static kui_go_pfn_kui_document_get_extension_state g_kui_document_get_extension_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_stop_devtools)(KuiWindow * w);
static kui_go_pfn_kui_window_stop_devtools g_kui_window_stop_devtools = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_get_devtools_state)(KuiWindow * w, int * out_running, uint16_t * out_port, uint32_t * out_clients);
static kui_go_pfn_kui_window_get_devtools_state g_kui_window_get_devtools_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_set_data)(KuiEvent * event, const char * format, size_t format_len, const char * data, size_t data_len);
static kui_go_pfn_kui_event_drag_set_data g_kui_event_drag_set_data = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_get_data)(KuiEvent * event, const char * format, size_t format_len, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_event_drag_get_data g_kui_event_drag_get_data = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_clear_data)(KuiEvent * event, const char * format, size_t format_len);
static kui_go_pfn_kui_event_drag_clear_data g_kui_event_drag_clear_data = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_get_type_count)(KuiEvent * event, size_t * out_count);
static kui_go_pfn_kui_event_drag_get_type_count g_kui_event_drag_get_type_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_get_type)(KuiEvent * event, size_t index, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_event_drag_get_type g_kui_event_drag_get_type = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_set_drop_effect)(KuiEvent * event, KuiCDropEffect effect);
static kui_go_pfn_kui_event_drag_set_drop_effect g_kui_event_drag_set_drop_effect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_get_drop_effect)(KuiEvent * event, KuiCDropEffect * out_effect);
static kui_go_pfn_kui_event_drag_get_drop_effect g_kui_event_drag_get_drop_effect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_set_effect_allowed)(KuiEvent * event, KuiCEffectAllowed effect);
static kui_go_pfn_kui_event_drag_set_effect_allowed g_kui_event_drag_set_effect_allowed = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_drag_get_effect_allowed)(KuiEvent * event, KuiCEffectAllowed * out_effect);
static kui_go_pfn_kui_event_drag_get_effect_allowed g_kui_event_drag_get_effect_allowed = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_eval_value)(KuiScriptEngine * engine, const char * code, size_t code_len, const char * source, size_t source_len, int * out_success, KuiCScriptValue * out_value);
static kui_go_pfn_kui_script_engine_eval_value g_kui_script_engine_eval_value = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_eval_file)(KuiScriptEngine * engine, const char * path, size_t path_len, int * out_success, KuiCScriptValue * out_value);
static kui_go_pfn_kui_script_engine_eval_file g_kui_script_engine_eval_file = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_call_function)(KuiScriptEngine * engine, const char * name, size_t name_len, const KuiCScriptValue * args, size_t arg_count, int * out_success, KuiCScriptValue * out_value);
static kui_go_pfn_kui_script_engine_call_function g_kui_script_engine_call_function = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_get_last_error)(KuiScriptEngine * engine, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_script_engine_get_last_error g_kui_script_engine_get_last_error = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_register_function)(KuiScriptEngine * engine, const char * name, size_t name_len, KuiScriptFunctionFn fn, void * user_data, KuiTaskFn release_fn);
static kui_go_pfn_kui_script_engine_register_function g_kui_script_engine_register_function = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_set_global)(KuiScriptEngine * engine, const char * name, size_t name_len, const KuiCScriptValue * value);
static kui_go_pfn_kui_script_engine_set_global g_kui_script_engine_set_global = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_get_global)(KuiScriptEngine * engine, const char * name, size_t name_len, KuiCScriptValue * out_value);
static kui_go_pfn_kui_script_engine_get_global g_kui_script_engine_get_global = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_set_print_handler)(KuiScriptEngine * engine, KuiScriptPrintFn fn, void * user_data, KuiTaskFn release_fn);
static kui_go_pfn_kui_script_engine_set_print_handler g_kui_script_engine_set_print_handler = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_set_memory_limit)(KuiScriptEngine * engine, size_t bytes);
static kui_go_pfn_kui_script_engine_set_memory_limit g_kui_script_engine_set_memory_limit = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_set_instruction_limit)(KuiScriptEngine * engine, uint64_t max_instructions);
static kui_go_pfn_kui_script_engine_set_instruction_limit g_kui_script_engine_set_instruction_limit = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_get_resource_usage)(KuiScriptEngine * engine, size_t * out_memory_bytes, size_t * out_memory_limit, uint64_t * out_instruction_limit);
static kui_go_pfn_kui_script_engine_get_resource_usage g_kui_script_engine_get_resource_usage = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_function_create)(KuiScriptEngine * engine, KuiScriptFunctionFn fn, void * user_data, KuiTaskFn release_fn, KuiScriptFunction ** out_function);
static kui_go_pfn_kui_script_function_create g_kui_script_function_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_function_call)(KuiScriptFunction * function, const KuiCScriptValue * args, size_t arg_count, int * out_success, KuiCScriptValue * out_value);
static kui_go_pfn_kui_script_function_call g_kui_script_function_call = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_function_retain)(KuiScriptFunction * function, KuiScriptFunction ** out_owned);
static kui_go_pfn_kui_script_function_retain g_kui_script_function_retain = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_function_release)(KuiScriptFunction * function);
static kui_go_pfn_kui_script_function_release g_kui_script_function_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_define_custom_element_ex)(KuiDocument * document, const char * tag_name, size_t len, const KuiCCustomElementDefinition * definition);
static kui_go_pfn_kui_document_define_custom_element_ex g_kui_document_define_custom_element_ex = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_get_info)(KuiPaintContext * paint, float * out_width, float * out_height, float * out_device_pixel_ratio);
static kui_go_pfn_kui_paint_get_info g_kui_paint_get_info = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_fill_rect)(KuiPaintContext * paint, float x, float y, float width, float height, uint32_t rgba);
static kui_go_pfn_kui_paint_fill_rect g_kui_paint_fill_rect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_fill_rounded_rect)(KuiPaintContext * paint, float x, float y, float width, float height, const KuiCCornerRadii * radii, uint32_t rgba);
static kui_go_pfn_kui_paint_fill_rounded_rect g_kui_paint_fill_rounded_rect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_stroke_rect)(KuiPaintContext * paint, float x, float y, float width, float height, float line_width, uint32_t rgba);
static kui_go_pfn_kui_paint_stroke_rect g_kui_paint_stroke_rect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_stroke_rounded_rect)(KuiPaintContext * paint, float x, float y, float width, float height, const KuiCCornerRadii * radii, float line_width, uint32_t rgba);
static kui_go_pfn_kui_paint_stroke_rounded_rect g_kui_paint_stroke_rounded_rect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_draw_text)(KuiPaintContext * paint, const char * text, size_t len, float x, float baseline_y, const KuiCTextStyle * style, uint32_t rgba);
static kui_go_pfn_kui_paint_draw_text g_kui_paint_draw_text = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_draw_pixels)(KuiPaintContext * paint, const uint32_t * pixels, uint32_t width, uint32_t height, float x, float y, float dest_width, float dest_height);
static kui_go_pfn_kui_paint_draw_pixels g_kui_paint_draw_pixels = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_draw_image)(KuiPaintContext * paint, const char * path, size_t len, const KuiRect * src, const KuiRect * dest);
static kui_go_pfn_kui_paint_draw_image g_kui_paint_draw_image = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_draw_decoration_line)(KuiPaintContext * paint, float x, float y, float width, float thickness, KuiCDecorationStyle style, uint32_t rgba);
static kui_go_pfn_kui_paint_draw_decoration_line g_kui_paint_draw_decoration_line = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_save)(KuiPaintContext * paint);
static kui_go_pfn_kui_paint_save g_kui_paint_save = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_restore)(KuiPaintContext * paint);
static kui_go_pfn_kui_paint_restore g_kui_paint_restore = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_translate)(KuiPaintContext * paint, float dx, float dy);
static kui_go_pfn_kui_paint_translate g_kui_paint_translate = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_scale)(KuiPaintContext * paint, float sx, float sy);
static kui_go_pfn_kui_paint_scale g_kui_paint_scale = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_rotate)(KuiPaintContext * paint, float degrees);
static kui_go_pfn_kui_paint_rotate g_kui_paint_rotate = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_set_opacity)(KuiPaintContext * paint, float opacity);
static kui_go_pfn_kui_paint_set_opacity g_kui_paint_set_opacity = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_clip_rect)(KuiPaintContext * paint, float x, float y, float width, float height);
static kui_go_pfn_kui_paint_clip_rect g_kui_paint_clip_rect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_paint_clip_rounded_rect)(KuiPaintContext * paint, float x, float y, float width, float height, const KuiCCornerRadii * radii);
static kui_go_pfn_kui_paint_clip_rounded_rect g_kui_paint_clip_rounded_rect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_measure_text)(KuiElement * element, const char * text, size_t len, const KuiCTextStyle * style, KuiCTextMetrics * out_metrics);
static kui_go_pfn_kui_element_measure_text g_kui_element_measure_text = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_request_repaint)(KuiElement * element);
static kui_go_pfn_kui_element_request_repaint g_kui_element_request_repaint = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_invalidate_rect)(KuiElement * element, float x, float y, float width, float height);
static kui_go_pfn_kui_element_invalidate_rect g_kui_element_invalidate_rect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_devtools_dispatch_ime_event)(KuiDocument * document, KuiCImeEventType type, const char * text, size_t len);
static kui_go_pfn_kui_devtools_dispatch_ime_event g_kui_devtools_dispatch_ime_event = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_register_builtin_components)(KuiDocument * document);
static kui_go_pfn_kui_document_register_builtin_components g_kui_document_register_builtin_components = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_create)(KuiThemeManager ** out_manager);
static kui_go_pfn_kui_theme_manager_create g_kui_theme_manager_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_release)(KuiThemeManager * manager);
static kui_go_pfn_kui_theme_manager_release g_kui_theme_manager_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_register)(KuiThemeManager * manager, const char * name, size_t name_len, const KuiCKeyValue * variables, size_t variable_count);
static kui_go_pfn_kui_theme_manager_register g_kui_theme_manager_register = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_apply)(KuiThemeManager * manager, const char * name, size_t name_len, KuiDocument * document);
static kui_go_pfn_kui_theme_manager_apply g_kui_theme_manager_apply = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_get_active)(KuiThemeManager * manager, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_theme_manager_get_active g_kui_theme_manager_get_active = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_get_theme_count)(KuiThemeManager * manager, size_t * out_count);
static kui_go_pfn_kui_theme_manager_get_theme_count g_kui_theme_manager_get_theme_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_get_theme_name)(KuiThemeManager * manager, size_t index, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_theme_manager_get_theme_name g_kui_theme_manager_get_theme_name = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_get_variable_count)(KuiThemeManager * manager, const char * theme, size_t theme_len, size_t * out_count);
static kui_go_pfn_kui_theme_manager_get_variable_count g_kui_theme_manager_get_variable_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_theme_manager_get_variable)(KuiThemeManager * manager, const char * theme, size_t theme_len, size_t index, char * name_buf, size_t name_capacity, size_t * out_name_len, char * value_buf, size_t value_capacity, size_t * out_value_len);
static kui_go_pfn_kui_theme_manager_get_variable g_kui_theme_manager_get_variable = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_create)(KuiElement * container, const KuiCVirtualListConfig * config, KuiVirtualList ** out_list);
static kui_go_pfn_kui_virtual_list_create g_kui_virtual_list_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_release)(KuiVirtualList * list);
static kui_go_pfn_kui_virtual_list_release g_kui_virtual_list_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_get_state)(KuiVirtualList * list, KuiCVirtualListState * out_state);
static kui_go_pfn_kui_virtual_list_get_state g_kui_virtual_list_get_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_set_total_items)(KuiVirtualList * list, uint32_t count);
static kui_go_pfn_kui_virtual_list_set_total_items g_kui_virtual_list_set_total_items = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_insert_items)(KuiVirtualList * list, uint32_t index, uint32_t count);
static kui_go_pfn_kui_virtual_list_insert_items g_kui_virtual_list_insert_items = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_remove_items)(KuiVirtualList * list, uint32_t index, uint32_t count);
static kui_go_pfn_kui_virtual_list_remove_items g_kui_virtual_list_remove_items = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_refresh)(KuiVirtualList * list);
static kui_go_pfn_kui_virtual_list_refresh g_kui_virtual_list_refresh = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_scroll_to_index)(KuiVirtualList * list, uint32_t index);
static kui_go_pfn_kui_virtual_list_scroll_to_index g_kui_virtual_list_scroll_to_index = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_scroll_to_top)(KuiVirtualList * list);
static kui_go_pfn_kui_virtual_list_scroll_to_top g_kui_virtual_list_scroll_to_top = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_scroll_to_bottom)(KuiVirtualList * list);
static kui_go_pfn_kui_virtual_list_scroll_to_bottom g_kui_virtual_list_scroll_to_bottom = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_set_scroll_unit)(KuiVirtualList * list, KuiCScrollUnit unit);
static kui_go_pfn_kui_virtual_list_set_scroll_unit g_kui_virtual_list_set_scroll_unit = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_set_focused_index)(KuiVirtualList * list, int32_t index);
static kui_go_pfn_kui_virtual_list_set_focused_index g_kui_virtual_list_set_focused_index = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_virtual_list_move_focus)(KuiVirtualList * list, int32_t delta);
static kui_go_pfn_kui_virtual_list_move_focus g_kui_virtual_list_move_focus = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_create)(KuiElement * container, const KuiCListViewConfig * config, KuiListView ** out_list);
static kui_go_pfn_kui_list_view_create g_kui_list_view_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_release)(KuiListView * list);
static kui_go_pfn_kui_list_view_release g_kui_list_view_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_select)(KuiListView * list, uint32_t index);
static kui_go_pfn_kui_list_view_select g_kui_list_view_select = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_deselect)(KuiListView * list, uint32_t index);
static kui_go_pfn_kui_list_view_deselect g_kui_list_view_deselect = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_clear_selection)(KuiListView * list);
static kui_go_pfn_kui_list_view_clear_selection g_kui_list_view_clear_selection = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_select_range)(KuiListView * list, uint32_t start, uint32_t end);
static kui_go_pfn_kui_list_view_select_range g_kui_list_view_select_range = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_is_selected)(KuiListView * list, uint32_t index, int * out_selected);
static kui_go_pfn_kui_list_view_is_selected g_kui_list_view_is_selected = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_get_selected)(KuiListView * list, uint32_t * buf, size_t buf_capacity, size_t * out_count);
static kui_go_pfn_kui_list_view_get_selected g_kui_list_view_get_selected = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_get_focused_index)(KuiListView * list, int32_t * out_index);
static kui_go_pfn_kui_list_view_get_focused_index g_kui_list_view_get_focused_index = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_set_item_count)(KuiListView * list, uint32_t count);
static kui_go_pfn_kui_list_view_set_item_count g_kui_list_view_set_item_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_refresh)(KuiListView * list);
static kui_go_pfn_kui_list_view_refresh g_kui_list_view_refresh = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_scroll_to_index)(KuiListView * list, uint32_t index);
static kui_go_pfn_kui_list_view_scroll_to_index g_kui_list_view_scroll_to_index = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_scroll_to_top)(KuiListView * list);
static kui_go_pfn_kui_list_view_scroll_to_top g_kui_list_view_scroll_to_top = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_scroll_to_bottom)(KuiListView * list);
static kui_go_pfn_kui_list_view_scroll_to_bottom g_kui_list_view_scroll_to_bottom = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_set_scroll_unit)(KuiListView * list, KuiCScrollUnit unit);
static kui_go_pfn_kui_list_view_set_scroll_unit g_kui_list_view_set_scroll_unit = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_list_view_get_state)(KuiListView * list, KuiCVirtualListState * out_state);
static kui_go_pfn_kui_list_view_get_state g_kui_list_view_get_state = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_create)(KuiTreeView ** out_tree);
static kui_go_pfn_kui_tree_view_create g_kui_tree_view_create = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_release)(KuiTreeView * tree);
static kui_go_pfn_kui_tree_view_release g_kui_tree_view_release = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_add_node)(KuiTreeView * tree, const char * text, size_t text_len, const char * icon, size_t icon_len, uint32_t * out_index);
static kui_go_pfn_kui_tree_view_add_node g_kui_tree_view_add_node = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_add_child)(KuiTreeView * tree, uint32_t parent, const char * text, size_t text_len, const char * icon, size_t icon_len, uint32_t * out_index);
static kui_go_pfn_kui_tree_view_add_child g_kui_tree_view_add_child = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_get_node_count)(KuiTreeView * tree, uint32_t * out_count);
static kui_go_pfn_kui_tree_view_get_node_count g_kui_tree_view_get_node_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_get_node)(KuiTreeView * tree, uint32_t index, KuiCTreeNode * out_node);
static kui_go_pfn_kui_tree_view_get_node g_kui_tree_view_get_node = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_get_node_child)(KuiTreeView * tree, uint32_t index, size_t position, uint32_t * out_child);
static kui_go_pfn_kui_tree_view_get_node_child g_kui_tree_view_get_node_child = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_set_node)(KuiTreeView * tree, uint32_t index, const char * text, size_t text_len, const char * icon, size_t icon_len, int disabled);
static kui_go_pfn_kui_tree_view_set_node g_kui_tree_view_set_node = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_setup)(KuiTreeView * tree, KuiElement * container, const KuiCTreeViewConfig * config);
static kui_go_pfn_kui_tree_view_setup g_kui_tree_view_setup = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_expand)(KuiTreeView * tree, uint32_t index);
static kui_go_pfn_kui_tree_view_expand g_kui_tree_view_expand = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_collapse)(KuiTreeView * tree, uint32_t index);
static kui_go_pfn_kui_tree_view_collapse g_kui_tree_view_collapse = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_toggle)(KuiTreeView * tree, uint32_t index);
static kui_go_pfn_kui_tree_view_toggle g_kui_tree_view_toggle = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_expand_all)(KuiTreeView * tree);
static kui_go_pfn_kui_tree_view_expand_all g_kui_tree_view_expand_all = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_collapse_all)(KuiTreeView * tree);
static kui_go_pfn_kui_tree_view_collapse_all g_kui_tree_view_collapse_all = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_select)(KuiTreeView * tree, uint32_t index);
static kui_go_pfn_kui_tree_view_select g_kui_tree_view_select = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_clear_selection)(KuiTreeView * tree);
static kui_go_pfn_kui_tree_view_clear_selection g_kui_tree_view_clear_selection = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_toggle_select)(KuiTreeView * tree, uint32_t index);
static kui_go_pfn_kui_tree_view_toggle_select g_kui_tree_view_toggle_select = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_select_range)(KuiTreeView * tree, uint32_t index);
static kui_go_pfn_kui_tree_view_select_range g_kui_tree_view_select_range = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_select_all)(KuiTreeView * tree);
static kui_go_pfn_kui_tree_view_select_all g_kui_tree_view_select_all = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_is_selected)(KuiTreeView * tree, uint32_t index, int * out_selected);
static kui_go_pfn_kui_tree_view_is_selected g_kui_tree_view_is_selected = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_get_selected_index)(KuiTreeView * tree, int32_t * out_index);
static kui_go_pfn_kui_tree_view_get_selected_index g_kui_tree_view_get_selected_index = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_get_selected)(KuiTreeView * tree, uint32_t * buf, size_t buf_capacity, size_t * out_count);
static kui_go_pfn_kui_tree_view_get_selected g_kui_tree_view_get_selected = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_refresh)(KuiTreeView * tree);
static kui_go_pfn_kui_tree_view_refresh g_kui_tree_view_refresh = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_scroll_to_node)(KuiTreeView * tree, uint32_t index);
static kui_go_pfn_kui_tree_view_scroll_to_node g_kui_tree_view_scroll_to_node = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_set_scroll_unit)(KuiTreeView * tree, KuiCScrollUnit unit);
static kui_go_pfn_kui_tree_view_set_scroll_unit g_kui_tree_view_set_scroll_unit = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_tree_view_get_visible_row_count)(KuiTreeView * tree, uint32_t * out_count);
static kui_go_pfn_kui_tree_view_get_visible_row_count g_kui_tree_view_get_visible_row_count = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_style_profile)(KuiDocument * document, KuiCStyleProfile profile);
static kui_go_pfn_kui_document_set_style_profile g_kui_document_set_style_profile = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_style_profile)(KuiDocument * document, KuiCStyleProfile * out_profile);
static kui_go_pfn_kui_document_get_style_profile g_kui_document_get_style_profile = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_start_momentum_scroll)(KuiElement * el, float velocity_x, float velocity_y);
static kui_go_pfn_kui_element_start_momentum_scroll g_kui_element_start_momentum_scroll = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_element_scroll_into_view)(KuiElement * el, int align_to_top);
static kui_go_pfn_kui_element_scroll_into_view g_kui_element_scroll_into_view = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_event_wheel_get_delta)(KuiEvent * evt, float * out_delta_x, float * out_delta_y);
static kui_go_pfn_kui_event_wheel_get_delta g_kui_event_wheel_get_delta = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_script_engine_tick_timers)(KuiScriptEngine * engine, double now_ms);
static kui_go_pfn_kui_script_engine_tick_timers g_kui_script_engine_tick_timers = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_base_path)(KuiDocument * doc, const char * path, size_t path_len);
static kui_go_pfn_kui_document_set_base_path g_kui_document_set_base_path = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_get_base_path)(KuiDocument * doc, char * buf, size_t buf_capacity, size_t * out_len);
static kui_go_pfn_kui_document_get_base_path g_kui_document_get_base_path = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_document_set_render_device_pixel_ratio)(KuiDocument * doc, float dpr);
static kui_go_pfn_kui_document_set_render_device_pixel_ratio g_kui_document_set_render_device_pixel_ratio = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_create_ex)(const KuiWindowPropsEx * props, KuiWindow ** out);
static kui_go_pfn_kui_window_create_ex g_kui_window_create_ex = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_enabled)(KuiWindow * w, int enabled);
static kui_go_pfn_kui_window_set_enabled g_kui_window_set_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_is_enabled)(KuiWindow * w, int * out_enabled);
static kui_go_pfn_kui_window_is_enabled g_kui_window_is_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_process_frame)(KuiWindow * w);
static kui_go_pfn_kui_window_process_frame g_kui_window_process_frame = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_window_set_devtools_token_auth)(KuiWindow * w, int enabled);
static kui_go_pfn_kui_window_set_devtools_token_auth g_kui_window_set_devtools_token_auth = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_platform_disable_ime)(void);
static kui_go_pfn_kui_platform_disable_ime g_kui_platform_disable_ime = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_console_set_enabled)(int enabled);
static kui_go_pfn_kui_console_set_enabled g_kui_console_set_enabled = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_console_set_stderr_output)(int enabled);
static kui_go_pfn_kui_console_set_stderr_output g_kui_console_set_stderr_output = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_console_set_min_level)(KuiCLogLevel level);
static kui_go_pfn_kui_console_set_min_level g_kui_console_set_min_level = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_stack_trace_install)(const KuiCStackTraceConfig * config, int * out_installed);
static kui_go_pfn_kui_stack_trace_install g_kui_stack_trace_install = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_stack_trace_uninstall)(void);
static kui_go_pfn_kui_stack_trace_uninstall g_kui_stack_trace_uninstall = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_crash_reporter_install)(const KuiCCrashReporterConfig * config, int * out_installed);
static kui_go_pfn_kui_crash_reporter_install g_kui_crash_reporter_install = NULL;
typedef KuiStatus (__cdecl *kui_go_pfn_kui_crash_reporter_uninstall)(void);
static kui_go_pfn_kui_crash_reporter_uninstall g_kui_crash_reporter_uninstall = NULL;

int kui_go_load(const char* path) {
    wchar_t wide_path[32768];
    const wchar_t* requested = L"kui_c.dll";
    if (g_kui_module != NULL) return 1;
    g_kui_loader_error[0] = '\0';
    if (path != NULL && path[0] != '\0') {
        int count = MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS, path, -1,
                                        wide_path, (int)(sizeof(wide_path) / sizeof(wide_path[0])));
        if (count == 0) {
            snprintf(g_kui_loader_error, sizeof(g_kui_loader_error),
                     "invalid UTF-8 runtime path (win32=%lu)", GetLastError());
            return 0;
        }
        requested = wide_path;
    }
    g_kui_module = LoadLibraryExW(requested, NULL, LOAD_WITH_ALTERED_SEARCH_PATH);
    if (g_kui_module == NULL) {
        snprintf(g_kui_loader_error, sizeof(g_kui_loader_error),
                 "LoadLibraryExW failed (win32=%lu)", GetLastError());
        return 0;
    }
#define KUI_GO_RESOLVE(name) do { \
    g_##name = (kui_go_pfn_##name)(void*)GetProcAddress(g_kui_module, #name); \
    if (g_##name == NULL) { \
        snprintf(g_kui_loader_error, sizeof(g_kui_loader_error), \
                 "missing C ABI export: %s", #name); \
        FreeLibrary(g_kui_module); g_kui_module = NULL; return 0; \
    } \
} while (0)
    KUI_GO_RESOLVE(kui_get_version);
    KUI_GO_RESOLVE(kui_status_message);
    KUI_GO_RESOLVE(kui_platform_init);
    KUI_GO_RESOLVE(kui_platform_shutdown);
    KUI_GO_RESOLVE(kui_platform_run);
    KUI_GO_RESOLVE(kui_platform_quit);
    KUI_GO_RESOLVE(kui_platform_poll_events);
    KUI_GO_RESOLVE(kui_platform_wait_events);
    KUI_GO_RESOLVE(kui_platform_set_timeout);
    KUI_GO_RESOLVE(kui_platform_set_interval);
    KUI_GO_RESOLVE(kui_platform_clear_timer);
    KUI_GO_RESOLVE(kui_platform_request_idle);
    KUI_GO_RESOLVE(kui_platform_post_to_main_thread);
    KUI_GO_RESOLVE(kui_platform_set_f12_devtools_enabled);
    KUI_GO_RESOLVE(kui_platform_get_primary_monitor);
    KUI_GO_RESOLVE(kui_window_create);
    KUI_GO_RESOLVE(kui_window_release);
    KUI_GO_RESOLVE(kui_window_show);
    KUI_GO_RESOLVE(kui_window_hide);
    KUI_GO_RESOLVE(kui_window_is_visible);
    KUI_GO_RESOLVE(kui_window_get_document);
    KUI_GO_RESOLVE(kui_window_get_title);
    KUI_GO_RESOLVE(kui_window_set_title);
    KUI_GO_RESOLVE(kui_window_get_size);
    KUI_GO_RESOLVE(kui_window_resize);
    KUI_GO_RESOLVE(kui_window_get_position);
    KUI_GO_RESOLVE(kui_window_set_position);
    KUI_GO_RESOLVE(kui_window_minimize);
    KUI_GO_RESOLVE(kui_window_maximize);
    KUI_GO_RESOLVE(kui_window_restore);
    KUI_GO_RESOLVE(kui_window_set_always_on_top);
    KUI_GO_RESOLVE(kui_window_set_opacity);
    KUI_GO_RESOLVE(kui_window_get_device_pixel_ratio);
    KUI_GO_RESOLVE(kui_window_is_minimized);
    KUI_GO_RESOLVE(kui_window_is_maximized);
    KUI_GO_RESOLVE(kui_window_close);
    KUI_GO_RESOLVE(kui_window_should_close);
    KUI_GO_RESOLVE(kui_window_on_resize);
    KUI_GO_RESOLVE(kui_window_start_devtools);
    KUI_GO_RESOLVE(kui_window_get_devtools_url);
    KUI_GO_RESOLVE(kui_window_get_native_handle);
    KUI_GO_RESOLVE(kui_document_create);
    KUI_GO_RESOLVE(kui_document_retain);
    KUI_GO_RESOLVE(kui_document_release);
    KUI_GO_RESOLVE(kui_document_load_html);
    KUI_GO_RESOLVE(kui_document_load_html_file);
    KUI_GO_RESOLVE(kui_document_add_stylesheet);
    KUI_GO_RESOLVE(kui_document_clear_stylesheets);
    KUI_GO_RESOLVE(kui_document_root);
    KUI_GO_RESOLVE(kui_document_body);
    KUI_GO_RESOLVE(kui_document_head);
    KUI_GO_RESOLVE(kui_document_get_element_by_id);
    KUI_GO_RESOLVE(kui_document_query_selector);
    KUI_GO_RESOLVE(kui_document_query_selector_all);
    KUI_GO_RESOLVE(kui_document_create_element);
    KUI_GO_RESOLVE(kui_element_retain);
    KUI_GO_RESOLVE(kui_element_release);
    KUI_GO_RESOLVE(kui_element_get_tag_name);
    KUI_GO_RESOLVE(kui_element_get_id);
    KUI_GO_RESOLVE(kui_element_get_class_name);
    KUI_GO_RESOLVE(kui_element_set_class_name);
    KUI_GO_RESOLVE(kui_element_set_attribute);
    KUI_GO_RESOLVE(kui_element_get_attribute);
    KUI_GO_RESOLVE(kui_element_has_attribute);
    KUI_GO_RESOLVE(kui_element_remove_attribute);
    KUI_GO_RESOLVE(kui_element_get_value);
    KUI_GO_RESOLVE(kui_element_set_value);
    KUI_GO_RESOLVE(kui_element_get_checked);
    KUI_GO_RESOLVE(kui_element_set_checked);
    KUI_GO_RESOLVE(kui_element_get_selected_index);
    KUI_GO_RESOLVE(kui_element_set_selected_index);
    KUI_GO_RESOLVE(kui_element_set_text_content);
    KUI_GO_RESOLVE(kui_element_get_text_content);
    KUI_GO_RESOLVE(kui_element_set_inner_html);
    KUI_GO_RESOLVE(kui_element_get_inner_html);
    KUI_GO_RESOLVE(kui_element_class_list_add);
    KUI_GO_RESOLVE(kui_element_class_list_remove);
    KUI_GO_RESOLVE(kui_element_class_list_toggle);
    KUI_GO_RESOLVE(kui_element_class_list_contains);
    KUI_GO_RESOLVE(kui_element_query_selector);
    KUI_GO_RESOLVE(kui_element_query_selector_all);
    KUI_GO_RESOLVE(kui_element_parent);
    KUI_GO_RESOLVE(kui_element_first_child);
    KUI_GO_RESOLVE(kui_element_last_child);
    KUI_GO_RESOLVE(kui_element_next_sibling);
    KUI_GO_RESOLVE(kui_element_previous_sibling);
    KUI_GO_RESOLVE(kui_element_child_count);
    KUI_GO_RESOLVE(kui_element_append_child);
    KUI_GO_RESOLVE(kui_element_remove_child);
    KUI_GO_RESOLVE(kui_element_remove);
    KUI_GO_RESOLVE(kui_element_set_scroll_position);
    KUI_GO_RESOLVE(kui_element_scroll_by);
    KUI_GO_RESOLVE(kui_element_focus);
    KUI_GO_RESOLVE(kui_element_blur);
    KUI_GO_RESOLVE(kui_element_add_event_listener);
    KUI_GO_RESOLVE(kui_element_add_event_listener_ex);
    KUI_GO_RESOLVE(kui_element_remove_event_listener);
    KUI_GO_RESOLVE(kui_element_remove_event_listeners);
    KUI_GO_RESOLVE(kui_element_dispatch_event);
    KUI_GO_RESOLVE(kui_event_get_type);
    KUI_GO_RESOLVE(kui_event_get_target);
    KUI_GO_RESOLVE(kui_event_get_current_target);
    KUI_GO_RESOLVE(kui_event_get_related_target);
    KUI_GO_RESOLVE(kui_event_prevent_default);
    KUI_GO_RESOLVE(kui_event_stop_propagation);
    KUI_GO_RESOLVE(kui_event_get_flags);
    KUI_GO_RESOLVE(kui_event_mouse_get_client_xy);
    KUI_GO_RESOLVE(kui_event_mouse_get_page_xy);
    KUI_GO_RESOLVE(kui_event_mouse_get_offset_xy);
    KUI_GO_RESOLVE(kui_event_mouse_get_button);
    KUI_GO_RESOLVE(kui_event_key_get_key);
    KUI_GO_RESOLVE(kui_event_key_get_code);
    KUI_GO_RESOLVE(kui_event_key_get_modifiers);
    KUI_GO_RESOLVE(kui_element_set_style_property);
    KUI_GO_RESOLVE(kui_element_get_style_property);
    KUI_GO_RESOLVE(kui_element_remove_style_property);
    KUI_GO_RESOLVE(kui_element_style_property_is_important);
    KUI_GO_RESOLVE(kui_element_get_style_text);
    KUI_GO_RESOLVE(kui_element_set_style_text);
    KUI_GO_RESOLVE(kui_element_get_style_property_count);
    KUI_GO_RESOLVE(kui_element_get_style_property_name);
    KUI_GO_RESOLVE(kui_element_get_bounding_rect);
    KUI_GO_RESOLVE(kui_element_get_client_metrics);
    KUI_GO_RESOLVE(kui_element_get_offset_metrics);
    KUI_GO_RESOLVE(kui_element_get_scroll_metrics);
    KUI_GO_RESOLVE(kui_devtools_dispatch_mouse_event);
    KUI_GO_RESOLVE(kui_devtools_dispatch_key_event);
    KUI_GO_RESOLVE(kui_devtools_get_box_model);
    KUI_GO_RESOLVE(kui_devtools_dispatch_command);
    KUI_GO_RESOLVE(kui_platform_get_monitor_count);
    KUI_GO_RESOLVE(kui_platform_get_monitor);
    KUI_GO_RESOLVE(kui_platform_get_primary_monitor_physical_size);
    KUI_GO_RESOLVE(kui_platform_get_clipboard_text);
    KUI_GO_RESOLVE(kui_platform_set_clipboard_text);
    KUI_GO_RESOLVE(kui_platform_has_clipboard_text);
    KUI_GO_RESOLVE(kui_platform_get_os_name);
    KUI_GO_RESOLVE(kui_platform_get_os_version);
    KUI_GO_RESOLVE(kui_platform_open_file_dialog);
    KUI_GO_RESOLVE(kui_platform_save_file_dialog);
    KUI_GO_RESOLVE(kui_platform_select_folder_dialog);
    KUI_GO_RESOLVE(kui_window_set_frameless_sizeable);
    KUI_GO_RESOLVE(kui_window_set_frame_radius);
    KUI_GO_RESOLVE(kui_window_clear_frame_radius);
    KUI_GO_RESOLVE(kui_window_get_devtools_port);
    KUI_GO_RESOLVE(kui_window_add_event_listener);
    KUI_GO_RESOLVE(kui_window_remove_event_listener);
    KUI_GO_RESOLVE(kui_window_remove_event_listeners);
    KUI_GO_RESOLVE(kui_event_get_type_tag);
    KUI_GO_RESOLVE(kui_event_drag_get_file_count);
    KUI_GO_RESOLVE(kui_event_drag_get_file);
    KUI_GO_RESOLVE(kui_element_position_relative_to);
    KUI_GO_RESOLVE(kui_media_play);
    KUI_GO_RESOLVE(kui_media_pause);
    KUI_GO_RESOLVE(kui_media_set_src);
    KUI_GO_RESOLVE(kui_media_set_controls);
    KUI_GO_RESOLVE(kui_media_set_current_time);
    KUI_GO_RESOLVE(kui_media_get_current_time);
    KUI_GO_RESOLVE(kui_media_get_duration);
    KUI_GO_RESOLVE(kui_media_set_volume);
    KUI_GO_RESOLVE(kui_media_get_volume);
    KUI_GO_RESOLVE(kui_media_set_muted);
    KUI_GO_RESOLVE(kui_media_get_muted);
    KUI_GO_RESOLVE(kui_media_get_paused);
    KUI_GO_RESOLVE(kui_media_get_ready_state);
    KUI_GO_RESOLVE(kui_media_get_network_state);
    KUI_GO_RESOLVE(kui_media_is_available);
    KUI_GO_RESOLVE(kui_lottie_set_src);
    KUI_GO_RESOLVE(kui_lottie_play);
    KUI_GO_RESOLVE(kui_lottie_pause);
    KUI_GO_RESOLVE(kui_lottie_get_complete);
    KUI_GO_RESOLVE(kui_lottie_get_natural_size);
    KUI_GO_RESOLVE(kui_script_engine_create);
    KUI_GO_RESOLVE(kui_script_engine_release);
    KUI_GO_RESOLVE(kui_script_engine_bind_document);
    KUI_GO_RESOLVE(kui_script_engine_unbind_document);
    KUI_GO_RESOLVE(kui_script_engine_eval);
    KUI_GO_RESOLVE(kui_script_engine_execute_scripts);
    KUI_GO_RESOLVE(kui_script_engine_set_error_handler);
    KUI_GO_RESOLVE(kui_window_set_script_engine);
    KUI_GO_RESOLVE(kui_http_client_create);
    KUI_GO_RESOLVE(kui_http_client_release);
    KUI_GO_RESOLVE(kui_http_client_set_user_agent);
    KUI_GO_RESOLVE(kui_http_client_fetch_sync);
    KUI_GO_RESOLVE(kui_http_client_fetch_async);
    KUI_GO_RESOLVE(kui_http_client_fetch_async_on_main_thread);
    KUI_GO_RESOLVE(kui_http_client_cancel_all);
    KUI_GO_RESOLVE(kui_http_response_release);
    KUI_GO_RESOLVE(kui_http_response_get_status);
    KUI_GO_RESOLVE(kui_http_response_get_body);
    KUI_GO_RESOLVE(kui_http_response_get_content_type);
    KUI_GO_RESOLVE(kui_http_response_get_error);
    KUI_GO_RESOLVE(kui_document_init_extensions);
    KUI_GO_RESOLVE(kui_document_extension_is_loaded);
    KUI_GO_RESOLVE(kui_document_extension_invoke);
    KUI_GO_RESOLVE(kui_extension_result_release);
    KUI_GO_RESOLVE(kui_extension_result_get_status);
    KUI_GO_RESOLVE(kui_extension_result_get_count);
    KUI_GO_RESOLVE(kui_extension_result_get_value);
    KUI_GO_RESOLVE(kui_extension_result_get_string);
    KUI_GO_RESOLVE(kui_extension_result_get_error);
    KUI_GO_RESOLVE(kui_document_extension_get_perf_stats);
    KUI_GO_RESOLVE(kui_document_extension_reset_perf_stats);
    KUI_GO_RESOLVE(kui_document_define_custom_element);
    KUI_GO_RESOLVE(kui_document_custom_element_is_defined);
    KUI_GO_RESOLVE(kui_document_print_to_pdf);
    KUI_GO_RESOLVE(kui_window_focus);
    KUI_GO_RESOLVE(kui_window_has_focus);
    KUI_GO_RESOLVE(kui_window_set_cursor);
    KUI_GO_RESOLVE(kui_window_set_cursor_visible);
    KUI_GO_RESOLVE(kui_window_set_fullscreen);
    KUI_GO_RESOLVE(kui_window_is_fullscreen);
    KUI_GO_RESOLVE(kui_window_set_frameless);
    KUI_GO_RESOLVE(kui_window_set_frameless_shadow);
    KUI_GO_RESOLVE(kui_window_get_frame_state);
    KUI_GO_RESOLVE(kui_window_set_frame_radii);
    KUI_GO_RESOLVE(kui_window_get_frame_radii);
    KUI_GO_RESOLVE(kui_window_set_ime_candidate_pos);
    KUI_GO_RESOLVE(kui_element_insert_before);
    KUI_GO_RESOLVE(kui_element_replace_child);
    KUI_GO_RESOLVE(kui_element_clone);
    KUI_GO_RESOLVE(kui_element_matches);
    KUI_GO_RESOLVE(kui_element_closest);
    KUI_GO_RESOLVE(kui_element_get_outer_html);
    KUI_GO_RESOLVE(kui_element_get_computed_style_value);
    KUI_GO_RESOLVE(kui_element_insert_adjacent_html);
    KUI_GO_RESOLVE(kui_element_append_text);
    KUI_GO_RESOLVE(kui_element_has_focus);
    KUI_GO_RESOLVE(kui_element_set_pointer_capture);
    KUI_GO_RESOLVE(kui_element_release_pointer_capture);
    KUI_GO_RESOLVE(kui_element_has_pointer_capture);
    KUI_GO_RESOLVE(kui_document_element_from_point);
    KUI_GO_RESOLVE(kui_document_get_elements_by_class_name);
    KUI_GO_RESOLVE(kui_document_get_elements_by_tag_name);
    KUI_GO_RESOLVE(kui_platform_message_box);
    KUI_GO_RESOLVE(kui_platform_open_url);
    KUI_GO_RESOLVE(kui_platform_now);
    KUI_GO_RESOLVE(kui_platform_get_monitor_from_point);
    KUI_GO_RESOLVE(kui_platform_get_monitor_from_window);
    KUI_GO_RESOLVE(kui_window_set_menu_bar);
    KUI_GO_RESOLVE(kui_window_clear_menu_bar);
    KUI_GO_RESOLVE(kui_window_menu_set_checked);
    KUI_GO_RESOLVE(kui_window_menu_set_enabled);
    KUI_GO_RESOLVE(kui_window_tray_create);
    KUI_GO_RESOLVE(kui_window_tray_destroy);
    KUI_GO_RESOLVE(kui_window_tray_is_active);
    KUI_GO_RESOLVE(kui_window_tray_set_visible);
    KUI_GO_RESOLVE(kui_window_tray_is_visible);
    KUI_GO_RESOLVE(kui_window_tray_set_tooltip);
    KUI_GO_RESOLVE(kui_window_tray_set_icon);
    KUI_GO_RESOLVE(kui_window_tray_set_context_menu);
    KUI_GO_RESOLVE(kui_window_tray_popup_context_menu);
    KUI_GO_RESOLVE(kui_window_tray_show_balloon);
    KUI_GO_RESOLVE(kui_window_tray_set_event_callback);
    KUI_GO_RESOLVE(kui_window_tray_simulate_event);
    KUI_GO_RESOLVE(kui_window_bind_shortcut);
    KUI_GO_RESOLVE(kui_window_unbind_shortcut);
    KUI_GO_RESOLVE(kui_window_unbind_all_shortcuts);
    KUI_GO_RESOLVE(kui_window_set_shortcuts_enabled);
    KUI_GO_RESOLVE(kui_document_show_toast);
    KUI_GO_RESOLVE(kui_document_dismiss_toast);
    KUI_GO_RESOLVE(kui_document_dismiss_all_toasts);
    KUI_GO_RESOLVE(kui_document_get_toast_count);
    KUI_GO_RESOLVE(kui_window_save_screenshot);
    KUI_GO_RESOLVE(kui_document_router_add_route);
    KUI_GO_RESOLVE(kui_document_router_remove_route);
    KUI_GO_RESOLVE(kui_document_router_clear_routes);
    KUI_GO_RESOLVE(kui_document_router_register_view);
    KUI_GO_RESOLVE(kui_document_router_unregister_view);
    KUI_GO_RESOLVE(kui_document_router_has_route);
    KUI_GO_RESOLVE(kui_document_router_get_route_count);
    KUI_GO_RESOLVE(kui_document_router_push);
    KUI_GO_RESOLVE(kui_document_router_replace);
    KUI_GO_RESOLVE(kui_document_router_go);
    KUI_GO_RESOLVE(kui_document_router_get_current);
    KUI_GO_RESOLVE(kui_document_router_get_current_path);
    KUI_GO_RESOLVE(kui_document_router_get_param);
    KUI_GO_RESOLVE(kui_document_router_get_history);
    KUI_GO_RESOLVE(kui_document_router_add_guard);
    KUI_GO_RESOLVE(kui_document_router_remove_guard);
    KUI_GO_RESOLVE(kui_document_router_add_navigated_callback);
    KUI_GO_RESOLVE(kui_document_router_remove_navigated_callback);
    KUI_GO_RESOLVE(kui_observable_create);
    KUI_GO_RESOLVE(kui_observable_retain);
    KUI_GO_RESOLVE(kui_observable_release);
    KUI_GO_RESOLVE(kui_observable_set);
    KUI_GO_RESOLVE(kui_observable_get);
    KUI_GO_RESOLVE(kui_observable_subscribe);
    KUI_GO_RESOLVE(kui_observable_unsubscribe);
    KUI_GO_RESOLVE(kui_observable_bind);
    KUI_GO_RESOLVE(kui_observable_unbind);
    KUI_GO_RESOLVE(kui_observable_unbind_all);
    KUI_GO_RESOLVE(kui_document_flush_bindings);
    KUI_GO_RESOLVE(kui_element_get_animation_state);
    KUI_GO_RESOLVE(kui_document_get_animation_counts);
    KUI_GO_RESOLVE(kui_mutation_observer_create);
    KUI_GO_RESOLVE(kui_resize_observer_create);
    KUI_GO_RESOLVE(kui_intersection_observer_create);
    KUI_GO_RESOLVE(kui_mutation_observer_observe);
    KUI_GO_RESOLVE(kui_observer_observe);
    KUI_GO_RESOLVE(kui_observer_unobserve);
    KUI_GO_RESOLVE(kui_observer_disconnect);
    KUI_GO_RESOLVE(kui_observer_retain);
    KUI_GO_RESOLVE(kui_observer_release);
    KUI_GO_RESOLVE(kui_document_show_tooltip);
    KUI_GO_RESOLVE(kui_document_show_context_menu);
    KUI_GO_RESOLVE(kui_document_dismiss_popup);
    KUI_GO_RESOLVE(kui_document_dismiss_all_popups);
    KUI_GO_RESOLVE(kui_document_dismiss_popups_of_type);
    KUI_GO_RESOLVE(kui_document_has_popups);
    KUI_GO_RESOLVE(kui_document_has_popup_of_type);
    KUI_GO_RESOLVE(kui_document_set_tooltip_delay);
    KUI_GO_RESOLVE(kui_document_get_tooltip_delay);
    KUI_GO_RESOLVE(kui_document_set_tooltip_theme);
    KUI_GO_RESOLVE(kui_document_get_tooltip_theme);
    KUI_GO_RESOLVE(kui_document_set_popup_mode);
    KUI_GO_RESOLVE(kui_document_set_default_context_menu_enabled);
    KUI_GO_RESOLVE(kui_document_get_default_context_menu_enabled);
    KUI_GO_RESOLVE(kui_element_open_input_picker);
    KUI_GO_RESOLVE(kui_document_close_input_picker);
    KUI_GO_RESOLVE(kui_document_commit_input_picker);
    KUI_GO_RESOLVE(kui_document_get_input_picker_state);
    KUI_GO_RESOLVE(kui_document_get_input_picker_target);
    KUI_GO_RESOLVE(kui_document_set_input_picker_enabled);
    KUI_GO_RESOLVE(kui_document_set_input_picker_month);
    KUI_GO_RESOLVE(kui_document_set_input_picker_color_hsv);
    KUI_GO_RESOLVE(kui_document_set_input_picker_time);
    KUI_GO_RESOLVE(kui_document_audio_is_available);
    KUI_GO_RESOLVE(kui_document_audio_play);
    KUI_GO_RESOLVE(kui_document_audio_stop);
    KUI_GO_RESOLVE(kui_document_audio_set_paused);
    KUI_GO_RESOLVE(kui_document_audio_set_volume);
    KUI_GO_RESOLVE(kui_document_audio_set_pan);
    KUI_GO_RESOLVE(kui_document_audio_set_pitch);
    KUI_GO_RESOLVE(kui_document_audio_seek);
    KUI_GO_RESOLVE(kui_document_audio_get_info);
    KUI_GO_RESOLVE(kui_document_audio_set_group);
    KUI_GO_RESOLVE(kui_document_audio_get_group);
    KUI_GO_RESOLVE(kui_document_audio_stop_all);
    KUI_GO_RESOLVE(kui_document_audio_set_all_paused);
    KUI_GO_RESOLVE(kui_document_audio_preload);
    KUI_GO_RESOLVE(kui_document_audio_unload);
    KUI_GO_RESOLVE(kui_document_audio_get_counts);
    KUI_GO_RESOLVE(kui_media_load);
    KUI_GO_RESOLVE(kui_media_set_loop);
    KUI_GO_RESOLVE(kui_media_set_autoplay);
    KUI_GO_RESOLVE(kui_media_get_playback_flags);
    KUI_GO_RESOLVE(kui_lottie_stop);
    KUI_GO_RESOLVE(kui_lottie_seek);
    KUI_GO_RESOLVE(kui_lottie_set_speed);
    KUI_GO_RESOLVE(kui_lottie_set_loop);
    KUI_GO_RESOLVE(kui_lottie_get_state);
    KUI_GO_RESOLVE(kui_document_load_extension);
    KUI_GO_RESOLVE(kui_document_unload_extension);
    KUI_GO_RESOLVE(kui_document_reload_extension);
    KUI_GO_RESOLVE(kui_document_unload_all_extensions);
    KUI_GO_RESOLVE(kui_document_watch_extensions);
    KUI_GO_RESOLVE(kui_document_get_extension_state);
    KUI_GO_RESOLVE(kui_window_stop_devtools);
    KUI_GO_RESOLVE(kui_window_get_devtools_state);
    KUI_GO_RESOLVE(kui_event_drag_set_data);
    KUI_GO_RESOLVE(kui_event_drag_get_data);
    KUI_GO_RESOLVE(kui_event_drag_clear_data);
    KUI_GO_RESOLVE(kui_event_drag_get_type_count);
    KUI_GO_RESOLVE(kui_event_drag_get_type);
    KUI_GO_RESOLVE(kui_event_drag_set_drop_effect);
    KUI_GO_RESOLVE(kui_event_drag_get_drop_effect);
    KUI_GO_RESOLVE(kui_event_drag_set_effect_allowed);
    KUI_GO_RESOLVE(kui_event_drag_get_effect_allowed);
    KUI_GO_RESOLVE(kui_script_engine_eval_value);
    KUI_GO_RESOLVE(kui_script_engine_eval_file);
    KUI_GO_RESOLVE(kui_script_engine_call_function);
    KUI_GO_RESOLVE(kui_script_engine_get_last_error);
    KUI_GO_RESOLVE(kui_script_engine_register_function);
    KUI_GO_RESOLVE(kui_script_engine_set_global);
    KUI_GO_RESOLVE(kui_script_engine_get_global);
    KUI_GO_RESOLVE(kui_script_engine_set_print_handler);
    KUI_GO_RESOLVE(kui_script_engine_set_memory_limit);
    KUI_GO_RESOLVE(kui_script_engine_set_instruction_limit);
    KUI_GO_RESOLVE(kui_script_engine_get_resource_usage);
    KUI_GO_RESOLVE(kui_script_function_create);
    KUI_GO_RESOLVE(kui_script_function_call);
    KUI_GO_RESOLVE(kui_script_function_retain);
    KUI_GO_RESOLVE(kui_script_function_release);
    KUI_GO_RESOLVE(kui_document_define_custom_element_ex);
    KUI_GO_RESOLVE(kui_paint_get_info);
    KUI_GO_RESOLVE(kui_paint_fill_rect);
    KUI_GO_RESOLVE(kui_paint_fill_rounded_rect);
    KUI_GO_RESOLVE(kui_paint_stroke_rect);
    KUI_GO_RESOLVE(kui_paint_stroke_rounded_rect);
    KUI_GO_RESOLVE(kui_paint_draw_text);
    KUI_GO_RESOLVE(kui_paint_draw_pixels);
    KUI_GO_RESOLVE(kui_paint_draw_image);
    KUI_GO_RESOLVE(kui_paint_draw_decoration_line);
    KUI_GO_RESOLVE(kui_paint_save);
    KUI_GO_RESOLVE(kui_paint_restore);
    KUI_GO_RESOLVE(kui_paint_translate);
    KUI_GO_RESOLVE(kui_paint_scale);
    KUI_GO_RESOLVE(kui_paint_rotate);
    KUI_GO_RESOLVE(kui_paint_set_opacity);
    KUI_GO_RESOLVE(kui_paint_clip_rect);
    KUI_GO_RESOLVE(kui_paint_clip_rounded_rect);
    KUI_GO_RESOLVE(kui_element_measure_text);
    KUI_GO_RESOLVE(kui_element_request_repaint);
    KUI_GO_RESOLVE(kui_element_invalidate_rect);
    KUI_GO_RESOLVE(kui_devtools_dispatch_ime_event);
    KUI_GO_RESOLVE(kui_document_register_builtin_components);
    KUI_GO_RESOLVE(kui_theme_manager_create);
    KUI_GO_RESOLVE(kui_theme_manager_release);
    KUI_GO_RESOLVE(kui_theme_manager_register);
    KUI_GO_RESOLVE(kui_theme_manager_apply);
    KUI_GO_RESOLVE(kui_theme_manager_get_active);
    KUI_GO_RESOLVE(kui_theme_manager_get_theme_count);
    KUI_GO_RESOLVE(kui_theme_manager_get_theme_name);
    KUI_GO_RESOLVE(kui_theme_manager_get_variable_count);
    KUI_GO_RESOLVE(kui_theme_manager_get_variable);
    KUI_GO_RESOLVE(kui_virtual_list_create);
    KUI_GO_RESOLVE(kui_virtual_list_release);
    KUI_GO_RESOLVE(kui_virtual_list_get_state);
    KUI_GO_RESOLVE(kui_virtual_list_set_total_items);
    KUI_GO_RESOLVE(kui_virtual_list_insert_items);
    KUI_GO_RESOLVE(kui_virtual_list_remove_items);
    KUI_GO_RESOLVE(kui_virtual_list_refresh);
    KUI_GO_RESOLVE(kui_virtual_list_scroll_to_index);
    KUI_GO_RESOLVE(kui_virtual_list_scroll_to_top);
    KUI_GO_RESOLVE(kui_virtual_list_scroll_to_bottom);
    KUI_GO_RESOLVE(kui_virtual_list_set_scroll_unit);
    KUI_GO_RESOLVE(kui_virtual_list_set_focused_index);
    KUI_GO_RESOLVE(kui_virtual_list_move_focus);
    KUI_GO_RESOLVE(kui_list_view_create);
    KUI_GO_RESOLVE(kui_list_view_release);
    KUI_GO_RESOLVE(kui_list_view_select);
    KUI_GO_RESOLVE(kui_list_view_deselect);
    KUI_GO_RESOLVE(kui_list_view_clear_selection);
    KUI_GO_RESOLVE(kui_list_view_select_range);
    KUI_GO_RESOLVE(kui_list_view_is_selected);
    KUI_GO_RESOLVE(kui_list_view_get_selected);
    KUI_GO_RESOLVE(kui_list_view_get_focused_index);
    KUI_GO_RESOLVE(kui_list_view_set_item_count);
    KUI_GO_RESOLVE(kui_list_view_refresh);
    KUI_GO_RESOLVE(kui_list_view_scroll_to_index);
    KUI_GO_RESOLVE(kui_list_view_scroll_to_top);
    KUI_GO_RESOLVE(kui_list_view_scroll_to_bottom);
    KUI_GO_RESOLVE(kui_list_view_set_scroll_unit);
    KUI_GO_RESOLVE(kui_list_view_get_state);
    KUI_GO_RESOLVE(kui_tree_view_create);
    KUI_GO_RESOLVE(kui_tree_view_release);
    KUI_GO_RESOLVE(kui_tree_view_add_node);
    KUI_GO_RESOLVE(kui_tree_view_add_child);
    KUI_GO_RESOLVE(kui_tree_view_get_node_count);
    KUI_GO_RESOLVE(kui_tree_view_get_node);
    KUI_GO_RESOLVE(kui_tree_view_get_node_child);
    KUI_GO_RESOLVE(kui_tree_view_set_node);
    KUI_GO_RESOLVE(kui_tree_view_setup);
    KUI_GO_RESOLVE(kui_tree_view_expand);
    KUI_GO_RESOLVE(kui_tree_view_collapse);
    KUI_GO_RESOLVE(kui_tree_view_toggle);
    KUI_GO_RESOLVE(kui_tree_view_expand_all);
    KUI_GO_RESOLVE(kui_tree_view_collapse_all);
    KUI_GO_RESOLVE(kui_tree_view_select);
    KUI_GO_RESOLVE(kui_tree_view_clear_selection);
    KUI_GO_RESOLVE(kui_tree_view_toggle_select);
    KUI_GO_RESOLVE(kui_tree_view_select_range);
    KUI_GO_RESOLVE(kui_tree_view_select_all);
    KUI_GO_RESOLVE(kui_tree_view_is_selected);
    KUI_GO_RESOLVE(kui_tree_view_get_selected_index);
    KUI_GO_RESOLVE(kui_tree_view_get_selected);
    KUI_GO_RESOLVE(kui_tree_view_refresh);
    KUI_GO_RESOLVE(kui_tree_view_scroll_to_node);
    KUI_GO_RESOLVE(kui_tree_view_set_scroll_unit);
    KUI_GO_RESOLVE(kui_tree_view_get_visible_row_count);
    KUI_GO_RESOLVE(kui_document_set_style_profile);
    KUI_GO_RESOLVE(kui_document_get_style_profile);
    KUI_GO_RESOLVE(kui_element_start_momentum_scroll);
    KUI_GO_RESOLVE(kui_element_scroll_into_view);
    KUI_GO_RESOLVE(kui_event_wheel_get_delta);
    KUI_GO_RESOLVE(kui_script_engine_tick_timers);
    KUI_GO_RESOLVE(kui_document_set_base_path);
    KUI_GO_RESOLVE(kui_document_get_base_path);
    KUI_GO_RESOLVE(kui_document_set_render_device_pixel_ratio);
    KUI_GO_RESOLVE(kui_window_create_ex);
    KUI_GO_RESOLVE(kui_window_set_enabled);
    KUI_GO_RESOLVE(kui_window_is_enabled);
    KUI_GO_RESOLVE(kui_window_process_frame);
    KUI_GO_RESOLVE(kui_window_set_devtools_token_auth);
    KUI_GO_RESOLVE(kui_platform_disable_ime);
    KUI_GO_RESOLVE(kui_console_set_enabled);
    KUI_GO_RESOLVE(kui_console_set_stderr_output);
    KUI_GO_RESOLVE(kui_console_set_min_level);
    KUI_GO_RESOLVE(kui_stack_trace_install);
    KUI_GO_RESOLVE(kui_stack_trace_uninstall);
    KUI_GO_RESOLVE(kui_crash_reporter_install);
    KUI_GO_RESOLVE(kui_crash_reporter_uninstall);
#undef KUI_GO_RESOLVE
    return 1;
}

int kui_go_is_loaded(void) { return g_kui_module != NULL; }
const char* kui_go_loader_error(void) { return g_kui_loader_error; }

KuiVersion kui_get_version(void) {
    if (g_kui_get_version == NULL) { KuiVersion value = {0, 0, 0}; return value; }
    return g_kui_get_version();
}

const char * kui_status_message(KuiStatus s) {
    if (g_kui_status_message == NULL) { return NULL; }
    return g_kui_status_message(s);
}

KuiStatus kui_platform_init(void) {
    if (g_kui_platform_init == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_init();
}

KuiStatus kui_platform_shutdown(void) {
    if (g_kui_platform_shutdown == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_shutdown();
}

KuiStatus kui_platform_run(int * out_exit_code) {
    if (g_kui_platform_run == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_run(out_exit_code);
}

KuiStatus kui_platform_quit(int exit_code) {
    if (g_kui_platform_quit == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_quit(exit_code);
}

KuiStatus kui_platform_poll_events(void) {
    if (g_kui_platform_poll_events == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_poll_events();
}

KuiStatus kui_platform_wait_events(double timeout_seconds) {
    if (g_kui_platform_wait_events == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_wait_events(timeout_seconds);
}

KuiStatus kui_platform_set_timeout(KuiTaskFn fn, void * user_data, uint32_t delay_ms, KuiTimerId * out_id) {
    if (g_kui_platform_set_timeout == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_set_timeout(fn, user_data, delay_ms, out_id);
}

KuiStatus kui_platform_set_interval(KuiTaskFn fn, void * user_data, uint32_t interval_ms, KuiTimerId * out_id) {
    if (g_kui_platform_set_interval == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_set_interval(fn, user_data, interval_ms, out_id);
}

KuiStatus kui_platform_clear_timer(KuiTimerId id) {
    if (g_kui_platform_clear_timer == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_clear_timer(id);
}

KuiStatus kui_platform_request_idle(KuiTaskFn fn, void * user_data) {
    if (g_kui_platform_request_idle == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_request_idle(fn, user_data);
}

KuiStatus kui_platform_post_to_main_thread(KuiTaskFn fn, KuiTaskFn discard_fn, void * user_data) {
    if (g_kui_platform_post_to_main_thread == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_post_to_main_thread(fn, discard_fn, user_data);
}

KuiStatus kui_platform_set_f12_devtools_enabled(int enabled) {
    if (g_kui_platform_set_f12_devtools_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_set_f12_devtools_enabled(enabled);
}

KuiStatus kui_platform_get_primary_monitor(KuiCMonitorInfo * out_monitor) {
    if (g_kui_platform_get_primary_monitor == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_primary_monitor(out_monitor);
}

KuiStatus kui_window_create(const KuiWindowProps * props, KuiWindow ** out) {
    if (g_kui_window_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_create(props, out);
}

KuiStatus kui_window_release(KuiWindow * w) {
    if (g_kui_window_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_release(w);
}

KuiStatus kui_window_show(KuiWindow * w) {
    if (g_kui_window_show == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_show(w);
}

KuiStatus kui_window_hide(KuiWindow * w) {
    if (g_kui_window_hide == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_hide(w);
}

KuiStatus kui_window_is_visible(KuiWindow * w, int * out_visible) {
    if (g_kui_window_is_visible == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_is_visible(w, out_visible);
}

KuiStatus kui_window_get_document(KuiWindow * w, KuiDocument ** out) {
    if (g_kui_window_get_document == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_document(w, out);
}

KuiStatus kui_window_get_title(KuiWindow * w, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_window_get_title == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_title(w, buf, buf_capacity, out_len);
}

KuiStatus kui_window_set_title(KuiWindow * w, const char * title, size_t len) {
    if (g_kui_window_set_title == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_title(w, title, len);
}

KuiStatus kui_window_get_size(KuiWindow * w, int * out_width, int * out_height) {
    if (g_kui_window_get_size == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_size(w, out_width, out_height);
}

KuiStatus kui_window_resize(KuiWindow * w, int width, int height) {
    if (g_kui_window_resize == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_resize(w, width, height);
}

KuiStatus kui_window_get_position(KuiWindow * w, int * out_x, int * out_y) {
    if (g_kui_window_get_position == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_position(w, out_x, out_y);
}

KuiStatus kui_window_set_position(KuiWindow * w, int x, int y) {
    if (g_kui_window_set_position == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_position(w, x, y);
}

KuiStatus kui_window_minimize(KuiWindow * w) {
    if (g_kui_window_minimize == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_minimize(w);
}

KuiStatus kui_window_maximize(KuiWindow * w) {
    if (g_kui_window_maximize == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_maximize(w);
}

KuiStatus kui_window_restore(KuiWindow * w) {
    if (g_kui_window_restore == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_restore(w);
}

KuiStatus kui_window_set_always_on_top(KuiWindow * w, int on_top) {
    if (g_kui_window_set_always_on_top == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_always_on_top(w, on_top);
}

KuiStatus kui_window_set_opacity(KuiWindow * w, float opacity) {
    if (g_kui_window_set_opacity == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_opacity(w, opacity);
}

KuiStatus kui_window_get_device_pixel_ratio(KuiWindow * w, float * out_ratio) {
    if (g_kui_window_get_device_pixel_ratio == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_device_pixel_ratio(w, out_ratio);
}

KuiStatus kui_window_is_minimized(KuiWindow * w, int * out_minimized) {
    if (g_kui_window_is_minimized == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_is_minimized(w, out_minimized);
}

KuiStatus kui_window_is_maximized(KuiWindow * w, int * out_maximized) {
    if (g_kui_window_is_maximized == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_is_maximized(w, out_maximized);
}

KuiStatus kui_window_close(KuiWindow * w) {
    if (g_kui_window_close == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_close(w);
}

KuiStatus kui_window_should_close(KuiWindow * w, int * out_should_close) {
    if (g_kui_window_should_close == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_should_close(w, out_should_close);
}

KuiStatus kui_window_on_resize(KuiWindow * w, KuiWindowResizeFn fn, void * user_data) {
    if (g_kui_window_on_resize == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_on_resize(w, fn, user_data);
}

KuiStatus kui_window_start_devtools(KuiWindow * w, uint16_t port, int * out_started) {
    if (g_kui_window_start_devtools == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_start_devtools(w, port, out_started);
}

KuiStatus kui_window_get_devtools_url(KuiWindow * w, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_window_get_devtools_url == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_devtools_url(w, buf, buf_capacity, out_len);
}

KuiStatus kui_window_get_native_handle(KuiWindow * w, void ** out_handle) {
    if (g_kui_window_get_native_handle == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_native_handle(w, out_handle);
}

KuiStatus kui_document_create(KuiDocument ** out) {
    if (g_kui_document_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_create(out);
}

KuiStatus kui_document_retain(KuiDocument * doc) {
    if (g_kui_document_retain == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_retain(doc);
}

KuiStatus kui_document_release(KuiDocument * doc) {
    if (g_kui_document_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_release(doc);
}

KuiStatus kui_document_load_html(KuiDocument * doc, const char * html, size_t html_len) {
    if (g_kui_document_load_html == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_load_html(doc, html, html_len);
}

KuiStatus kui_document_load_html_file(KuiDocument * doc, const char * path, size_t path_len) {
    if (g_kui_document_load_html_file == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_load_html_file(doc, path, path_len);
}

KuiStatus kui_document_add_stylesheet(KuiDocument * doc, const char * css, size_t css_len) {
    if (g_kui_document_add_stylesheet == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_add_stylesheet(doc, css, css_len);
}

KuiStatus kui_document_clear_stylesheets(KuiDocument * doc) {
    if (g_kui_document_clear_stylesheets == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_clear_stylesheets(doc);
}

KuiStatus kui_document_root(KuiDocument * doc, KuiElement ** out) {
    if (g_kui_document_root == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_root(doc, out);
}

KuiStatus kui_document_body(KuiDocument * doc, KuiElement ** out) {
    if (g_kui_document_body == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_body(doc, out);
}

KuiStatus kui_document_head(KuiDocument * doc, KuiElement ** out) {
    if (g_kui_document_head == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_head(doc, out);
}

KuiStatus kui_document_get_element_by_id(KuiDocument * doc, const char * id, size_t id_len, KuiElement ** out) {
    if (g_kui_document_get_element_by_id == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_element_by_id(doc, id, id_len, out);
}

KuiStatus kui_document_query_selector(KuiDocument * doc, const char * sel, size_t sel_len, KuiElement ** out) {
    if (g_kui_document_query_selector == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_query_selector(doc, sel, sel_len, out);
}

KuiStatus kui_document_query_selector_all(KuiDocument * doc, const char * sel, size_t sel_len, KuiElement ** out_buf, size_t buf_capacity, size_t * out_count) {
    if (g_kui_document_query_selector_all == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_query_selector_all(doc, sel, sel_len, out_buf, buf_capacity, out_count);
}

KuiStatus kui_document_create_element(KuiDocument * doc, const char * tag, size_t tag_len, KuiElement ** out) {
    if (g_kui_document_create_element == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_create_element(doc, tag, tag_len, out);
}

KuiStatus kui_element_retain(KuiElement * el) {
    if (g_kui_element_retain == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_retain(el);
}

KuiStatus kui_element_release(KuiElement * el) {
    if (g_kui_element_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_release(el);
}

KuiStatus kui_element_get_tag_name(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_tag_name == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_tag_name(el, buf, buf_capacity, out_len);
}

KuiStatus kui_element_get_id(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_id == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_id(el, buf, buf_capacity, out_len);
}

KuiStatus kui_element_get_class_name(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_class_name == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_class_name(el, buf, buf_capacity, out_len);
}

KuiStatus kui_element_set_class_name(KuiElement * el, const char * value, size_t value_len) {
    if (g_kui_element_set_class_name == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_class_name(el, value, value_len);
}

KuiStatus kui_element_set_attribute(KuiElement * el, const char * name, size_t name_len, const char * value, size_t value_len) {
    if (g_kui_element_set_attribute == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_attribute(el, name, name_len, value, value_len);
}

KuiStatus kui_element_get_attribute(KuiElement * el, const char * name, size_t name_len, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_attribute == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_attribute(el, name, name_len, buf, buf_capacity, out_len);
}

KuiStatus kui_element_has_attribute(KuiElement * el, const char * name, size_t name_len, int * out_has) {
    if (g_kui_element_has_attribute == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_has_attribute(el, name, name_len, out_has);
}

KuiStatus kui_element_remove_attribute(KuiElement * el, const char * name, size_t name_len) {
    if (g_kui_element_remove_attribute == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_remove_attribute(el, name, name_len);
}

KuiStatus kui_element_get_value(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_value == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_value(el, buf, buf_capacity, out_len);
}

KuiStatus kui_element_set_value(KuiElement * el, const char * value, size_t value_len) {
    if (g_kui_element_set_value == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_value(el, value, value_len);
}

KuiStatus kui_element_get_checked(KuiElement * el, int * out_checked) {
    if (g_kui_element_get_checked == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_checked(el, out_checked);
}

KuiStatus kui_element_set_checked(KuiElement * el, int checked) {
    if (g_kui_element_set_checked == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_checked(el, checked);
}

KuiStatus kui_element_get_selected_index(KuiElement * el, int32_t * out_index) {
    if (g_kui_element_get_selected_index == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_selected_index(el, out_index);
}

KuiStatus kui_element_set_selected_index(KuiElement * el, int32_t index) {
    if (g_kui_element_set_selected_index == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_selected_index(el, index);
}

KuiStatus kui_element_set_text_content(KuiElement * el, const char * text, size_t text_len) {
    if (g_kui_element_set_text_content == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_text_content(el, text, text_len);
}

KuiStatus kui_element_get_text_content(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_text_content == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_text_content(el, buf, buf_capacity, out_len);
}

KuiStatus kui_element_set_inner_html(KuiElement * el, const char * html, size_t html_len) {
    if (g_kui_element_set_inner_html == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_inner_html(el, html, html_len);
}

KuiStatus kui_element_get_inner_html(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_inner_html == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_inner_html(el, buf, buf_capacity, out_len);
}

KuiStatus kui_element_class_list_add(KuiElement * el, const char * token, size_t len) {
    if (g_kui_element_class_list_add == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_class_list_add(el, token, len);
}

KuiStatus kui_element_class_list_remove(KuiElement * el, const char * token, size_t len) {
    if (g_kui_element_class_list_remove == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_class_list_remove(el, token, len);
}

KuiStatus kui_element_class_list_toggle(KuiElement * el, const char * token, size_t len) {
    if (g_kui_element_class_list_toggle == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_class_list_toggle(el, token, len);
}

KuiStatus kui_element_class_list_contains(KuiElement * el, const char * token, size_t len, int * out_contains) {
    if (g_kui_element_class_list_contains == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_class_list_contains(el, token, len, out_contains);
}

KuiStatus kui_element_query_selector(KuiElement * el, const char * sel, size_t sel_len, KuiElement ** out) {
    if (g_kui_element_query_selector == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_query_selector(el, sel, sel_len, out);
}

KuiStatus kui_element_query_selector_all(KuiElement * el, const char * sel, size_t sel_len, KuiElement ** out_buf, size_t buf_capacity, size_t * out_count) {
    if (g_kui_element_query_selector_all == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_query_selector_all(el, sel, sel_len, out_buf, buf_capacity, out_count);
}

KuiStatus kui_element_parent(KuiElement * el, KuiElement ** out) {
    if (g_kui_element_parent == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_parent(el, out);
}

KuiStatus kui_element_first_child(KuiElement * el, KuiElement ** out) {
    if (g_kui_element_first_child == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_first_child(el, out);
}

KuiStatus kui_element_last_child(KuiElement * el, KuiElement ** out) {
    if (g_kui_element_last_child == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_last_child(el, out);
}

KuiStatus kui_element_next_sibling(KuiElement * el, KuiElement ** out) {
    if (g_kui_element_next_sibling == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_next_sibling(el, out);
}

KuiStatus kui_element_previous_sibling(KuiElement * el, KuiElement ** out) {
    if (g_kui_element_previous_sibling == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_previous_sibling(el, out);
}

KuiStatus kui_element_child_count(KuiElement * el, size_t * out_count) {
    if (g_kui_element_child_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_child_count(el, out_count);
}

KuiStatus kui_element_append_child(KuiElement * parent, KuiElement * child) {
    if (g_kui_element_append_child == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_append_child(parent, child);
}

KuiStatus kui_element_remove_child(KuiElement * parent, KuiElement * child) {
    if (g_kui_element_remove_child == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_remove_child(parent, child);
}

KuiStatus kui_element_remove(KuiElement * el) {
    if (g_kui_element_remove == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_remove(el);
}

KuiStatus kui_element_set_scroll_position(KuiElement * el, float left, float top) {
    if (g_kui_element_set_scroll_position == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_scroll_position(el, left, top);
}

KuiStatus kui_element_scroll_by(KuiElement * el, float dx, float dy) {
    if (g_kui_element_scroll_by == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_scroll_by(el, dx, dy);
}

KuiStatus kui_element_focus(KuiElement * el) {
    if (g_kui_element_focus == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_focus(el);
}

KuiStatus kui_element_blur(KuiElement * el) {
    if (g_kui_element_blur == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_blur(el);
}

KuiStatus kui_element_add_event_listener(KuiElement * el, const char * event_name, size_t len, KuiEventListenerFn fn, void * user_data, KuiListenerId * out_id) {
    if (g_kui_element_add_event_listener == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_add_event_listener(el, event_name, len, fn, user_data, out_id);
}

KuiStatus kui_element_add_event_listener_ex(KuiElement * el, const char * event_name, size_t len, KuiEventListenerFn fn, void * user_data, int capture, KuiListenerId * out_id) {
    if (g_kui_element_add_event_listener_ex == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_add_event_listener_ex(el, event_name, len, fn, user_data, capture, out_id);
}

KuiStatus kui_element_remove_event_listener(KuiElement * el, const char * event_name, size_t len, KuiListenerId id) {
    if (g_kui_element_remove_event_listener == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_remove_event_listener(el, event_name, len, id);
}

KuiStatus kui_element_remove_event_listeners(KuiElement * el, const char * event_name, size_t len) {
    if (g_kui_element_remove_event_listeners == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_remove_event_listeners(el, event_name, len);
}

KuiStatus kui_element_dispatch_event(KuiElement * el, const char * event_name, size_t len, int bubbles, int cancelable, int * out_not_canceled) {
    if (g_kui_element_dispatch_event == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_dispatch_event(el, event_name, len, bubbles, cancelable, out_not_canceled);
}

KuiStatus kui_event_get_type(KuiEvent * evt, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_event_get_type == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_get_type(evt, buf, buf_capacity, out_len);
}

KuiStatus kui_event_get_target(KuiEvent * evt, KuiElement ** out) {
    if (g_kui_event_get_target == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_get_target(evt, out);
}

KuiStatus kui_event_get_current_target(KuiEvent * evt, KuiElement ** out) {
    if (g_kui_event_get_current_target == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_get_current_target(evt, out);
}

KuiStatus kui_event_get_related_target(KuiEvent * evt, KuiElement ** out) {
    if (g_kui_event_get_related_target == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_get_related_target(evt, out);
}

KuiStatus kui_event_prevent_default(KuiEvent * evt) {
    if (g_kui_event_prevent_default == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_prevent_default(evt);
}

KuiStatus kui_event_stop_propagation(KuiEvent * evt) {
    if (g_kui_event_stop_propagation == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_stop_propagation(evt);
}

KuiStatus kui_event_get_flags(KuiEvent * evt, int * out_bubbles, int * out_cancelable, int * out_default_prevented) {
    if (g_kui_event_get_flags == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_get_flags(evt, out_bubbles, out_cancelable, out_default_prevented);
}

KuiStatus kui_event_mouse_get_client_xy(KuiEvent * evt, float * x, float * y) {
    if (g_kui_event_mouse_get_client_xy == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_mouse_get_client_xy(evt, x, y);
}

KuiStatus kui_event_mouse_get_page_xy(KuiEvent * evt, float * x, float * y) {
    if (g_kui_event_mouse_get_page_xy == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_mouse_get_page_xy(evt, x, y);
}

KuiStatus kui_event_mouse_get_offset_xy(KuiEvent * evt, float * x, float * y) {
    if (g_kui_event_mouse_get_offset_xy == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_mouse_get_offset_xy(evt, x, y);
}

KuiStatus kui_event_mouse_get_button(KuiEvent * evt, int * out_button) {
    if (g_kui_event_mouse_get_button == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_mouse_get_button(evt, out_button);
}

KuiStatus kui_event_key_get_key(KuiEvent * evt, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_event_key_get_key == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_key_get_key(evt, buf, buf_capacity, out_len);
}

KuiStatus kui_event_key_get_code(KuiEvent * evt, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_event_key_get_code == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_key_get_code(evt, buf, buf_capacity, out_len);
}

KuiStatus kui_event_key_get_modifiers(KuiEvent * evt, int * out_ctrl, int * out_shift, int * out_alt, int * out_meta) {
    if (g_kui_event_key_get_modifiers == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_key_get_modifiers(evt, out_ctrl, out_shift, out_alt, out_meta);
}

KuiStatus kui_element_set_style_property(KuiElement * el, const char * prop, size_t prop_len, const char * value, size_t value_len, int important) {
    if (g_kui_element_set_style_property == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_style_property(el, prop, prop_len, value, value_len, important);
}

KuiStatus kui_element_get_style_property(KuiElement * el, const char * prop, size_t prop_len, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_style_property == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_style_property(el, prop, prop_len, buf, buf_capacity, out_len);
}

KuiStatus kui_element_remove_style_property(KuiElement * el, const char * prop, size_t prop_len) {
    if (g_kui_element_remove_style_property == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_remove_style_property(el, prop, prop_len);
}

KuiStatus kui_element_style_property_is_important(KuiElement * el, const char * prop, size_t prop_len, int * out_important) {
    if (g_kui_element_style_property_is_important == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_style_property_is_important(el, prop, prop_len, out_important);
}

KuiStatus kui_element_get_style_text(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_style_text == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_style_text(el, buf, buf_capacity, out_len);
}

KuiStatus kui_element_set_style_text(KuiElement * el, const char * css, size_t css_len) {
    if (g_kui_element_set_style_text == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_style_text(el, css, css_len);
}

KuiStatus kui_element_get_style_property_count(KuiElement * el, size_t * out_count) {
    if (g_kui_element_get_style_property_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_style_property_count(el, out_count);
}

KuiStatus kui_element_get_style_property_name(KuiElement * el, size_t index, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_style_property_name == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_style_property_name(el, index, buf, buf_capacity, out_len);
}

KuiStatus kui_element_get_bounding_rect(KuiElement * el, KuiRect * out) {
    if (g_kui_element_get_bounding_rect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_bounding_rect(el, out);
}

KuiStatus kui_element_get_client_metrics(KuiElement * el, float * out_left, float * out_top, float * out_width, float * out_height) {
    if (g_kui_element_get_client_metrics == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_client_metrics(el, out_left, out_top, out_width, out_height);
}

KuiStatus kui_element_get_offset_metrics(KuiElement * el, float * out_left, float * out_top, float * out_width, float * out_height) {
    if (g_kui_element_get_offset_metrics == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_offset_metrics(el, out_left, out_top, out_width, out_height);
}

KuiStatus kui_element_get_scroll_metrics(KuiElement * el, float * out_left, float * out_top, float * out_width, float * out_height) {
    if (g_kui_element_get_scroll_metrics == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_scroll_metrics(el, out_left, out_top, out_width, out_height);
}

KuiStatus kui_devtools_dispatch_mouse_event(KuiDocument * doc, KuiCMouseEventType type, float x, float y, KuiCMouseButton button, int modifiers, int click_count, float delta_x, float delta_y) {
    if (g_kui_devtools_dispatch_mouse_event == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_devtools_dispatch_mouse_event(doc, type, x, y, button, modifiers, click_count, delta_x, delta_y);
}

KuiStatus kui_devtools_dispatch_key_event(KuiDocument * doc, KuiCKeyEventType type, const char * key, size_t key_len, const char * code, size_t code_len, int modifiers, const char * text, size_t text_len, int is_auto_repeat) {
    if (g_kui_devtools_dispatch_key_event == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_devtools_dispatch_key_event(doc, type, key, key_len, code, code_len, modifiers, text, text_len, is_auto_repeat);
}

KuiStatus kui_devtools_get_box_model(KuiDocument * doc, const char * selector, float * out_quads, float * out_width, float * out_height) {
    if (g_kui_devtools_get_box_model == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_devtools_get_box_model(doc, selector, out_quads, out_width, out_height);
}

KuiStatus kui_devtools_dispatch_command(KuiDocument * doc, const char * json_command, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_devtools_dispatch_command == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_devtools_dispatch_command(doc, json_command, buf, buf_capacity, out_len);
}

KuiStatus kui_platform_get_monitor_count(size_t * out_count) {
    if (g_kui_platform_get_monitor_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_monitor_count(out_count);
}

KuiStatus kui_platform_get_monitor(size_t index, KuiCMonitorInfo * out_monitor, char * name_buf, size_t name_capacity, size_t * out_name_len) {
    if (g_kui_platform_get_monitor == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_monitor(index, out_monitor, name_buf, name_capacity, out_name_len);
}

KuiStatus kui_platform_get_primary_monitor_physical_size(float * out_width, float * out_height) {
    if (g_kui_platform_get_primary_monitor_physical_size == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_primary_monitor_physical_size(out_width, out_height);
}

KuiStatus kui_platform_get_clipboard_text(char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_platform_get_clipboard_text == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_clipboard_text(buf, capacity, out_len);
}

KuiStatus kui_platform_set_clipboard_text(const char * text, size_t len) {
    if (g_kui_platform_set_clipboard_text == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_set_clipboard_text(text, len);
}

KuiStatus kui_platform_has_clipboard_text(int * out_has_text) {
    if (g_kui_platform_has_clipboard_text == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_has_clipboard_text(out_has_text);
}

KuiStatus kui_platform_get_os_name(char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_platform_get_os_name == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_os_name(buf, capacity, out_len);
}

KuiStatus kui_platform_get_os_version(char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_platform_get_os_version == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_os_version(buf, capacity, out_len);
}

KuiStatus kui_platform_open_file_dialog(const KuiCFileDialogOptions * options, char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_platform_open_file_dialog == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_open_file_dialog(options, buf, capacity, out_len);
}

KuiStatus kui_platform_save_file_dialog(const KuiCFileDialogOptions * options, char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_platform_save_file_dialog == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_save_file_dialog(options, buf, capacity, out_len);
}

KuiStatus kui_platform_select_folder_dialog(const char * title, size_t title_len, char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_platform_select_folder_dialog == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_select_folder_dialog(title, title_len, buf, capacity, out_len);
}

KuiStatus kui_window_set_frameless_sizeable(KuiWindow * window, int enabled) {
    if (g_kui_window_set_frameless_sizeable == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_frameless_sizeable(window, enabled);
}

KuiStatus kui_window_set_frame_radius(KuiWindow * window, float radius) {
    if (g_kui_window_set_frame_radius == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_frame_radius(window, radius);
}

KuiStatus kui_window_clear_frame_radius(KuiWindow * window) {
    if (g_kui_window_clear_frame_radius == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_clear_frame_radius(window);
}

KuiStatus kui_window_get_devtools_port(KuiWindow * window, uint16_t * out_port) {
    if (g_kui_window_get_devtools_port == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_devtools_port(window, out_port);
}

KuiStatus kui_window_add_event_listener(KuiWindow * window, const char * event_name, size_t len, KuiWindowEventFn fn, void * user_data, int once, KuiListenerId * out_id) {
    if (g_kui_window_add_event_listener == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_add_event_listener(window, event_name, len, fn, user_data, once, out_id);
}

KuiStatus kui_window_remove_event_listener(KuiWindow * window, const char * event_name, size_t len, KuiListenerId id) {
    if (g_kui_window_remove_event_listener == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_remove_event_listener(window, event_name, len, id);
}

KuiStatus kui_window_remove_event_listeners(KuiWindow * window, const char * event_name, size_t len) {
    if (g_kui_window_remove_event_listeners == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_remove_event_listeners(window, event_name, len);
}

KuiStatus kui_event_get_type_tag(KuiEvent * event, KuiCEventTypeTag * out_tag) {
    if (g_kui_event_get_type_tag == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_get_type_tag(event, out_tag);
}

KuiStatus kui_event_drag_get_file_count(KuiEvent * event, size_t * out_count) {
    if (g_kui_event_drag_get_file_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_get_file_count(event, out_count);
}

KuiStatus kui_event_drag_get_file(KuiEvent * event, size_t index, char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_event_drag_get_file == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_get_file(event, index, buf, capacity, out_len);
}

KuiStatus kui_element_position_relative_to(KuiElement * element, KuiElement * anchor, const char * placement, size_t placement_len, const KuiCPositioningOptions * options) {
    if (g_kui_element_position_relative_to == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_position_relative_to(element, anchor, placement, placement_len, options);
}

KuiStatus kui_media_play(KuiElement * element) {
    if (g_kui_media_play == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_play(element);
}

KuiStatus kui_media_pause(KuiElement * element) {
    if (g_kui_media_pause == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_pause(element);
}

KuiStatus kui_media_set_src(KuiElement * element, const char * src, size_t len) {
    if (g_kui_media_set_src == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_set_src(element, src, len);
}

KuiStatus kui_media_set_controls(KuiElement * element, int enabled) {
    if (g_kui_media_set_controls == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_set_controls(element, enabled);
}

KuiStatus kui_media_set_current_time(KuiElement * element, double seconds) {
    if (g_kui_media_set_current_time == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_set_current_time(element, seconds);
}

KuiStatus kui_media_get_current_time(KuiElement * element, double * out_seconds) {
    if (g_kui_media_get_current_time == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_get_current_time(element, out_seconds);
}

KuiStatus kui_media_get_duration(KuiElement * element, double * out_seconds) {
    if (g_kui_media_get_duration == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_get_duration(element, out_seconds);
}

KuiStatus kui_media_set_volume(KuiElement * element, float volume) {
    if (g_kui_media_set_volume == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_set_volume(element, volume);
}

KuiStatus kui_media_get_volume(KuiElement * element, float * out_volume) {
    if (g_kui_media_get_volume == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_get_volume(element, out_volume);
}

KuiStatus kui_media_set_muted(KuiElement * element, int muted) {
    if (g_kui_media_set_muted == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_set_muted(element, muted);
}

KuiStatus kui_media_get_muted(KuiElement * element, int * out_muted) {
    if (g_kui_media_get_muted == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_get_muted(element, out_muted);
}

KuiStatus kui_media_get_paused(KuiElement * element, int * out_paused) {
    if (g_kui_media_get_paused == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_get_paused(element, out_paused);
}

KuiStatus kui_media_get_ready_state(KuiElement * element, int * out_state) {
    if (g_kui_media_get_ready_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_get_ready_state(element, out_state);
}

KuiStatus kui_media_get_network_state(KuiElement * element, int * out_state) {
    if (g_kui_media_get_network_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_get_network_state(element, out_state);
}

KuiStatus kui_media_is_available(KuiElement * element, int * out_available) {
    if (g_kui_media_is_available == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_is_available(element, out_available);
}

KuiStatus kui_lottie_set_src(KuiElement * element, const char * src, size_t len) {
    if (g_kui_lottie_set_src == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_set_src(element, src, len);
}

KuiStatus kui_lottie_play(KuiElement * element) {
    if (g_kui_lottie_play == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_play(element);
}

KuiStatus kui_lottie_pause(KuiElement * element) {
    if (g_kui_lottie_pause == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_pause(element);
}

KuiStatus kui_lottie_get_complete(KuiElement * element, int * out_complete) {
    if (g_kui_lottie_get_complete == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_get_complete(element, out_complete);
}

KuiStatus kui_lottie_get_natural_size(KuiElement * element, float * out_width, float * out_height) {
    if (g_kui_lottie_get_natural_size == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_get_natural_size(element, out_width, out_height);
}

KuiStatus kui_script_engine_create(KuiScriptEngine ** out_engine) {
    if (g_kui_script_engine_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_create(out_engine);
}

KuiStatus kui_script_engine_release(KuiScriptEngine * engine) {
    if (g_kui_script_engine_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_release(engine);
}

KuiStatus kui_script_engine_bind_document(KuiScriptEngine * engine, KuiDocument * document) {
    if (g_kui_script_engine_bind_document == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_bind_document(engine, document);
}

KuiStatus kui_script_engine_unbind_document(KuiScriptEngine * engine) {
    if (g_kui_script_engine_unbind_document == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_unbind_document(engine);
}

KuiStatus kui_script_engine_eval(KuiScriptEngine * engine, const char * code, size_t code_len, const char * source, size_t source_len, int * out_success, char * error_buf, size_t error_capacity, size_t * out_error_len) {
    if (g_kui_script_engine_eval == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_eval(engine, code, code_len, source, source_len, out_success, error_buf, error_capacity, out_error_len);
}

KuiStatus kui_script_engine_execute_scripts(KuiScriptEngine * engine) {
    if (g_kui_script_engine_execute_scripts == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_execute_scripts(engine);
}

KuiStatus kui_script_engine_set_error_handler(KuiScriptEngine * engine, KuiScriptErrorFn fn, void * user_data) {
    if (g_kui_script_engine_set_error_handler == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_set_error_handler(engine, fn, user_data);
}

KuiStatus kui_window_set_script_engine(KuiWindow * window, KuiScriptEngine * engine) {
    if (g_kui_window_set_script_engine == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_script_engine(window, engine);
}

KuiStatus kui_http_client_create(KuiHttpClient ** out_client) {
    if (g_kui_http_client_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_client_create(out_client);
}

KuiStatus kui_http_client_release(KuiHttpClient * client) {
    if (g_kui_http_client_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_client_release(client);
}

KuiStatus kui_http_client_set_user_agent(KuiHttpClient * client, const char * user_agent, size_t len) {
    if (g_kui_http_client_set_user_agent == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_client_set_user_agent(client, user_agent, len);
}

KuiStatus kui_http_client_fetch_sync(KuiHttpClient * client, const char * url, size_t len, double timeout_seconds, KuiHttpResponse ** out_response) {
    if (g_kui_http_client_fetch_sync == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_client_fetch_sync(client, url, len, timeout_seconds, out_response);
}

KuiStatus kui_http_client_fetch_async(KuiHttpClient * client, const char * url, size_t len, double timeout_seconds, KuiHttpResponseFn fn, void * user_data) {
    if (g_kui_http_client_fetch_async == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_client_fetch_async(client, url, len, timeout_seconds, fn, user_data);
}

KuiStatus kui_http_client_fetch_async_on_main_thread(KuiHttpClient * client, const char * url, size_t len, double timeout_seconds, KuiHttpResponseFn fn, void * user_data) {
    if (g_kui_http_client_fetch_async_on_main_thread == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_client_fetch_async_on_main_thread(client, url, len, timeout_seconds, fn, user_data);
}

KuiStatus kui_http_client_cancel_all(KuiHttpClient * client) {
    if (g_kui_http_client_cancel_all == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_client_cancel_all(client);
}

KuiStatus kui_http_response_release(KuiHttpResponse * response) {
    if (g_kui_http_response_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_response_release(response);
}

KuiStatus kui_http_response_get_status(KuiHttpResponse * response, int * out_status_code) {
    if (g_kui_http_response_get_status == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_response_get_status(response, out_status_code);
}

KuiStatus kui_http_response_get_body(KuiHttpResponse * response, uint8_t * buf, size_t capacity, size_t * out_len) {
    if (g_kui_http_response_get_body == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_response_get_body(response, buf, capacity, out_len);
}

KuiStatus kui_http_response_get_content_type(KuiHttpResponse * response, char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_http_response_get_content_type == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_response_get_content_type(response, buf, capacity, out_len);
}

KuiStatus kui_http_response_get_error(KuiHttpResponse * response, char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_http_response_get_error == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_http_response_get_error(response, buf, capacity, out_len);
}

KuiStatus kui_document_init_extensions(KuiDocument * document, const char * directory, size_t len, size_t * out_loaded_count) {
    if (g_kui_document_init_extensions == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_init_extensions(document, directory, len, out_loaded_count);
}

KuiStatus kui_document_extension_is_loaded(KuiDocument * document, const char * tag_name, size_t len, int * out_loaded) {
    if (g_kui_document_extension_is_loaded == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_extension_is_loaded(document, tag_name, len, out_loaded);
}

KuiStatus kui_document_extension_invoke(KuiDocument * document, KuiElement * element, const char * method, size_t method_len, const KuiCExtensionValue * args, size_t arg_count, size_t max_results, KuiExtensionResult ** out_result) {
    if (g_kui_document_extension_invoke == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_extension_invoke(document, element, method, method_len, args, arg_count, max_results, out_result);
}

KuiStatus kui_extension_result_release(KuiExtensionResult * result) {
    if (g_kui_extension_result_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_extension_result_release(result);
}

KuiStatus kui_extension_result_get_status(KuiExtensionResult * result, int * out_status) {
    if (g_kui_extension_result_get_status == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_extension_result_get_status(result, out_status);
}

KuiStatus kui_extension_result_get_count(KuiExtensionResult * result, size_t * out_count) {
    if (g_kui_extension_result_get_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_extension_result_get_count(result, out_count);
}

KuiStatus kui_extension_result_get_value(KuiExtensionResult * result, size_t index, KuiCExtensionValue * out_value) {
    if (g_kui_extension_result_get_value == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_extension_result_get_value(result, index, out_value);
}

KuiStatus kui_extension_result_get_string(KuiExtensionResult * result, size_t index, char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_extension_result_get_string == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_extension_result_get_string(result, index, buf, capacity, out_len);
}

KuiStatus kui_extension_result_get_error(KuiExtensionResult * result, char * buf, size_t capacity, size_t * out_len) {
    if (g_kui_extension_result_get_error == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_extension_result_get_error(result, buf, capacity, out_len);
}

KuiStatus kui_document_extension_get_perf_stats(KuiDocument * document, const char * tag_name, size_t len, KuiCExtensionPerfStats * out_stats) {
    if (g_kui_document_extension_get_perf_stats == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_extension_get_perf_stats(document, tag_name, len, out_stats);
}

KuiStatus kui_document_extension_reset_perf_stats(KuiDocument * document) {
    if (g_kui_document_extension_reset_perf_stats == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_extension_reset_perf_stats(document);
}

KuiStatus kui_document_define_custom_element(KuiDocument * document, const char * tag_name, size_t len, const KuiCCustomElementCallbacks * callbacks) {
    if (g_kui_document_define_custom_element == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_define_custom_element(document, tag_name, len, callbacks);
}

KuiStatus kui_document_custom_element_is_defined(KuiDocument * document, const char * tag_name, size_t len, int * out_defined) {
    if (g_kui_document_custom_element_is_defined == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_custom_element_is_defined(document, tag_name, len, out_defined);
}

KuiStatus kui_document_print_to_pdf(KuiDocument * document, const char * path, size_t path_len, const KuiCPDFOptions * options, int * out_written) {
    if (g_kui_document_print_to_pdf == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_print_to_pdf(document, path, path_len, options, out_written);
}

KuiStatus kui_window_focus(KuiWindow * window) {
    if (g_kui_window_focus == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_focus(window);
}

KuiStatus kui_window_has_focus(KuiWindow * window, int * out_focused) {
    if (g_kui_window_has_focus == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_has_focus(window, out_focused);
}

KuiStatus kui_window_set_cursor(KuiWindow * window, KuiCCursor cursor) {
    if (g_kui_window_set_cursor == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_cursor(window, cursor);
}

KuiStatus kui_window_set_cursor_visible(KuiWindow * window, int visible) {
    if (g_kui_window_set_cursor_visible == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_cursor_visible(window, visible);
}

KuiStatus kui_window_set_fullscreen(KuiWindow * window, int fullscreen) {
    if (g_kui_window_set_fullscreen == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_fullscreen(window, fullscreen);
}

KuiStatus kui_window_is_fullscreen(KuiWindow * window, int * out_fullscreen) {
    if (g_kui_window_is_fullscreen == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_is_fullscreen(window, out_fullscreen);
}

KuiStatus kui_window_set_frameless(KuiWindow * window, int frameless) {
    if (g_kui_window_set_frameless == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_frameless(window, frameless);
}

KuiStatus kui_window_set_frameless_shadow(KuiWindow * window, int shadow) {
    if (g_kui_window_set_frameless_shadow == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_frameless_shadow(window, shadow);
}

KuiStatus kui_window_get_frame_state(KuiWindow * window, int * out_frameless, int * out_shadow, int * out_sizeable) {
    if (g_kui_window_get_frame_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_frame_state(window, out_frameless, out_shadow, out_sizeable);
}

KuiStatus kui_window_set_frame_radii(KuiWindow * window, float top_left, float top_right, float bottom_right, float bottom_left) {
    if (g_kui_window_set_frame_radii == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_frame_radii(window, top_left, top_right, bottom_right, bottom_left);
}

KuiStatus kui_window_get_frame_radii(KuiWindow * window, float * out_top_left, float * out_top_right, float * out_bottom_right, float * out_bottom_left) {
    if (g_kui_window_get_frame_radii == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_frame_radii(window, out_top_left, out_top_right, out_bottom_right, out_bottom_left);
}

KuiStatus kui_window_set_ime_candidate_pos(KuiWindow * window, float x, float y) {
    if (g_kui_window_set_ime_candidate_pos == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_ime_candidate_pos(window, x, y);
}

KuiStatus kui_element_insert_before(KuiElement * parent, KuiElement * child, KuiElement * ref_child) {
    if (g_kui_element_insert_before == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_insert_before(parent, child, ref_child);
}

KuiStatus kui_element_replace_child(KuiElement * parent, KuiElement * new_child, KuiElement * old_child) {
    if (g_kui_element_replace_child == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_replace_child(parent, new_child, old_child);
}

KuiStatus kui_element_clone(KuiElement * el, int deep, KuiElement ** out) {
    if (g_kui_element_clone == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_clone(el, deep, out);
}

KuiStatus kui_element_matches(KuiElement * el, const char * sel, size_t sel_len, int * out_matches) {
    if (g_kui_element_matches == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_matches(el, sel, sel_len, out_matches);
}

KuiStatus kui_element_closest(KuiElement * el, const char * sel, size_t sel_len, KuiElement ** out) {
    if (g_kui_element_closest == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_closest(el, sel, sel_len, out);
}

KuiStatus kui_element_get_outer_html(KuiElement * el, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_outer_html == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_outer_html(el, buf, buf_capacity, out_len);
}

KuiStatus kui_element_get_computed_style_value(KuiElement * el, const char * name, size_t name_len, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_element_get_computed_style_value == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_computed_style_value(el, name, name_len, buf, buf_capacity, out_len);
}

KuiStatus kui_element_insert_adjacent_html(KuiElement * el, const char * position, size_t position_len, const char * html, size_t html_len) {
    if (g_kui_element_insert_adjacent_html == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_insert_adjacent_html(el, position, position_len, html, html_len);
}

KuiStatus kui_element_append_text(KuiElement * el, const char * text, size_t len) {
    if (g_kui_element_append_text == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_append_text(el, text, len);
}

KuiStatus kui_element_has_focus(KuiElement * el, int * out_focused) {
    if (g_kui_element_has_focus == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_has_focus(el, out_focused);
}

KuiStatus kui_element_set_pointer_capture(KuiElement * el, int pointer_id) {
    if (g_kui_element_set_pointer_capture == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_set_pointer_capture(el, pointer_id);
}

KuiStatus kui_element_release_pointer_capture(KuiElement * el, int pointer_id) {
    if (g_kui_element_release_pointer_capture == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_release_pointer_capture(el, pointer_id);
}

KuiStatus kui_element_has_pointer_capture(KuiElement * el, int pointer_id, int * out_captured) {
    if (g_kui_element_has_pointer_capture == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_has_pointer_capture(el, pointer_id, out_captured);
}

KuiStatus kui_document_element_from_point(KuiDocument * doc, float x, float y, KuiElement ** out) {
    if (g_kui_document_element_from_point == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_element_from_point(doc, x, y, out);
}

KuiStatus kui_document_get_elements_by_class_name(KuiDocument * doc, const char * class_name, size_t len, KuiElement ** out_buf, size_t buf_capacity, size_t * out_count) {
    if (g_kui_document_get_elements_by_class_name == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_elements_by_class_name(doc, class_name, len, out_buf, buf_capacity, out_count);
}

KuiStatus kui_document_get_elements_by_tag_name(KuiDocument * doc, const char * tag_name, size_t len, KuiElement ** out_buf, size_t buf_capacity, size_t * out_count) {
    if (g_kui_document_get_elements_by_tag_name == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_elements_by_tag_name(doc, tag_name, len, out_buf, buf_capacity, out_count);
}

KuiStatus kui_platform_message_box(const char * title, size_t title_len, const char * message, size_t message_len, KuiCMessageBoxType type, KuiCMessageBoxButtons buttons, KuiCMessageBoxResult * out_result) {
    if (g_kui_platform_message_box == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_message_box(title, title_len, message, message_len, type, buttons, out_result);
}

KuiStatus kui_platform_open_url(const char * url, size_t len) {
    if (g_kui_platform_open_url == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_open_url(url, len);
}

KuiStatus kui_platform_now(double * out_ms) {
    if (g_kui_platform_now == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_now(out_ms);
}

KuiStatus kui_platform_get_monitor_from_point(float x, float y, KuiCMonitorInfo * out_monitor, char * name_buf, size_t name_capacity, size_t * out_name_len) {
    if (g_kui_platform_get_monitor_from_point == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_monitor_from_point(x, y, out_monitor, name_buf, name_capacity, out_name_len);
}

KuiStatus kui_platform_get_monitor_from_window(void * native_window, KuiCMonitorInfo * out_monitor, char * name_buf, size_t name_capacity, size_t * out_name_len) {
    if (g_kui_platform_get_monitor_from_window == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_get_monitor_from_window(native_window, out_monitor, name_buf, name_capacity, out_name_len);
}

KuiStatus kui_window_set_menu_bar(KuiWindow * window, const KuiCMenu * menus, size_t menu_count, KuiMenuActivateFn fn, void * user_data) {
    if (g_kui_window_set_menu_bar == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_menu_bar(window, menus, menu_count, fn, user_data);
}

KuiStatus kui_window_clear_menu_bar(KuiWindow * window) {
    if (g_kui_window_clear_menu_bar == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_clear_menu_bar(window);
}

KuiStatus kui_window_menu_set_checked(KuiWindow * window, uint32_t item_id, int checked) {
    if (g_kui_window_menu_set_checked == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_menu_set_checked(window, item_id, checked);
}

KuiStatus kui_window_menu_set_enabled(KuiWindow * window, uint32_t item_id, int enabled) {
    if (g_kui_window_menu_set_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_menu_set_enabled(window, item_id, enabled);
}

KuiStatus kui_window_tray_create(KuiWindow * window, const char * icon_path, size_t icon_path_len, const char * tooltip, size_t tooltip_len, int visible) {
    if (g_kui_window_tray_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_create(window, icon_path, icon_path_len, tooltip, tooltip_len, visible);
}

KuiStatus kui_window_tray_destroy(KuiWindow * window) {
    if (g_kui_window_tray_destroy == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_destroy(window);
}

KuiStatus kui_window_tray_is_active(KuiWindow * window, int * out_active) {
    if (g_kui_window_tray_is_active == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_is_active(window, out_active);
}

KuiStatus kui_window_tray_set_visible(KuiWindow * window, int visible) {
    if (g_kui_window_tray_set_visible == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_set_visible(window, visible);
}

KuiStatus kui_window_tray_is_visible(KuiWindow * window, int * out_visible) {
    if (g_kui_window_tray_is_visible == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_is_visible(window, out_visible);
}

KuiStatus kui_window_tray_set_tooltip(KuiWindow * window, const char * tooltip, size_t len) {
    if (g_kui_window_tray_set_tooltip == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_set_tooltip(window, tooltip, len);
}

KuiStatus kui_window_tray_set_icon(KuiWindow * window, const char * icon_path, size_t len) {
    if (g_kui_window_tray_set_icon == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_set_icon(window, icon_path, len);
}

KuiStatus kui_window_tray_set_context_menu(KuiWindow * window, const KuiCMenuItem * items, size_t item_count, KuiMenuActivateFn fn, void * user_data) {
    if (g_kui_window_tray_set_context_menu == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_set_context_menu(window, items, item_count, fn, user_data);
}

KuiStatus kui_window_tray_popup_context_menu(KuiWindow * window) {
    if (g_kui_window_tray_popup_context_menu == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_popup_context_menu(window);
}

KuiStatus kui_window_tray_show_balloon(KuiWindow * window, const char * title, size_t title_len, const char * message, size_t message_len, int icon_type, int duration_ms) {
    if (g_kui_window_tray_show_balloon == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_show_balloon(window, title, title_len, message, message_len, icon_type, duration_ms);
}

KuiStatus kui_window_tray_set_event_callback(KuiWindow * window, KuiTrayEventFn fn, void * user_data) {
    if (g_kui_window_tray_set_event_callback == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_set_event_callback(window, fn, user_data);
}

KuiStatus kui_window_tray_simulate_event(KuiWindow * window, KuiCTrayEvent event) {
    if (g_kui_window_tray_simulate_event == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_tray_simulate_event(window, event);
}

KuiStatus kui_window_bind_shortcut(KuiWindow * window, const char * shortcut, size_t len, KuiTaskFn fn, void * user_data, uint32_t * out_id) {
    if (g_kui_window_bind_shortcut == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_bind_shortcut(window, shortcut, len, fn, user_data, out_id);
}

KuiStatus kui_window_unbind_shortcut(KuiWindow * window, uint32_t id) {
    if (g_kui_window_unbind_shortcut == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_unbind_shortcut(window, id);
}

KuiStatus kui_window_unbind_all_shortcuts(KuiWindow * window) {
    if (g_kui_window_unbind_all_shortcuts == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_unbind_all_shortcuts(window);
}

KuiStatus kui_window_set_shortcuts_enabled(KuiWindow * window, int enabled) {
    if (g_kui_window_set_shortcuts_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_shortcuts_enabled(window, enabled);
}

KuiStatus kui_document_show_toast(KuiDocument * doc, const KuiCToastOptions * options, uint32_t * out_id) {
    if (g_kui_document_show_toast == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_show_toast(doc, options, out_id);
}

KuiStatus kui_document_dismiss_toast(KuiDocument * doc, uint32_t id) {
    if (g_kui_document_dismiss_toast == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_dismiss_toast(doc, id);
}

KuiStatus kui_document_dismiss_all_toasts(KuiDocument * doc) {
    if (g_kui_document_dismiss_all_toasts == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_dismiss_all_toasts(doc);
}

KuiStatus kui_document_get_toast_count(KuiDocument * doc, int * out_count) {
    if (g_kui_document_get_toast_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_toast_count(doc, out_count);
}

KuiStatus kui_window_save_screenshot(KuiWindow * window, const char * path, size_t path_len, int * out_saved) {
    if (g_kui_window_save_screenshot == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_save_screenshot(window, path, path_len, out_saved);
}

KuiStatus kui_document_router_add_route(KuiDocument * doc, const char * pattern, size_t pattern_len, const char * name, size_t name_len, KuiRouteViewFn fn, void * user_data, KuiTaskFn release_fn) {
    if (g_kui_document_router_add_route == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_add_route(doc, pattern, pattern_len, name, name_len, fn, user_data, release_fn);
}

KuiStatus kui_document_router_remove_route(KuiDocument * doc, const char * pattern, size_t len) {
    if (g_kui_document_router_remove_route == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_remove_route(doc, pattern, len);
}

KuiStatus kui_document_router_clear_routes(KuiDocument * doc) {
    if (g_kui_document_router_clear_routes == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_clear_routes(doc);
}

KuiStatus kui_document_router_register_view(KuiDocument * doc, KuiElement * el) {
    if (g_kui_document_router_register_view == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_register_view(doc, el);
}

KuiStatus kui_document_router_unregister_view(KuiDocument * doc, KuiElement * el) {
    if (g_kui_document_router_unregister_view == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_unregister_view(doc, el);
}

KuiStatus kui_document_router_has_route(KuiDocument * doc, const char * pattern, size_t len, int * out_has) {
    if (g_kui_document_router_has_route == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_has_route(doc, pattern, len, out_has);
}

KuiStatus kui_document_router_get_route_count(KuiDocument * doc, uint32_t * out_count) {
    if (g_kui_document_router_get_route_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_get_route_count(doc, out_count);
}

KuiStatus kui_document_router_push(KuiDocument * doc, const char * path, size_t len, int * out_navigated) {
    if (g_kui_document_router_push == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_push(doc, path, len, out_navigated);
}

KuiStatus kui_document_router_replace(KuiDocument * doc, const char * path, size_t len, int * out_navigated) {
    if (g_kui_document_router_replace == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_replace(doc, path, len, out_navigated);
}

KuiStatus kui_document_router_go(KuiDocument * doc, int delta, int * out_navigated) {
    if (g_kui_document_router_go == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_go(doc, delta, out_navigated);
}

KuiStatus kui_document_router_get_current(KuiDocument * doc, KuiRouteInfoFn fn, void * user_data) {
    if (g_kui_document_router_get_current == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_get_current(doc, fn, user_data);
}

KuiStatus kui_document_router_get_current_path(KuiDocument * doc, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_document_router_get_current_path == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_get_current_path(doc, buf, buf_capacity, out_len);
}

KuiStatus kui_document_router_get_param(KuiDocument * doc, const char * name, size_t name_len, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_document_router_get_param == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_get_param(doc, name, name_len, buf, buf_capacity, out_len);
}

KuiStatus kui_document_router_get_history(KuiDocument * doc, int * out_index, int * out_size) {
    if (g_kui_document_router_get_history == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_get_history(doc, out_index, out_size);
}

KuiStatus kui_document_router_add_guard(KuiDocument * doc, KuiCRouteGuardKind kind, KuiRouteGuardFn fn, void * user_data, KuiTaskFn release_fn, uint32_t * out_id) {
    if (g_kui_document_router_add_guard == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_add_guard(doc, kind, fn, user_data, release_fn, out_id);
}

KuiStatus kui_document_router_remove_guard(KuiDocument * doc, uint32_t id) {
    if (g_kui_document_router_remove_guard == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_remove_guard(doc, id);
}

KuiStatus kui_document_router_add_navigated_callback(KuiDocument * doc, KuiRouteInfoFn fn, void * user_data, KuiTaskFn release_fn, uint32_t * out_id) {
    if (g_kui_document_router_add_navigated_callback == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_add_navigated_callback(doc, fn, user_data, release_fn, out_id);
}

KuiStatus kui_document_router_remove_navigated_callback(KuiDocument * doc, uint32_t id) {
    if (g_kui_document_router_remove_navigated_callback == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_router_remove_navigated_callback(doc, id);
}

KuiStatus kui_observable_create(const char * initial, size_t len, KuiObservable ** out) {
    if (g_kui_observable_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_create(initial, len, out);
}

KuiStatus kui_observable_retain(KuiObservable * obs) {
    if (g_kui_observable_retain == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_retain(obs);
}

KuiStatus kui_observable_release(KuiObservable * obs) {
    if (g_kui_observable_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_release(obs);
}

KuiStatus kui_observable_set(KuiObservable * obs, const char * value, size_t len) {
    if (g_kui_observable_set == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_set(obs, value, len);
}

KuiStatus kui_observable_get(KuiObservable * obs, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_observable_get == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_get(obs, buf, buf_capacity, out_len);
}

KuiStatus kui_observable_subscribe(KuiObservable * obs, KuiTaskFn fn, void * user_data, KuiTaskFn release_fn, uint32_t * out_id) {
    if (g_kui_observable_subscribe == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_subscribe(obs, fn, user_data, release_fn, out_id);
}

KuiStatus kui_observable_unsubscribe(KuiObservable * obs, uint32_t id) {
    if (g_kui_observable_unsubscribe == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_unsubscribe(obs, id);
}

KuiStatus kui_observable_bind(KuiObservable * obs, KuiElement * el, const char * property, size_t property_len, int two_way, uint32_t * out_binding_id) {
    if (g_kui_observable_bind == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_bind(obs, el, property, property_len, two_way, out_binding_id);
}

KuiStatus kui_observable_unbind(KuiObservable * obs, uint32_t binding_id) {
    if (g_kui_observable_unbind == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_unbind(obs, binding_id);
}

KuiStatus kui_observable_unbind_all(KuiObservable * obs) {
    if (g_kui_observable_unbind_all == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observable_unbind_all(obs);
}

KuiStatus kui_document_flush_bindings(KuiDocument * doc) {
    if (g_kui_document_flush_bindings == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_flush_bindings(doc);
}

KuiStatus kui_element_get_animation_state(KuiElement * el, int * out_animating, int * out_transitioning) {
    if (g_kui_element_get_animation_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_get_animation_state(el, out_animating, out_transitioning);
}

KuiStatus kui_document_get_animation_counts(KuiDocument * doc, uint32_t * out_animations, uint32_t * out_transitions) {
    if (g_kui_document_get_animation_counts == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_animation_counts(doc, out_animations, out_transitions);
}

KuiStatus kui_mutation_observer_create(KuiMutationFn fn, void * user_data, KuiTaskFn release_fn, KuiObserver ** out) {
    if (g_kui_mutation_observer_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_mutation_observer_create(fn, user_data, release_fn, out);
}

KuiStatus kui_resize_observer_create(KuiResizeFn fn, void * user_data, KuiTaskFn release_fn, KuiObserver ** out) {
    if (g_kui_resize_observer_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_resize_observer_create(fn, user_data, release_fn, out);
}

KuiStatus kui_intersection_observer_create(const KuiCIntersectionOptions * options, KuiIntersectionFn fn, void * user_data, KuiTaskFn release_fn, KuiObserver ** out) {
    if (g_kui_intersection_observer_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_intersection_observer_create(options, fn, user_data, release_fn, out);
}

KuiStatus kui_mutation_observer_observe(KuiObserver * obs, KuiElement * target, const KuiCMutationOptions * options) {
    if (g_kui_mutation_observer_observe == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_mutation_observer_observe(obs, target, options);
}

KuiStatus kui_observer_observe(KuiObserver * obs, KuiElement * target) {
    if (g_kui_observer_observe == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observer_observe(obs, target);
}

KuiStatus kui_observer_unobserve(KuiObserver * obs, KuiElement * target) {
    if (g_kui_observer_unobserve == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observer_unobserve(obs, target);
}

KuiStatus kui_observer_disconnect(KuiObserver * obs) {
    if (g_kui_observer_disconnect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observer_disconnect(obs);
}

KuiStatus kui_observer_retain(KuiObserver * obs) {
    if (g_kui_observer_retain == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observer_retain(obs);
}

KuiStatus kui_observer_release(KuiObserver * obs) {
    if (g_kui_observer_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_observer_release(obs);
}

KuiStatus kui_document_show_tooltip(KuiDocument * doc, const char * text, size_t len, const KuiRect * anchor, KuiCPopupPosition position, uint32_t * out_id) {
    if (g_kui_document_show_tooltip == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_show_tooltip(doc, text, len, anchor, position, out_id);
}

KuiStatus kui_document_show_context_menu(KuiDocument * doc, float x, float y, const KuiCMenuItem * items, size_t item_count, KuiMenuActivateFn fn, void * user_data, KuiTaskFn release_fn, uint32_t * out_id) {
    if (g_kui_document_show_context_menu == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_show_context_menu(doc, x, y, items, item_count, fn, user_data, release_fn, out_id);
}

KuiStatus kui_document_dismiss_popup(KuiDocument * doc, uint32_t id) {
    if (g_kui_document_dismiss_popup == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_dismiss_popup(doc, id);
}

KuiStatus kui_document_dismiss_all_popups(KuiDocument * doc) {
    if (g_kui_document_dismiss_all_popups == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_dismiss_all_popups(doc);
}

KuiStatus kui_document_dismiss_popups_of_type(KuiDocument * doc, KuiCPopupType type) {
    if (g_kui_document_dismiss_popups_of_type == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_dismiss_popups_of_type(doc, type);
}

KuiStatus kui_document_has_popups(KuiDocument * doc, int * out_any) {
    if (g_kui_document_has_popups == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_has_popups(doc, out_any);
}

KuiStatus kui_document_has_popup_of_type(KuiDocument * doc, KuiCPopupType type, int * out_has) {
    if (g_kui_document_has_popup_of_type == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_has_popup_of_type(doc, type, out_has);
}

KuiStatus kui_document_set_tooltip_delay(KuiDocument * doc, double ms) {
    if (g_kui_document_set_tooltip_delay == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_tooltip_delay(doc, ms);
}

KuiStatus kui_document_get_tooltip_delay(KuiDocument * doc, double * out_ms) {
    if (g_kui_document_get_tooltip_delay == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_tooltip_delay(doc, out_ms);
}

KuiStatus kui_document_set_tooltip_theme(KuiDocument * doc, const KuiCTooltipTheme * theme) {
    if (g_kui_document_set_tooltip_theme == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_tooltip_theme(doc, theme);
}

KuiStatus kui_document_get_tooltip_theme(KuiDocument * doc, KuiCTooltipTheme * out_theme) {
    if (g_kui_document_get_tooltip_theme == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_tooltip_theme(doc, out_theme);
}

KuiStatus kui_document_set_popup_mode(KuiDocument * doc, KuiCPopupMode mode) {
    if (g_kui_document_set_popup_mode == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_popup_mode(doc, mode);
}

KuiStatus kui_document_set_default_context_menu_enabled(KuiDocument * doc, int enabled) {
    if (g_kui_document_set_default_context_menu_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_default_context_menu_enabled(doc, enabled);
}

KuiStatus kui_document_get_default_context_menu_enabled(KuiDocument * doc, int * out_enabled) {
    if (g_kui_document_get_default_context_menu_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_default_context_menu_enabled(doc, out_enabled);
}

KuiStatus kui_element_open_input_picker(KuiElement * el, KuiCPickerMode mode) {
    if (g_kui_element_open_input_picker == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_open_input_picker(el, mode);
}

KuiStatus kui_document_close_input_picker(KuiDocument * doc) {
    if (g_kui_document_close_input_picker == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_close_input_picker(doc);
}

KuiStatus kui_document_commit_input_picker(KuiDocument * doc) {
    if (g_kui_document_commit_input_picker == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_commit_input_picker(doc);
}

KuiStatus kui_document_get_input_picker_state(KuiDocument * doc, KuiCInputPickerState * out_state) {
    if (g_kui_document_get_input_picker_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_input_picker_state(doc, out_state);
}

KuiStatus kui_document_get_input_picker_target(KuiDocument * doc, KuiElement ** out) {
    if (g_kui_document_get_input_picker_target == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_input_picker_target(doc, out);
}

KuiStatus kui_document_set_input_picker_enabled(KuiDocument * doc, int enabled) {
    if (g_kui_document_set_input_picker_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_input_picker_enabled(doc, enabled);
}

KuiStatus kui_document_set_input_picker_month(KuiDocument * doc, int year, int month) {
    if (g_kui_document_set_input_picker_month == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_input_picker_month(doc, year, month);
}

KuiStatus kui_document_set_input_picker_color_hsv(KuiDocument * doc, float h, float s, float v) {
    if (g_kui_document_set_input_picker_color_hsv == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_input_picker_color_hsv(doc, h, s, v);
}

KuiStatus kui_document_set_input_picker_time(KuiDocument * doc, int hour, int minute) {
    if (g_kui_document_set_input_picker_time == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_input_picker_time(doc, hour, minute);
}

KuiStatus kui_document_audio_is_available(KuiDocument * doc, int * out_available) {
    if (g_kui_document_audio_is_available == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_is_available(doc, out_available);
}

KuiStatus kui_document_audio_play(KuiDocument * doc, const char * path, size_t len, const KuiCAudioPlayOptions * options, uint32_t * out_handle) {
    if (g_kui_document_audio_play == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_play(doc, path, len, options, out_handle);
}

KuiStatus kui_document_audio_stop(KuiDocument * doc, uint32_t handle) {
    if (g_kui_document_audio_stop == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_stop(doc, handle);
}

KuiStatus kui_document_audio_set_paused(KuiDocument * doc, uint32_t handle, int paused) {
    if (g_kui_document_audio_set_paused == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_set_paused(doc, handle, paused);
}

KuiStatus kui_document_audio_set_volume(KuiDocument * doc, uint32_t handle, float volume) {
    if (g_kui_document_audio_set_volume == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_set_volume(doc, handle, volume);
}

KuiStatus kui_document_audio_set_pan(KuiDocument * doc, uint32_t handle, float pan) {
    if (g_kui_document_audio_set_pan == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_set_pan(doc, handle, pan);
}

KuiStatus kui_document_audio_set_pitch(KuiDocument * doc, uint32_t handle, float pitch) {
    if (g_kui_document_audio_set_pitch == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_set_pitch(doc, handle, pitch);
}

KuiStatus kui_document_audio_seek(KuiDocument * doc, uint32_t handle, double seconds) {
    if (g_kui_document_audio_seek == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_seek(doc, handle, seconds);
}

KuiStatus kui_document_audio_get_info(KuiDocument * doc, uint32_t handle, KuiCSoundInfo * out_info) {
    if (g_kui_document_audio_get_info == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_get_info(doc, handle, out_info);
}

KuiStatus kui_document_audio_set_group(KuiDocument * doc, KuiCAudioGroup group, float volume, int muted) {
    if (g_kui_document_audio_set_group == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_set_group(doc, group, volume, muted);
}

KuiStatus kui_document_audio_get_group(KuiDocument * doc, KuiCAudioGroup group, float * out_volume, int * out_muted) {
    if (g_kui_document_audio_get_group == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_get_group(doc, group, out_volume, out_muted);
}

KuiStatus kui_document_audio_stop_all(KuiDocument * doc, int group) {
    if (g_kui_document_audio_stop_all == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_stop_all(doc, group);
}

KuiStatus kui_document_audio_set_all_paused(KuiDocument * doc, int group, int paused) {
    if (g_kui_document_audio_set_all_paused == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_set_all_paused(doc, group, paused);
}

KuiStatus kui_document_audio_preload(KuiDocument * doc, const char * path, size_t len, uint32_t count, uint32_t * out_loaded) {
    if (g_kui_document_audio_preload == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_preload(doc, path, len, count, out_loaded);
}

KuiStatus kui_document_audio_unload(KuiDocument * doc, const char * path, size_t len) {
    if (g_kui_document_audio_unload == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_unload(doc, path, len);
}

KuiStatus kui_document_audio_get_counts(KuiDocument * doc, int group, uint32_t * out_active, uint32_t * out_pooled) {
    if (g_kui_document_audio_get_counts == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_audio_get_counts(doc, group, out_active, out_pooled);
}

KuiStatus kui_media_load(KuiElement * element) {
    if (g_kui_media_load == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_load(element);
}

KuiStatus kui_media_set_loop(KuiElement * element, int loop) {
    if (g_kui_media_set_loop == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_set_loop(element, loop);
}

KuiStatus kui_media_set_autoplay(KuiElement * element, int autoplay) {
    if (g_kui_media_set_autoplay == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_set_autoplay(element, autoplay);
}

KuiStatus kui_media_get_playback_flags(KuiElement * element, int * out_loop, int * out_autoplay, int * out_ended) {
    if (g_kui_media_get_playback_flags == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_media_get_playback_flags(element, out_loop, out_autoplay, out_ended);
}

KuiStatus kui_lottie_stop(KuiElement * element) {
    if (g_kui_lottie_stop == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_stop(element);
}

KuiStatus kui_lottie_seek(KuiElement * element, double position) {
    if (g_kui_lottie_seek == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_seek(element, position);
}

KuiStatus kui_lottie_set_speed(KuiElement * element, float speed) {
    if (g_kui_lottie_set_speed == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_set_speed(element, speed);
}

KuiStatus kui_lottie_set_loop(KuiElement * element, int loop) {
    if (g_kui_lottie_set_loop == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_set_loop(element, loop);
}

KuiStatus kui_lottie_get_state(KuiElement * element, KuiCLottieState * out_state) {
    if (g_kui_lottie_get_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_lottie_get_state(element, out_state);
}

KuiStatus kui_document_load_extension(KuiDocument * doc, const char * path, size_t len, int * out_loaded) {
    if (g_kui_document_load_extension == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_load_extension(doc, path, len, out_loaded);
}

KuiStatus kui_document_unload_extension(KuiDocument * doc, const char * tag_name, size_t len) {
    if (g_kui_document_unload_extension == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_unload_extension(doc, tag_name, len);
}

KuiStatus kui_document_reload_extension(KuiDocument * doc, const char * path, size_t len, int * out_reloaded) {
    if (g_kui_document_reload_extension == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_reload_extension(doc, path, len, out_reloaded);
}

KuiStatus kui_document_unload_all_extensions(KuiDocument * doc) {
    if (g_kui_document_unload_all_extensions == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_unload_all_extensions(doc);
}

KuiStatus kui_document_watch_extensions(KuiDocument * doc, const char * directory, size_t len) {
    if (g_kui_document_watch_extensions == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_watch_extensions(doc, directory, len);
}

KuiStatus kui_document_get_extension_state(KuiDocument * doc, size_t * out_count, int * out_watching) {
    if (g_kui_document_get_extension_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_extension_state(doc, out_count, out_watching);
}

KuiStatus kui_window_stop_devtools(KuiWindow * w) {
    if (g_kui_window_stop_devtools == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_stop_devtools(w);
}

KuiStatus kui_window_get_devtools_state(KuiWindow * w, int * out_running, uint16_t * out_port, uint32_t * out_clients) {
    if (g_kui_window_get_devtools_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_get_devtools_state(w, out_running, out_port, out_clients);
}

KuiStatus kui_event_drag_set_data(KuiEvent * event, const char * format, size_t format_len, const char * data, size_t data_len) {
    if (g_kui_event_drag_set_data == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_set_data(event, format, format_len, data, data_len);
}

KuiStatus kui_event_drag_get_data(KuiEvent * event, const char * format, size_t format_len, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_event_drag_get_data == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_get_data(event, format, format_len, buf, buf_capacity, out_len);
}

KuiStatus kui_event_drag_clear_data(KuiEvent * event, const char * format, size_t format_len) {
    if (g_kui_event_drag_clear_data == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_clear_data(event, format, format_len);
}

KuiStatus kui_event_drag_get_type_count(KuiEvent * event, size_t * out_count) {
    if (g_kui_event_drag_get_type_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_get_type_count(event, out_count);
}

KuiStatus kui_event_drag_get_type(KuiEvent * event, size_t index, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_event_drag_get_type == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_get_type(event, index, buf, buf_capacity, out_len);
}

KuiStatus kui_event_drag_set_drop_effect(KuiEvent * event, KuiCDropEffect effect) {
    if (g_kui_event_drag_set_drop_effect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_set_drop_effect(event, effect);
}

KuiStatus kui_event_drag_get_drop_effect(KuiEvent * event, KuiCDropEffect * out_effect) {
    if (g_kui_event_drag_get_drop_effect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_get_drop_effect(event, out_effect);
}

KuiStatus kui_event_drag_set_effect_allowed(KuiEvent * event, KuiCEffectAllowed effect) {
    if (g_kui_event_drag_set_effect_allowed == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_set_effect_allowed(event, effect);
}

KuiStatus kui_event_drag_get_effect_allowed(KuiEvent * event, KuiCEffectAllowed * out_effect) {
    if (g_kui_event_drag_get_effect_allowed == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_drag_get_effect_allowed(event, out_effect);
}

KuiStatus kui_script_engine_eval_value(KuiScriptEngine * engine, const char * code, size_t code_len, const char * source, size_t source_len, int * out_success, KuiCScriptValue * out_value) {
    if (g_kui_script_engine_eval_value == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_eval_value(engine, code, code_len, source, source_len, out_success, out_value);
}

KuiStatus kui_script_engine_eval_file(KuiScriptEngine * engine, const char * path, size_t path_len, int * out_success, KuiCScriptValue * out_value) {
    if (g_kui_script_engine_eval_file == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_eval_file(engine, path, path_len, out_success, out_value);
}

KuiStatus kui_script_engine_call_function(KuiScriptEngine * engine, const char * name, size_t name_len, const KuiCScriptValue * args, size_t arg_count, int * out_success, KuiCScriptValue * out_value) {
    if (g_kui_script_engine_call_function == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_call_function(engine, name, name_len, args, arg_count, out_success, out_value);
}

KuiStatus kui_script_engine_get_last_error(KuiScriptEngine * engine, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_script_engine_get_last_error == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_get_last_error(engine, buf, buf_capacity, out_len);
}

KuiStatus kui_script_engine_register_function(KuiScriptEngine * engine, const char * name, size_t name_len, KuiScriptFunctionFn fn, void * user_data, KuiTaskFn release_fn) {
    if (g_kui_script_engine_register_function == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_register_function(engine, name, name_len, fn, user_data, release_fn);
}

KuiStatus kui_script_engine_set_global(KuiScriptEngine * engine, const char * name, size_t name_len, const KuiCScriptValue * value) {
    if (g_kui_script_engine_set_global == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_set_global(engine, name, name_len, value);
}

KuiStatus kui_script_engine_get_global(KuiScriptEngine * engine, const char * name, size_t name_len, KuiCScriptValue * out_value) {
    if (g_kui_script_engine_get_global == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_get_global(engine, name, name_len, out_value);
}

KuiStatus kui_script_engine_set_print_handler(KuiScriptEngine * engine, KuiScriptPrintFn fn, void * user_data, KuiTaskFn release_fn) {
    if (g_kui_script_engine_set_print_handler == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_set_print_handler(engine, fn, user_data, release_fn);
}

KuiStatus kui_script_engine_set_memory_limit(KuiScriptEngine * engine, size_t bytes) {
    if (g_kui_script_engine_set_memory_limit == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_set_memory_limit(engine, bytes);
}

KuiStatus kui_script_engine_set_instruction_limit(KuiScriptEngine * engine, uint64_t max_instructions) {
    if (g_kui_script_engine_set_instruction_limit == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_set_instruction_limit(engine, max_instructions);
}

KuiStatus kui_script_engine_get_resource_usage(KuiScriptEngine * engine, size_t * out_memory_bytes, size_t * out_memory_limit, uint64_t * out_instruction_limit) {
    if (g_kui_script_engine_get_resource_usage == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_get_resource_usage(engine, out_memory_bytes, out_memory_limit, out_instruction_limit);
}

KuiStatus kui_script_function_create(KuiScriptEngine * engine, KuiScriptFunctionFn fn, void * user_data, KuiTaskFn release_fn, KuiScriptFunction ** out_function) {
    if (g_kui_script_function_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_function_create(engine, fn, user_data, release_fn, out_function);
}

KuiStatus kui_script_function_call(KuiScriptFunction * function, const KuiCScriptValue * args, size_t arg_count, int * out_success, KuiCScriptValue * out_value) {
    if (g_kui_script_function_call == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_function_call(function, args, arg_count, out_success, out_value);
}

KuiStatus kui_script_function_retain(KuiScriptFunction * function, KuiScriptFunction ** out_owned) {
    if (g_kui_script_function_retain == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_function_retain(function, out_owned);
}

KuiStatus kui_script_function_release(KuiScriptFunction * function) {
    if (g_kui_script_function_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_function_release(function);
}

KuiStatus kui_document_define_custom_element_ex(KuiDocument * document, const char * tag_name, size_t len, const KuiCCustomElementDefinition * definition) {
    if (g_kui_document_define_custom_element_ex == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_define_custom_element_ex(document, tag_name, len, definition);
}

KuiStatus kui_paint_get_info(KuiPaintContext * paint, float * out_width, float * out_height, float * out_device_pixel_ratio) {
    if (g_kui_paint_get_info == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_get_info(paint, out_width, out_height, out_device_pixel_ratio);
}

KuiStatus kui_paint_fill_rect(KuiPaintContext * paint, float x, float y, float width, float height, uint32_t rgba) {
    if (g_kui_paint_fill_rect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_fill_rect(paint, x, y, width, height, rgba);
}

KuiStatus kui_paint_fill_rounded_rect(KuiPaintContext * paint, float x, float y, float width, float height, const KuiCCornerRadii * radii, uint32_t rgba) {
    if (g_kui_paint_fill_rounded_rect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_fill_rounded_rect(paint, x, y, width, height, radii, rgba);
}

KuiStatus kui_paint_stroke_rect(KuiPaintContext * paint, float x, float y, float width, float height, float line_width, uint32_t rgba) {
    if (g_kui_paint_stroke_rect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_stroke_rect(paint, x, y, width, height, line_width, rgba);
}

KuiStatus kui_paint_stroke_rounded_rect(KuiPaintContext * paint, float x, float y, float width, float height, const KuiCCornerRadii * radii, float line_width, uint32_t rgba) {
    if (g_kui_paint_stroke_rounded_rect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_stroke_rounded_rect(paint, x, y, width, height, radii, line_width, rgba);
}

KuiStatus kui_paint_draw_text(KuiPaintContext * paint, const char * text, size_t len, float x, float baseline_y, const KuiCTextStyle * style, uint32_t rgba) {
    if (g_kui_paint_draw_text == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_draw_text(paint, text, len, x, baseline_y, style, rgba);
}

KuiStatus kui_paint_draw_pixels(KuiPaintContext * paint, const uint32_t * pixels, uint32_t width, uint32_t height, float x, float y, float dest_width, float dest_height) {
    if (g_kui_paint_draw_pixels == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_draw_pixels(paint, pixels, width, height, x, y, dest_width, dest_height);
}

KuiStatus kui_paint_draw_image(KuiPaintContext * paint, const char * path, size_t len, const KuiRect * src, const KuiRect * dest) {
    if (g_kui_paint_draw_image == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_draw_image(paint, path, len, src, dest);
}

KuiStatus kui_paint_draw_decoration_line(KuiPaintContext * paint, float x, float y, float width, float thickness, KuiCDecorationStyle style, uint32_t rgba) {
    if (g_kui_paint_draw_decoration_line == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_draw_decoration_line(paint, x, y, width, thickness, style, rgba);
}

KuiStatus kui_paint_save(KuiPaintContext * paint) {
    if (g_kui_paint_save == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_save(paint);
}

KuiStatus kui_paint_restore(KuiPaintContext * paint) {
    if (g_kui_paint_restore == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_restore(paint);
}

KuiStatus kui_paint_translate(KuiPaintContext * paint, float dx, float dy) {
    if (g_kui_paint_translate == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_translate(paint, dx, dy);
}

KuiStatus kui_paint_scale(KuiPaintContext * paint, float sx, float sy) {
    if (g_kui_paint_scale == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_scale(paint, sx, sy);
}

KuiStatus kui_paint_rotate(KuiPaintContext * paint, float degrees) {
    if (g_kui_paint_rotate == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_rotate(paint, degrees);
}

KuiStatus kui_paint_set_opacity(KuiPaintContext * paint, float opacity) {
    if (g_kui_paint_set_opacity == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_set_opacity(paint, opacity);
}

KuiStatus kui_paint_clip_rect(KuiPaintContext * paint, float x, float y, float width, float height) {
    if (g_kui_paint_clip_rect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_clip_rect(paint, x, y, width, height);
}

KuiStatus kui_paint_clip_rounded_rect(KuiPaintContext * paint, float x, float y, float width, float height, const KuiCCornerRadii * radii) {
    if (g_kui_paint_clip_rounded_rect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_paint_clip_rounded_rect(paint, x, y, width, height, radii);
}

KuiStatus kui_element_measure_text(KuiElement * element, const char * text, size_t len, const KuiCTextStyle * style, KuiCTextMetrics * out_metrics) {
    if (g_kui_element_measure_text == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_measure_text(element, text, len, style, out_metrics);
}

KuiStatus kui_element_request_repaint(KuiElement * element) {
    if (g_kui_element_request_repaint == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_request_repaint(element);
}

KuiStatus kui_element_invalidate_rect(KuiElement * element, float x, float y, float width, float height) {
    if (g_kui_element_invalidate_rect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_invalidate_rect(element, x, y, width, height);
}

KuiStatus kui_devtools_dispatch_ime_event(KuiDocument * document, KuiCImeEventType type, const char * text, size_t len) {
    if (g_kui_devtools_dispatch_ime_event == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_devtools_dispatch_ime_event(document, type, text, len);
}

KuiStatus kui_document_register_builtin_components(KuiDocument * document) {
    if (g_kui_document_register_builtin_components == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_register_builtin_components(document);
}

KuiStatus kui_theme_manager_create(KuiThemeManager ** out_manager) {
    if (g_kui_theme_manager_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_create(out_manager);
}

KuiStatus kui_theme_manager_release(KuiThemeManager * manager) {
    if (g_kui_theme_manager_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_release(manager);
}

KuiStatus kui_theme_manager_register(KuiThemeManager * manager, const char * name, size_t name_len, const KuiCKeyValue * variables, size_t variable_count) {
    if (g_kui_theme_manager_register == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_register(manager, name, name_len, variables, variable_count);
}

KuiStatus kui_theme_manager_apply(KuiThemeManager * manager, const char * name, size_t name_len, KuiDocument * document) {
    if (g_kui_theme_manager_apply == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_apply(manager, name, name_len, document);
}

KuiStatus kui_theme_manager_get_active(KuiThemeManager * manager, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_theme_manager_get_active == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_get_active(manager, buf, buf_capacity, out_len);
}

KuiStatus kui_theme_manager_get_theme_count(KuiThemeManager * manager, size_t * out_count) {
    if (g_kui_theme_manager_get_theme_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_get_theme_count(manager, out_count);
}

KuiStatus kui_theme_manager_get_theme_name(KuiThemeManager * manager, size_t index, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_theme_manager_get_theme_name == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_get_theme_name(manager, index, buf, buf_capacity, out_len);
}

KuiStatus kui_theme_manager_get_variable_count(KuiThemeManager * manager, const char * theme, size_t theme_len, size_t * out_count) {
    if (g_kui_theme_manager_get_variable_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_get_variable_count(manager, theme, theme_len, out_count);
}

KuiStatus kui_theme_manager_get_variable(KuiThemeManager * manager, const char * theme, size_t theme_len, size_t index, char * name_buf, size_t name_capacity, size_t * out_name_len, char * value_buf, size_t value_capacity, size_t * out_value_len) {
    if (g_kui_theme_manager_get_variable == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_theme_manager_get_variable(manager, theme, theme_len, index, name_buf, name_capacity, out_name_len, value_buf, value_capacity, out_value_len);
}

KuiStatus kui_virtual_list_create(KuiElement * container, const KuiCVirtualListConfig * config, KuiVirtualList ** out_list) {
    if (g_kui_virtual_list_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_create(container, config, out_list);
}

KuiStatus kui_virtual_list_release(KuiVirtualList * list) {
    if (g_kui_virtual_list_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_release(list);
}

KuiStatus kui_virtual_list_get_state(KuiVirtualList * list, KuiCVirtualListState * out_state) {
    if (g_kui_virtual_list_get_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_get_state(list, out_state);
}

KuiStatus kui_virtual_list_set_total_items(KuiVirtualList * list, uint32_t count) {
    if (g_kui_virtual_list_set_total_items == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_set_total_items(list, count);
}

KuiStatus kui_virtual_list_insert_items(KuiVirtualList * list, uint32_t index, uint32_t count) {
    if (g_kui_virtual_list_insert_items == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_insert_items(list, index, count);
}

KuiStatus kui_virtual_list_remove_items(KuiVirtualList * list, uint32_t index, uint32_t count) {
    if (g_kui_virtual_list_remove_items == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_remove_items(list, index, count);
}

KuiStatus kui_virtual_list_refresh(KuiVirtualList * list) {
    if (g_kui_virtual_list_refresh == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_refresh(list);
}

KuiStatus kui_virtual_list_scroll_to_index(KuiVirtualList * list, uint32_t index) {
    if (g_kui_virtual_list_scroll_to_index == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_scroll_to_index(list, index);
}

KuiStatus kui_virtual_list_scroll_to_top(KuiVirtualList * list) {
    if (g_kui_virtual_list_scroll_to_top == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_scroll_to_top(list);
}

KuiStatus kui_virtual_list_scroll_to_bottom(KuiVirtualList * list) {
    if (g_kui_virtual_list_scroll_to_bottom == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_scroll_to_bottom(list);
}

KuiStatus kui_virtual_list_set_scroll_unit(KuiVirtualList * list, KuiCScrollUnit unit) {
    if (g_kui_virtual_list_set_scroll_unit == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_set_scroll_unit(list, unit);
}

KuiStatus kui_virtual_list_set_focused_index(KuiVirtualList * list, int32_t index) {
    if (g_kui_virtual_list_set_focused_index == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_set_focused_index(list, index);
}

KuiStatus kui_virtual_list_move_focus(KuiVirtualList * list, int32_t delta) {
    if (g_kui_virtual_list_move_focus == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_virtual_list_move_focus(list, delta);
}

KuiStatus kui_list_view_create(KuiElement * container, const KuiCListViewConfig * config, KuiListView ** out_list) {
    if (g_kui_list_view_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_create(container, config, out_list);
}

KuiStatus kui_list_view_release(KuiListView * list) {
    if (g_kui_list_view_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_release(list);
}

KuiStatus kui_list_view_select(KuiListView * list, uint32_t index) {
    if (g_kui_list_view_select == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_select(list, index);
}

KuiStatus kui_list_view_deselect(KuiListView * list, uint32_t index) {
    if (g_kui_list_view_deselect == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_deselect(list, index);
}

KuiStatus kui_list_view_clear_selection(KuiListView * list) {
    if (g_kui_list_view_clear_selection == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_clear_selection(list);
}

KuiStatus kui_list_view_select_range(KuiListView * list, uint32_t start, uint32_t end) {
    if (g_kui_list_view_select_range == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_select_range(list, start, end);
}

KuiStatus kui_list_view_is_selected(KuiListView * list, uint32_t index, int * out_selected) {
    if (g_kui_list_view_is_selected == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_is_selected(list, index, out_selected);
}

KuiStatus kui_list_view_get_selected(KuiListView * list, uint32_t * buf, size_t buf_capacity, size_t * out_count) {
    if (g_kui_list_view_get_selected == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_get_selected(list, buf, buf_capacity, out_count);
}

KuiStatus kui_list_view_get_focused_index(KuiListView * list, int32_t * out_index) {
    if (g_kui_list_view_get_focused_index == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_get_focused_index(list, out_index);
}

KuiStatus kui_list_view_set_item_count(KuiListView * list, uint32_t count) {
    if (g_kui_list_view_set_item_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_set_item_count(list, count);
}

KuiStatus kui_list_view_refresh(KuiListView * list) {
    if (g_kui_list_view_refresh == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_refresh(list);
}

KuiStatus kui_list_view_scroll_to_index(KuiListView * list, uint32_t index) {
    if (g_kui_list_view_scroll_to_index == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_scroll_to_index(list, index);
}

KuiStatus kui_list_view_scroll_to_top(KuiListView * list) {
    if (g_kui_list_view_scroll_to_top == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_scroll_to_top(list);
}

KuiStatus kui_list_view_scroll_to_bottom(KuiListView * list) {
    if (g_kui_list_view_scroll_to_bottom == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_scroll_to_bottom(list);
}

KuiStatus kui_list_view_set_scroll_unit(KuiListView * list, KuiCScrollUnit unit) {
    if (g_kui_list_view_set_scroll_unit == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_set_scroll_unit(list, unit);
}

KuiStatus kui_list_view_get_state(KuiListView * list, KuiCVirtualListState * out_state) {
    if (g_kui_list_view_get_state == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_list_view_get_state(list, out_state);
}

KuiStatus kui_tree_view_create(KuiTreeView ** out_tree) {
    if (g_kui_tree_view_create == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_create(out_tree);
}

KuiStatus kui_tree_view_release(KuiTreeView * tree) {
    if (g_kui_tree_view_release == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_release(tree);
}

KuiStatus kui_tree_view_add_node(KuiTreeView * tree, const char * text, size_t text_len, const char * icon, size_t icon_len, uint32_t * out_index) {
    if (g_kui_tree_view_add_node == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_add_node(tree, text, text_len, icon, icon_len, out_index);
}

KuiStatus kui_tree_view_add_child(KuiTreeView * tree, uint32_t parent, const char * text, size_t text_len, const char * icon, size_t icon_len, uint32_t * out_index) {
    if (g_kui_tree_view_add_child == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_add_child(tree, parent, text, text_len, icon, icon_len, out_index);
}

KuiStatus kui_tree_view_get_node_count(KuiTreeView * tree, uint32_t * out_count) {
    if (g_kui_tree_view_get_node_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_get_node_count(tree, out_count);
}

KuiStatus kui_tree_view_get_node(KuiTreeView * tree, uint32_t index, KuiCTreeNode * out_node) {
    if (g_kui_tree_view_get_node == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_get_node(tree, index, out_node);
}

KuiStatus kui_tree_view_get_node_child(KuiTreeView * tree, uint32_t index, size_t position, uint32_t * out_child) {
    if (g_kui_tree_view_get_node_child == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_get_node_child(tree, index, position, out_child);
}

KuiStatus kui_tree_view_set_node(KuiTreeView * tree, uint32_t index, const char * text, size_t text_len, const char * icon, size_t icon_len, int disabled) {
    if (g_kui_tree_view_set_node == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_set_node(tree, index, text, text_len, icon, icon_len, disabled);
}

KuiStatus kui_tree_view_setup(KuiTreeView * tree, KuiElement * container, const KuiCTreeViewConfig * config) {
    if (g_kui_tree_view_setup == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_setup(tree, container, config);
}

KuiStatus kui_tree_view_expand(KuiTreeView * tree, uint32_t index) {
    if (g_kui_tree_view_expand == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_expand(tree, index);
}

KuiStatus kui_tree_view_collapse(KuiTreeView * tree, uint32_t index) {
    if (g_kui_tree_view_collapse == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_collapse(tree, index);
}

KuiStatus kui_tree_view_toggle(KuiTreeView * tree, uint32_t index) {
    if (g_kui_tree_view_toggle == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_toggle(tree, index);
}

KuiStatus kui_tree_view_expand_all(KuiTreeView * tree) {
    if (g_kui_tree_view_expand_all == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_expand_all(tree);
}

KuiStatus kui_tree_view_collapse_all(KuiTreeView * tree) {
    if (g_kui_tree_view_collapse_all == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_collapse_all(tree);
}

KuiStatus kui_tree_view_select(KuiTreeView * tree, uint32_t index) {
    if (g_kui_tree_view_select == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_select(tree, index);
}

KuiStatus kui_tree_view_clear_selection(KuiTreeView * tree) {
    if (g_kui_tree_view_clear_selection == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_clear_selection(tree);
}

KuiStatus kui_tree_view_toggle_select(KuiTreeView * tree, uint32_t index) {
    if (g_kui_tree_view_toggle_select == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_toggle_select(tree, index);
}

KuiStatus kui_tree_view_select_range(KuiTreeView * tree, uint32_t index) {
    if (g_kui_tree_view_select_range == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_select_range(tree, index);
}

KuiStatus kui_tree_view_select_all(KuiTreeView * tree) {
    if (g_kui_tree_view_select_all == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_select_all(tree);
}

KuiStatus kui_tree_view_is_selected(KuiTreeView * tree, uint32_t index, int * out_selected) {
    if (g_kui_tree_view_is_selected == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_is_selected(tree, index, out_selected);
}

KuiStatus kui_tree_view_get_selected_index(KuiTreeView * tree, int32_t * out_index) {
    if (g_kui_tree_view_get_selected_index == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_get_selected_index(tree, out_index);
}

KuiStatus kui_tree_view_get_selected(KuiTreeView * tree, uint32_t * buf, size_t buf_capacity, size_t * out_count) {
    if (g_kui_tree_view_get_selected == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_get_selected(tree, buf, buf_capacity, out_count);
}

KuiStatus kui_tree_view_refresh(KuiTreeView * tree) {
    if (g_kui_tree_view_refresh == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_refresh(tree);
}

KuiStatus kui_tree_view_scroll_to_node(KuiTreeView * tree, uint32_t index) {
    if (g_kui_tree_view_scroll_to_node == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_scroll_to_node(tree, index);
}

KuiStatus kui_tree_view_set_scroll_unit(KuiTreeView * tree, KuiCScrollUnit unit) {
    if (g_kui_tree_view_set_scroll_unit == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_set_scroll_unit(tree, unit);
}

KuiStatus kui_tree_view_get_visible_row_count(KuiTreeView * tree, uint32_t * out_count) {
    if (g_kui_tree_view_get_visible_row_count == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_tree_view_get_visible_row_count(tree, out_count);
}

KuiStatus kui_document_set_style_profile(KuiDocument * document, KuiCStyleProfile profile) {
    if (g_kui_document_set_style_profile == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_style_profile(document, profile);
}

KuiStatus kui_document_get_style_profile(KuiDocument * document, KuiCStyleProfile * out_profile) {
    if (g_kui_document_get_style_profile == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_style_profile(document, out_profile);
}

KuiStatus kui_element_start_momentum_scroll(KuiElement * el, float velocity_x, float velocity_y) {
    if (g_kui_element_start_momentum_scroll == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_start_momentum_scroll(el, velocity_x, velocity_y);
}

KuiStatus kui_element_scroll_into_view(KuiElement * el, int align_to_top) {
    if (g_kui_element_scroll_into_view == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_element_scroll_into_view(el, align_to_top);
}

KuiStatus kui_event_wheel_get_delta(KuiEvent * evt, float * out_delta_x, float * out_delta_y) {
    if (g_kui_event_wheel_get_delta == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_event_wheel_get_delta(evt, out_delta_x, out_delta_y);
}

KuiStatus kui_script_engine_tick_timers(KuiScriptEngine * engine, double now_ms) {
    if (g_kui_script_engine_tick_timers == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_script_engine_tick_timers(engine, now_ms);
}

KuiStatus kui_document_set_base_path(KuiDocument * doc, const char * path, size_t path_len) {
    if (g_kui_document_set_base_path == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_base_path(doc, path, path_len);
}

KuiStatus kui_document_get_base_path(KuiDocument * doc, char * buf, size_t buf_capacity, size_t * out_len) {
    if (g_kui_document_get_base_path == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_get_base_path(doc, buf, buf_capacity, out_len);
}

KuiStatus kui_document_set_render_device_pixel_ratio(KuiDocument * doc, float dpr) {
    if (g_kui_document_set_render_device_pixel_ratio == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_document_set_render_device_pixel_ratio(doc, dpr);
}

KuiStatus kui_window_create_ex(const KuiWindowPropsEx * props, KuiWindow ** out) {
    if (g_kui_window_create_ex == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_create_ex(props, out);
}

KuiStatus kui_window_set_enabled(KuiWindow * w, int enabled) {
    if (g_kui_window_set_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_enabled(w, enabled);
}

KuiStatus kui_window_is_enabled(KuiWindow * w, int * out_enabled) {
    if (g_kui_window_is_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_is_enabled(w, out_enabled);
}

KuiStatus kui_window_process_frame(KuiWindow * w) {
    if (g_kui_window_process_frame == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_process_frame(w);
}

KuiStatus kui_window_set_devtools_token_auth(KuiWindow * w, int enabled) {
    if (g_kui_window_set_devtools_token_auth == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_window_set_devtools_token_auth(w, enabled);
}

KuiStatus kui_platform_disable_ime(void) {
    if (g_kui_platform_disable_ime == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_platform_disable_ime();
}

KuiStatus kui_console_set_enabled(int enabled) {
    if (g_kui_console_set_enabled == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_console_set_enabled(enabled);
}

KuiStatus kui_console_set_stderr_output(int enabled) {
    if (g_kui_console_set_stderr_output == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_console_set_stderr_output(enabled);
}

KuiStatus kui_console_set_min_level(KuiCLogLevel level) {
    if (g_kui_console_set_min_level == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_console_set_min_level(level);
}

KuiStatus kui_stack_trace_install(const KuiCStackTraceConfig * config, int * out_installed) {
    if (g_kui_stack_trace_install == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_stack_trace_install(config, out_installed);
}

KuiStatus kui_stack_trace_uninstall(void) {
    if (g_kui_stack_trace_uninstall == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_stack_trace_uninstall();
}

KuiStatus kui_crash_reporter_install(const KuiCCrashReporterConfig * config, int * out_installed) {
    if (g_kui_crash_reporter_install == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_crash_reporter_install(config, out_installed);
}

KuiStatus kui_crash_reporter_uninstall(void) {
    if (g_kui_crash_reporter_uninstall == NULL) { return KUI_ERR_NOT_SUPPORTED; }
    return g_kui_crash_reporter_uninstall();
}
