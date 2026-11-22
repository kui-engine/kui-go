# KUI Go SDK Changelog

## [Unreleased]

### Changed

- canonical 示例增至 225 个（211 个与 C++ 原版一一对应 + 14 个 API 巡览示例），`bin/` 共 233 个程序；`build.ps1` / `build.sh`
  的数量门改读包内 `examples/canonical/manifest.json`，不再写死。预编译示例包的 `bin/` 另带 `kui_script_example` 用的 `script_html/`。
- `build.ps1` / `build.sh` 不给运行库参数时，依次找 SDK 包布局 `native/bin`（Linux / macOS `native/lib`）、对外仓库（`git clone kui-go`）
  布局 `runtime/<平台>/bin|lib`、KUI 仓库内开发布局；`git clone` 后直接 `./build.sh` / `build.ps1` 就能构建全部示例。都找不到时报错并列出
  试过的目录；`-RuntimeDir` / `-FullProfileRuntimeDir`（`-r` / `-f`）仍然优先。`kui.Load(path)` 的行为不变。

### Fixed

- `ScriptEngine.Eval` 以前用 4 KB 缓冲接 Lua 错误文本，更长的错误让 `Eval` 返回 `ErrBufferTooSmall`
  而不是结果；现在读完整的错误文本。
- `PostTask` 现在真正可以从任何 goroutine 调用：以前它走 `kui_platform_request_idle`，
  定时器挂在调用线程的消息队列上（并且无锁访问内部表），从非 UI goroutine 调用时回调
  可能永不执行或跑在错误的线程上——而包文档正是这样教的。改用
  `kui_platform_post_to_main_thread`；`Init` 之前或 `Shutdown` 开始之后调用返回
  `ErrNotSupported`，关闭时还在排队的任务不执行，handle 照常释放。
- `build.ps1 -Configuration Debug` 改用「Debug 核心（`kui_c.dll`、`kui.dll`）+ Release 扩展运行时」组合，与 C++ / C SDK 的部署约定相同。以前解包后的包用 Debug 构建直接失败（Debug 目录里没有扩展），仓库内 Debug 的 `kui.dll` 又会被 Release 版覆盖。Debug 默认输出改为 `bin\debug`，不再覆盖 `bin/` 里的标准 Release 产物；Debug 构建保留 DWARF（`-gcflags=all=-N -l`，不再 `-s -w`），可以用 Delve 调试。

### Added

- 宿主集成：`WindowProps` 的 `Sizeable` / `Hidden` / `FrameRadius`、`Window.SetEnabled` / `IsEnabled` / `ProcessFrame` / `SetDevToolsTokenAuth`、`Element.StartMomentumScroll` / `ScrollIntoView`、`ScriptEngine.TickTimers`、`Event.WheelDelta`、`Element.AddCaptureEventListener`（捕获阶段监听）、`Document.SetBasePath` / `BasePath` / `SetRenderDevicePixelRatio`、`SetConsole*`、`InstallStackTrace` / `InstallCrashReporter`、`DisableIME`。
- 补齐缺口：`Observable.Clone` / `Observer.Clone`（对应 retain）、`Router.CurrentPath`、`TreeView.UpdateNode`（`TreeNodeUpdate`，为 nil 的字段保持原值）。
- `Event.RelatedTarget`：鼠标与焦点事件的 `relatedTarget`；`<kui-context-menu>` 的 `menu-open` 里是被右键的元素。
- `Popups.SetDefaultContextMenuEnabled` / `DefaultContextMenuEnabled`：引擎自带的默认右键菜单 3.61.8 起默认不显示，
  用它开启（页面也可以写 `<html context-menu>`）。
- Linux x64 / macOS arm64：全部实现文件的构建约束放宽为 `(windows && amd64) || (linux && amd64) || (darwin && arm64)`，
  加载器在这两个平台上用 `dlopen`（`Load("")` 先找可执行文件所在目录，再交给系统搜索规则）；新增常量
  `RuntimeLibraryName`（`kui_c.dll` / `libkui_c.so` / `libkui_c.dylib`）。包内 `build.sh` 与 `build.ps1` 同一组步骤
  （`go test`、8 个 quickstart、211 个 canonical 示例）；只有 Release。`HTTPClient` 在这两个平台上由内核的 libcurl（Linux，
  运行期加载发行版的 `libcurl.so.4`）/ NSURLSession（macOS）实现，语义与 Windows 相同（`T-DEBT-POSIX-HTTPCLIENT-NOT-IMPLEMENTED`）；
  `TestHTTPBodyUsesABIBufferCapacity` 三个平台同一组断言。
- `HTTPClient.FetchAsyncOnMainThread`：回调在 UI 线程执行。`FetchAsync` 的回调一直在
  HTTP 工作线程上，文档现在写明。
