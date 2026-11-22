> 原样取自 Go SDK 包（`kui-go-sdk-3.1.0-<平台>.zip`）的 README：文中包里的 `native/bin`（Windows）/ `native/lib`（Linux / macOS）是本仓库的 `runtime/windows-x64/bin` / `runtime/<平台>/lib`，`bin/` 是 `build.ps1` / `build.sh` 的输出目录。在本仓库里跑这两个脚本不用参数（默认会找 `runtime/<平台>/`），见 [README.md](README.md)。
> The Go SDK package README as is: its `native/bin` (Windows) / `native/lib` (Linux / macOS) are this repository's `runtime/windows-x64/bin` / `runtime/<platform>/lib`, and `bin/` is the output folder of `build.ps1` / `build.sh`. In this repository the scripts need no parameters (they find `runtime/<platform>/`), see [README.md](README.md).

# KUI Go SDK 3.1.0

状态：Windows x64 Full Profile GA；Linux x64 / macOS arm64 同口径 Release 包（3.61.8 起）。

本模块通过冻结的 KUI C ABI 工作。生成器会从 `kui_c.h` 同步并校验全部
冻结清单内的全部 product-major-3 导出；运行时加载阶段只有在全部符号解析后才成功。
Go 层提供资源拥有权、错误、UTF-8 字符串和回调桥接，不跨 DLL 传递 C++ STL。
除全部符号的加载完整性外，Go 门面也逐项实际引用全部公开函数；低频的 retain、
event target、文件对话框和窗口状态查询同样不是占位声明。

覆盖范围包括 Window、Document/DOM、style/geometry、events、DevTools、平台服务、
Lua、Canvas2D/WebGL 脚本执行、HTTP、media/Lottie、custom elements、PDF，以及
extension/3D 通用调用。3D 能力仍要求单独部署 `kui_ext_3d.dll`；媒体解码仍要求
对应 `kui_media_*` 后端，这些是 Full Profile 的运行时组件，不会被 Go 重写。

## 构建

要求：Windows x64、Go 1.22+、Python 3、LLVM Clang，以及运行库 `kui_c.dll`、`kui.dll` 与扩展 / 媒体插件。`build.ps1` 不给参数时
依次找 SDK 包里的 `native/bin`、对外仓库（`git clone kui-go`）里的 `runtime/windows-x64/bin`、KUI 仓库内开发的
`../c/build/{Release,Debug}`，所以解压包或 clone 仓库后都可以直接运行；都找不到时报错并列出试过的目录，
`-RuntimeDir` / `-FullProfileRuntimeDir` 可显式指定。

```powershell
pwsh -File build.ps1 -Configuration Release -Output bin
```

脚本先运行 `go test ./...`，再精确编译 8 个 quickstart 和 225 个 canonical
example（211 个与 C++ 原版一一对应 + 14 个 API 巡览；数量读 `examples/canonical/manifest.json`），并把页面、素材和 Full Profile runtime 复制到输出目录。最终 `bin/`
固定包含 233 个 `kui_go_*.exe`。官方 MSVC-target Clang 不接受 Go 固定传入的 MinGW 参数 `-mthreads`，
`tools/clang_go_cgo.py` 移除该参数（以及 Go 外部链接器重复传入的 GNU strip 开关 `-s`）并原样转发其余参数和退出码，因此无需额外安装
MinGW，也保持与 KUI C ABI 一致的 Windows ABI。

默认输出就是仓库内 `sdk/go/bin/`。该目录包含 233 个 EXE、assets 和 Full Profile
runtime，是可直接运行的本地 SDK 产物；它属于可重建二进制，因此由 Git 忽略。

`build.ps1 -Configuration Debug` 默认输出到 `bin\debug`：包里只有 Release 核心（`native/bin`，
2026-11-20 起不再带 `native/bin/debug`），Debug 构建同样用它（C ABI 不传 CRT 对象）；仓库内开发取
`../c/build/Debug`。扩展与媒体运行时用 Release 版；`-FullProfileRuntimeDir` 可另指扩展运行时目录。Debug 构建保留
DWARF，可用 Delve 调试。

直接开发时：

