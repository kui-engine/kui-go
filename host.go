//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

/*
#cgo CFLAGS: -std=c11 -I${SRCDIR}
#include <stdlib.h>
#include "loader.h"
*/
import "C"

import "unsafe"

// 宿主集成类接口（C ABI 追加）：惯性滚动 / 滚入视图、脚本定时器、基准路径、窗口启用与单帧渲染、
// DevTools 令牌开关、输入法开关、控制台与崩溃诊断。

// StartMomentumScroll 以速度（px/ms，y 为正向下）启动惯性滚动。
func (e *Element) StartMomentumScroll(velocityX, velocityY float32) error {
	return statusError(C.kui_element_start_momentum_scroll(e.raw(), C.float(velocityX), C.float(velocityY)), "Element.StartMomentumScroll")
}

// ScrollIntoView 滚动最近的可滚动祖先让元素可见；alignToTop 为真时顶对齐，否则底对齐。
func (e *Element) ScrollIntoView(alignToTop bool) error {
	return statusError(C.kui_element_scroll_into_view(e.raw(), boolInt(alignToTop)), "Element.ScrollIntoView")
}

// TickTimers 运行到期的 Lua 定时器（setTimeout / setInterval）；nowMs 与 Now() 同一时钟（毫秒）。
// 引擎自己不会走时间：加载了带定时器脚本的宿主每帧调一次。
func (s *ScriptEngine) TickTimers(nowMs float64) error {
	return statusError(C.kui_script_engine_tick_timers(s.raw(), C.double(nowMs)), "ScriptEngine.TickTimers")
}

// SetBasePath 设置文档解析相对资源路径（图片 / 样式表 / 脚本）的目录，以路径分隔符结尾；不会重新加载。
func (d *Document) SetBasePath(path string) error {
	raw, size, release := inputString(path)
	defer release()
	return statusError(C.kui_document_set_base_path(d.raw(), raw, size), "Document.SetBasePath")
}

// BasePath 读回基准路径。
func (d *Document) BasePath() (string, error) {
	var size C.size_t
	if err := bufferStatus(C.kui_document_get_base_path(d.raw(), nil, 0, &size), "Document.BasePath"); err != nil {
		return "", err
	}
	buffer := make([]byte, int(size)+1)
	if err := statusError(C.kui_document_get_base_path(d.raw(), (*C.char)(unsafe.Pointer(&buffer[0])), C.size_t(len(buffer)), &size), "Document.BasePath"); err != nil {
		return "", err
	}
	return string(buffer[:int(size)]), nil
}

// SetRenderDevicePixelRatio 设置渲染后端的设备像素比（页面光栅化的倍率）；文档还没有渲染器（没有窗口）时返回 KUI_ERR_NOT_SUPPORTED。
func (d *Document) SetRenderDevicePixelRatio(ratio float32) error {
	return statusError(C.kui_document_set_render_device_pixel_ratio(d.raw(), C.float(ratio)), "Document.SetRenderDevicePixelRatio")
}

// SetEnabled 启用 / 禁用窗口的鼠标键盘输入（模态父窗口）；仅 Windows，其他平台返回 KUI_ERR_NOT_SUPPORTED。
func (w *Window) SetEnabled(enabled bool) error {
	return statusError(C.kui_window_set_enabled(w.raw(), boolInt(enabled)), "Window.SetEnabled")
}

// IsEnabled 读回窗口是否启用输入；仅 Windows。
func (w *Window) IsEnabled() (bool, error) {
	var value C.int
	err := statusError(C.kui_window_is_enabled(w.raw(), &value), "Window.IsEnabled")
	return value != 0, err
}

// ProcessFrame 立即处理待处理事件、更新文档并绘制一帧（平台循环每帧做的事）；典型用法：窗口还隐藏时先渲染首帧再 Show。
func (w *Window) ProcessFrame() error {
	return statusError(C.kui_window_process_frame(w.raw()), "Window.ProcessFrame")
}

// SetDevToolsTokenAuth 打开 / 关闭 DevTools WebSocket 的令牌校验；DevTools 服务必须已启动，否则 KUI_ERR_NOT_FOUND。
// 关掉之后任何本机客户端都能连，只给无法带令牌的工具用。
func (w *Window) SetDevToolsTokenAuth(enabled bool) error {
	return statusError(C.kui_window_set_devtools_token_auth(w.raw(), boolInt(enabled)), "Window.SetDevToolsTokenAuth")
}