- 常用窗口、DOM 与平台操作：`Window.Focus` / `HasFocus` / `SetCursor` / `SetCursorVisible` /
  `SetFullscreen` / `IsFullscreen` / `SetFrameless` / `SetFramelessShadow` / `FrameState` /
  `SetFrameRadii` / `FrameRadii` / `SetIMECandidatePos`；`Element.InsertBefore` / `ReplaceChild` /
  `CloneNode` / `Matches` / `Closest` / `OuterHTML` / `InsertAdjacentHTML` / `AppendText` /
  `HasFocus` / 指针捕获；`Document.GetElementsByClassName` / `GetElementsByTagName` /
  `ElementFromPoint`；`MessageBox`、`OpenURL`、`Now`、`MonitorFromPoint`、`MonitorFromWindow`。
- 桌面集成：`Window.SetMenuBar` / `ClearMenuBar` / `SetMenuItemChecked` / `SetMenuItemEnabled`
  （`Menu` / `MenuItem`）；托盘 `CreateTray` / `DestroyTray` / `SetTrayVisible` / `SetTrayTooltip` /
  `SetTrayIcon` / `SetTrayContextMenu` / `PopupTrayContextMenu` / `ShowTrayBalloon` /
  `SetTrayEventHandler` / `SimulateTrayEvent`；`BindShortcut` / `UnbindShortcut` /
  `UnbindAllShortcuts` / `SetShortcutsEnabled`；`Document.ShowToast` / `DismissToast` /
  `DismissAllToasts` / `ToastCount`；`Window.SaveScreenshot`。回调句柄在替换、移除或
  `Window.Close` 时释放。
- 应用框架：`Document.Router()`（`Router.AddRoute` / `RemoveRoute` / `Clear` / `HasRoute` /
  `RouteCount` / `RegisterView` / `UnregisterView` / `Push` / `Replace` / `Go` / `Back` / `Forward` /
  `Current` / `Param` / `History` / `AddGuard` / `RemoveGuard` / `OnNavigated` / `RemoveNavigated`，
  `RouteInfo`）；`NewObservable`（`Set` / `Get` / `Subscribe` / `Unsubscribe` / `Bind` / `Unbind` /
  `UnbindAll` / `Close`）与 `Document.FlushBindings`；`Element.AnimationState`、
  `Document.AnimationCounts`；`NewMutationObserver` / `NewResizeObserver` /
  `NewIntersectionObserver`（`ObserveMutations` / `Observe` / `Unobserve` / `Disconnect` / `Close`）。
  回调句柄由 KUI 通知释放。
- `Element.ComputedStyleValue`：按 CSS 属性名读取计算后的值（等于
  `getComputedStyle(el).getPropertyValue(name)`）。
- 弹出层 / 选择器 / 音频：`Document.Popups()`（`ShowTooltip` / `ShowContextMenu` / `Dismiss` /
  `DismissAll` / `DismissType` / `Any` / `Has` / 提示延迟与配色 / `SetMode`）、`Element.OpenInputPicker` 与
  `Document.InputPicker()`（`Close` / `Commit` / `State` / `Target` / `SetEnabled` / `SetMonth` /
  `SetColorHSV` / `SetTime`）、`Document.Audio()`（`Available` / `Play` / `Stop` / `SetPaused` / 音量 /
  声像 / 音高 / `Seek` / `Info` / `SetGroup` / `Group` / `StopAll` / `SetAllPaused` / `Preload` /
  `Unload` / `UnloadAll` / `Counts`）。
- 媒体与 Lottie 播放控制：`Element.MediaLoad` / `SetMediaLoop` / `SetMediaAutoplay` / `MediaPlaybackFlags`，
  `Element.LottieStop` / `LottieSeek` / `SetLottieSpeed` / `SetLottieLoop` / `LottieState`（`LottieState`、
  `LottieLoadState`）。
- 扩展与 DevTools 生命周期：`Document.LoadExtension` / `UnloadExtension` / `ReloadExtension` /
  `UnloadAllExtensions` / `WatchExtensions` / `StopWatchingExtensions` / `ExtensionState`，
  `Window.StopDevTools` / `DevToolsState`。
- 拖放数据：`Event.SetDragData` / `DragData`（第二个返回值表示是否设置过该格式）/ `ClearDragData` /
  `ClearAllDragData` / `DragTypes` / `SetDropEffect` / `DropEffect` / `SetEffectAllowed` / `EffectAllowed`
  （`DropEffect`、`EffectAllowed` 类型）。
- 脚本引擎控制：`ScriptEngine.RegisterFunction` / `NewFunction`（Go 函数交给 Lua，返回 error 即 Lua 错误）/
  `CallFunction` / `EvalFile` / `SetGlobal` / `Global` / `SetPrintHandler` / `SetMemoryLimit` /
  `SetInstructionLimit` / `ResourceUsage`；`ScriptValue`（`ScriptNil` … `ScriptUserData`）、`ScriptFunction`
  （`Call` / `Clone` / `Close`）、`ScriptResources`；`ScriptResult` 新增 `Value`（第一个返回值）。