```powershell
$env:CGO_ENABLED = '1'
# 必须是绝对路径：cgo 编译 GOROOT 里的 runtime/cgo 时工作目录不在这里，相对路径会找不到
$env:CC = "python `"$PWD\tools\clang_go_cgo.py`""
go test ./...
go build .\examples\hello_world
```

程序启动时必须先 `kui.Load()` 指向 `kui_c.dll`，然后 `kui.Init()`；UI API 应在
初始化线程调用，跨 goroutine 工作通过 `PostTask` 回到 UI 线程（`PostTask` 可以从任何
goroutine 调用）。`HTTPClient.FetchAsync` 的回调在 HTTP 工作线程上执行，要操作 DOM 时改用
`FetchAsyncOnMainThread`，或在回调里 `PostTask`。拥有型对象支持
显式 `Close()`，建议始终使用 `defer`，不要依赖 GC 决定原生资源释放时机。

## Examples

SDK 固定交付 8 个 quickstart，以及 225 个 canonical 源码与页面：211 个与正式 C++ SDK example
产品一一对应，另有 14 个 API 巡览示例。canonical 入口位于
`examples/canonical/generated/`，由 `scripts/generate_go_rust_examples.py` 以 C
canonical manifest 为权威生成。

**预编译程序在单独的示例包里**（2026-11-20 起 SDK 包不带 `bin/`）：`kui-go-examples-<版本>-<平台>.zip`（同一个 GitHub
Release）的 `bin/` 就是上面 `build.ps1` / `build.sh` 的输出——233 个程序、与本包 `native/` 逐字节相同的运行库和 assets（另带 `kui_script_example` 用的 `script_html/`），
解压即可运行；也可以用本包的源码自己编出同样的 `bin/`。程序默认从 EXE 同目录加载 `kui_c.dll`；quickstart 仍接受命令行路径
作为显式覆盖：

```powershell
.\bin\kui_go_hello_world.exe
.\bin\kui_go_canonical_native_3d_engine_01_hello_triangle.exe
```

### Linux x64 / macOS arm64

要求：Go 1.22+ 与系统 C 编译器（cgo；Linux 用 gcc 或 clang，macOS 用 Xcode 命令行工具），以及运行库
`libkui_c.so|.dylib`、`libkui.so|.dylib` 与扩展 / 媒体插件。`build.sh` 不给参数时依次找 SDK 包里的 `native/lib`、对外仓库里的
`runtime/<linux-x64|macos-arm64>/lib`、KUI 仓库内开发的 `../c/build`（按本版 C++ staging 构建的 sdk/c）与 `../cpp/lib`；`-r` / `-f` 可显式指定。

```bash
./build.sh -o bin        # Linux 没有桌面时：xvfb-run -a ./build.sh -o bin
```

与 `build.ps1` 同一组步骤与数量门：`go test ./...`、8 个 quickstart 与 225 个 canonical example，`bin/` 里是 233 个
无扩展名的程序，旁边放着 `libkui_c`、`libkui`、插件与 assets。只有 Release（Windows 的 Debug 变体是为了 MSVC Debug CRT）。
`Load("")` 在可执行文件所在目录找 `RuntimeLibraryName`（`kui_c.dll` / `libkui_c.so` / `libkui_c.dylib`），找不到再交给
系统的库搜索规则；`libkui_c` 自己的 RPATH 是 `$ORIGIN`（`@loader_path`），会在同一目录找到 `libkui`。

macOS 只允许在主线程初始化平台、创建窗口：在 `main` 开头、主 goroutine 里调用 `Init()`（示例都是这样写的），之后的窗口
与 DOM 操作留在这个 goroutine 上，别的 goroutine 用 `PostTask` 投递。

## 发布边界

- SDK 版本线为 3.1.0；底层 KUI 产品版本独立记录在发布包清单中。
- C ABI 导出集合是兼容性基线；生成文件禁止手改。
- GA 声明适用于 Windows x64、Linux x64 与 macOS arm64 的 Full Profile；Linux / macOS 包只有 Release。
  `HTTPClient` 在 Linux 上需要发行版的 libcurl（`libcurl.so.4` 或 `libcurl-gnutls.so.4`，运行期加载；没有时请求如实报错），
  macOS `.app` 包里请求明文 `http://` 域名需要宿主在 `Info.plist` 配置 App Transport Security 例外。
- SDK 包不带预编译程序与 `bin/`（`SDK_MANIFEST.json` 的 `prebuilt_bin` 为 `false`，仍列出构建脚本会产出的 233 个程序名）；
  它们在 `kui-go-examples-*` 示例包里，示例包另带同格式的 `SDK_MANIFEST.json`（`prebuilt_bin` 为 `true`）。
- 包里还带 `tools/devtools_mcp/`（调试运行中的 KUI 程序，用法见其 README）和 `docs/`（14 份与语言无关的文档，索引在 `docs/README.md`）；C++ 专属的入门、教程与组件 / 扩展编写指南只在 C++ 包里。
- 网络 URL、媒体文件、3D extension 等外部资源是否可用仍取决于部署与后端。