// DisableIME 关闭调用线程的系统输入法，字母键和空格不再被中日韩输入法截走；仅 Windows。
func DisableIME() error {
	return statusError(C.kui_platform_disable_ime(), "DisableIME")
}

// LogLevel 是控制台消息级别。
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// SetConsoleEnabled 打开 / 关闭进程内控制台（DevTools 里看到的日志）。
func SetConsoleEnabled(enabled bool) error {
	return statusError(C.kui_console_set_enabled(boolInt(enabled)), "SetConsoleEnabled")
}

// SetConsoleStderrOutput 让每条消息同时实时打印到 stderr。
func SetConsoleStderrOutput(enabled bool) error {
	return statusError(C.kui_console_set_stderr_output(boolInt(enabled)), "SetConsoleStderrOutput")
}

// SetConsoleMinLevel 低于该级别的消息被丢弃。
func SetConsoleMinLevel(level LogLevel) error {
	return statusError(C.kui_console_set_min_level(C.KuiCLogLevel(level)), "SetConsoleMinLevel")
}

// 堆栈追踪输出位。
const (
	StackTraceFile       = 1
	StackTraceStderr     = 2
	StackTraceMessageBox = 4
)

// StackTraceConfig 零值用库默认值（文件 + stderr、64 帧）。
type StackTraceConfig struct {
	OutputMode    uint
	LogFilePath   string
	MaxFrames     int
	AppName       string
	WriteMiniDump bool
	DumpDir       string
}

// InstallStackTrace 安装未处理异常 / 信号的堆栈追踪写出；返回是否装上。尽早调用（启动线程之前）。
func InstallStackTrace(config StackTraceConfig) (bool, error) {
	logFile, _, releaseLog := inputString(config.LogFilePath)
	defer releaseLog()
	appName, _, releaseApp := inputString(config.AppName)
	defer releaseApp()
	dumpDir, _, releaseDump := inputString(config.DumpDir)
	defer releaseDump()
	raw := C.KuiCStackTraceConfig{
		output_mode: C.uint(config.OutputMode), log_file_path: logFile, max_frames: C.int(config.MaxFrames),
		app_name: appName, write_mini_dump: boolInt(config.WriteMiniDump), dump_dir: dumpDir,
	}
	raw.struct_size = C.size_t(unsafe.Sizeof(raw))
	var installed C.int
	err := statusError(C.kui_stack_trace_install(&raw, &installed), "InstallStackTrace")
	return installed != 0, err
}

// UninstallStackTrace 卸载堆栈追踪写出。
func UninstallStackTrace() error {
	return statusError(C.kui_stack_trace_uninstall(), "UninstallStackTrace")
}

// CrashReporterConfig 零值用库默认值（exe 旁的 crashpad_handler，不上传）。
type CrashReporterConfig struct {
	HandlerExe, DumpDir string
	AppName, AppVersion string
	ReportURL           string
	EnableUpload        bool
}

// InstallCrashReporter 安装 Crashpad 崩溃处理；返回是否装上（比如找不到 crashpad_handler 时为 false，调用本身仍成功）。
func InstallCrashReporter(config CrashReporterConfig) (bool, error) {
	handler, _, releaseHandler := inputString(config.HandlerExe)
	defer releaseHandler()
	dumpDir, _, releaseDump := inputString(config.DumpDir)
	defer releaseDump()
	appName, _, releaseApp := inputString(config.AppName)
	defer releaseApp()
	appVersion, _, releaseVersion := inputString(config.AppVersion)
	defer releaseVersion()
	reportURL, _, releaseURL := inputString(config.ReportURL)
	defer releaseURL()
	raw := C.KuiCCrashReporterConfig{
		handler_exe: handler, dump_dir: dumpDir, app_name: appName, app_version: appVersion,
		report_url: reportURL, enable_upload: boolInt(config.EnableUpload),
	}
	raw.struct_size = C.size_t(unsafe.Sizeof(raw))
	var installed C.int
	err := statusError(C.kui_crash_reporter_install(&raw, &installed), "InstallCrashReporter")
	return installed != 0, err
}

// UninstallCrashReporter 卸载崩溃处理。
func UninstallCrashReporter() error {
	return statusError(C.kui_crash_reporter_uninstall(), "UninstallCrashReporter")
}