- 自定义元素自绘、tick 与输入：`CustomElementDefinition` 新增 `Paint` / `Tick` / `NeedsTick` / `Key` / `Mouse` /
  `Wheel` / `IME`；`PaintContext`（`FillRect` / `FillRoundedRect` / `StrokeRect` / `StrokeRoundedRect` / `DrawText` /
  `DrawPixels` / `DrawImage` / `DrawDecorationLine` / `Save` / `Restore` / `Translate` / `Scale` / `Rotate` /
  `SetOpacity` / `ClipRect` / `ClipRoundedRect`）、`TextStyle` / `TextMetrics` / `CornerRadii`；
  `Element.MeasureText` / `RequestRepaint` / `InvalidateRect`、`Document.DispatchIME`。
- 组件：`Document.RegisterBuiltinComponents`；`ThemeManager`（`RegisterTheme` / `Apply` / `ActiveTheme` /
  `ThemeNames` / `Theme`）；`NewVirtualList`（`VirtualListConfig`、`State` 等，`RenderItem` 返回的元素由 KUI 接管）；
  `NewListView`（`ListViewConfig` / `ListViewItem` / `SelectionMode`）；`NewTreeView`（`AddNode` / `AddChild` / `Node` /
  `SetNode` / `Setup(TreeViewConfig)` 与展开、选择、滚动方法）；`ScrollUnit`。
- 文档样式配置：`Document.SetStyleProfile` / `StyleProfile`（`StyleNativeGUI` / `StyleBrowserCompat`）。
- 结构化的低层接口：`Document.DispatchMouseEvent(MouseInjection)` / `DispatchKeyEvent(KeyInjection)`（修饰键位
  `InjectAlt` / `InjectCtrl` / `InjectMeta` / `InjectShift`）、`Document.ElementBoxModel` → `BoxModelInfo`（四层盒
  的顶点与内容尺寸）、`Event.FlagValues` → `EventFlags`、`Event.ModifierKeys` → `KeyModifiers`、
  `Document.DevToolsCall(method, params, result)`（结果解进结构体，命令错误为 `*DevToolsError`）。
- canonical 示例 runner 与 C++ 原版对齐：托盘（图标、菜单、单击切换、气泡、先缩到托盘）、多窗口一起关闭、
  html_loader 的浏览器默认值 / 内置组件 / 原生弹出层 / 三个列表；窗口选项照原版（不再给圆角示例设窗口圆角）。
- 需要 3.61.8 及以上的 `kui_c.dll`（新增导出）。
- `Element.Value / SetValue / Checked / SetChecked / SelectedIndex / SetSelectedIndex`：读写表单控件的当前状态（对应 C ABI 新增的 6 个导出）。

### Changed

- 弃用（保留一个版本）：`Document.DispatchMouse`（8 个位置参数）、`Document.DispatchKey`、`Document.BoxModel`
  （`[32]float32`）、`Event.Flags` / `Event.Modifiers`（裸 bool 元组），改用上面的结构化版本。
- README 更正：`tools/clang_go_cgo.py` 除 `-mthreads` 外还移除 Go 外部链接器重复传入的 `-s`。
- 随包的冻结头 `native/include/kui_c.h` 重新与权威 `sdk/c/include/kui_c.h` 同步：此前一直停在
  3.54.0（只差版本宏与几何取整注释，函数签名未变，绑定代码不变）。发版预检、打包与
  `scripts/check_sdk_versions.py` 现在都会校验它。
- 211 个 canonical 示例里有 47 个的窗口尺寸改为与 C++ 原版一致：生成器以前对写在共享头里、
  或用变量给出的尺寸一律退回 960×700（例如 3D 系列应为 1100×760）。manifest 同步更新。

### 更正

- 3.1.0 的发布说明写「正式包……进行解包重建和全量启动」，实际包验证只做了解包重建与
  数量检查，从未启动过这 219 个程序；全量启动门另行补上（仓库 `TODO.md` 的
  `T-DEBT-SDK-GORUST-NO-LAUNCH-GATE`）。

## [3.1.0] - 2026-08-26

- 首个 Windows x64 Full Profile GA 版本。
- 从权威 `kui_c.h` 生成动态加载桥，加载时完整校验 195 个 C ABI 导出。
- 提供 Window、Document/DOM/style/events、DevTools、Lua/Canvas/WebGL、HTTP、
  media/Lottie、custom elements、PDF 与 extension/3D Go API。
- 使用 `runtime/cgo.Handle` 完成 task/timer/event/resize/script/HTTP/custom element
  回调桥接，并提供确定性 `Close()`。
- 增加自托管 `build.ps1`、Clang cgo 适配器、运行时 smoke tests、8 个 quickstart
  和与 C++ 产品清单一一对应的 211 个 canonical examples；正式包交付全部 219 个
  标准预编译 EXE、页面、素材与 Full Profile runtime，并进行解包重建和全量启动。
- 模块路径升级为 `github.com/kui-engine/kui-go/v3`。

## [3.1.0-preview] - 2026-05-23

- 发布 API 设计草案；没有连接可执行 C ABI，不能用于开发或发布。
