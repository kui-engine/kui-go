# KUI Go 快速上手

[English](QUICKSTART.md) · **简体中文**

在一个空的 Go 项目里，用一个 HTML 文件做出原生窗口：一张带小表单和计数器的卡片，每点一次按钮，计数加一。

<p align="center"><img src="docs/media/quickstart.gif" width="360" alt="快速开始窗口的录屏：输入姓名、选套餐和日期、切换选项、拖动滑块，再点三次按钮"></p>
<p align="center"><sub>真实窗口的录屏：改表单，再点三次按钮。</sub></p>

页面就是普通的 HTML 和 CSS。除了计数器，还有一小块表单（文本框、下拉选择、日期选择、复选框、单选和滑块）。这些控件由 KUI 自己绘制和处理，不需要写代码：上面的录屏里依次输入了姓名、选了套餐和日期、切换了选项、拖了滑块，又点了按钮。

完整 API、构建脚本和全部示例见 [`README.md`](README.md) 与 [`GUIDE.md`](GUIDE.md)。

## 1. 需要什么

本页在 Windows 10 x64 上用以下版本验证过：

- Go 1.26.7（SDK 的 `go.mod` 声明 `go 1.22`，更低的版本没试过）
- LLVM clang 22.1.6 作为 cgo 的 C 编译器（不需要 MinGW），Python 3.13.3 用来运行 clang 包装脚本
- Git 2.45.1

## 2. 拿到 SDK

运行库在仓库里，所以即使 Go 模块从别处取，也要克隆一份。在一个空目录里执行：

```powershell
mkdir kui-quickstart; cd kui-quickstart
git clone --depth 1 https://github.com/kui-engine/kui-go.git
$KUI = "$PWD\kui-go"
mkdir hello-kui; cd hello-kui
```

固定版本的写法：`git clone --depth 1 --branch v3.1.0 https://github.com/kui-engine/kui-go.git`。

> 请保留 `--depth 1`。每个版本都把运行库提交进历史，完整历史很大。

## 3. 建项目

`hello-kui` 里放三个文件：`go.mod`、`main.go` 和 `app.html`。

`go.mod`：

```go-mod
module example.com/hello-kui

go 1.22

require github.com/kui-engine/kui-go/v3 v3.1.0

replace github.com/kui-engine/kui-go/v3 => ../kui-go
```

`replace` 这一行让编译直接用项目旁边的克隆：不下载，也不需要 `go.sum`。本页就是用这种写法实测的。另一种写法是删掉 `replace` 行，执行 `go get github.com/kui-engine/kui-go/v3@v3.1.0`，从 Go 代理取模块并写出 `go.sum`。

> `go get` 下载的模块 zip 里没有 `runtime/` 和 `tools/`，所以无论哪种写法，DLL 和 clang 包装脚本都要从克隆（或 Release 上的 SDK 包）里取。

`main.go`：

```go
package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	kui "github.com/kui-engine/kui-go/v3"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	check(kui.Load("")) // "" = kui_c.dll next to this executable
	check(kui.Init())   // locks this goroutine to the UI thread
	defer kui.Shutdown()

	window, err := kui.NewWindow(kui.WindowProps{
		Title: "KUI Quick Start", Width: 560, Height: 640, NoResize: true, Centered: true,
	})
	check(err)
	defer window.Close() // destroys the window and releases its document

	doc, err := window.Document() // owned by the window, do not Close it yourself
	check(err)
	exe, err := os.Executable()
	check(err)
	check(doc.LoadFile(filepath.Join(filepath.Dir(exe), "app.html")))

	count, err := doc.GetElementByID("count")
	check(err)
	defer count.Close()
	button, err := doc.GetElementByID("inc")
	check(err)
	defer button.Close()               // also removes the click listener below
	if count == nil || button == nil { // GetElementByID returns nil, nil for an unknown id
		log.Fatal("app.html has no #count or #inc")
	}

	clicks := 0
	_, err = button.AddEventListener("click", func(*kui.Event) {
		clicks++
		_ = count.SetTextContent(strconv.Itoa(clicks)) // the DOM update
	})
	check(err)

	check(window.Show())
	_, err = kui.Run() // returns when the window is closed
	check(err)
}
```

`app.html` 是界面。程序只依赖其中两个 id：`#count`（数字）和 `#inc`（按钮）。

<details>
<summary>app.html（界面，复制即可）</summary>

```html
﻿<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>KUI Quick Start</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  html, body { height: 100%; }
  body {
    display: flex; align-items: center; justify-content: center;
    font-family: "Segoe UI", system-ui, -apple-system, "Helvetica Neue", Arial, sans-serif;
    font-size: 14px; color: #e2e8f0;
    background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 55%, #312e81 100%);
  }
  .card {
    display: flex; flex-direction: column;
    width: 480px; padding: 28px 30px 20px;
    border-radius: 24px;
    background: rgba(255, 255, 255, 0.07);
    border: 1px solid rgba(255, 255, 255, 0.14);
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.45);
  }
  .head { display: flex; flex-direction: column; align-items: center; text-align: center; }
  .badge {
    display: inline-block; padding: 4px 14px; border-radius: 999px;
    font-size: 12px; font-weight: 600; letter-spacing: 2px;
    color: #c7d2fe; background: rgba(99, 102, 241, 0.28);
  }
  h1 { margin-top: 12px; font-size: 26px; font-weight: 700; color: #ffffff; }
  .sub { margin-top: 6px; font-size: 14px; color: #94a3b8; }

  .field { display: flex; flex-direction: column; margin-top: 16px; }
  .field label { margin-bottom: 6px; font-size: 12px; font-weight: 600; letter-spacing: 0.5px; color: #a5b4fc; }
  .pair { display: flex; gap: 14px; }
  .pair .field { flex: 1; }
  input[type="text"], input[type="date"], select {
    width: 100%; padding: 9px 12px; font-size: 14px; color: #f1f5f9;
    background: rgba(15, 23, 42, 0.55);
    border: 1px solid rgba(165, 180, 252, 0.28); border-radius: 10px;
  }
  input[type="text"]:focus, input[type="date"]:focus, select:focus { border-color: #818cf8; outline: none; }

  .opts { display: flex; align-items: center; justify-content: space-between; margin-top: 16px; }
  .opts label { display: flex; align-items: center; font-size: 14px; color: #cbd5e1; }
  .opts input { margin-right: 8px; }
  .radios { display: flex; gap: 16px; }
  input[type="range"] { width: 100%; }

  .action { display: flex; align-items: center; justify-content: space-between; margin-top: 22px; padding-top: 18px;
            border-top: 1px solid rgba(255, 255, 255, 0.1); }
  .counter { display: flex; align-items: baseline; }
  #count { font-size: 44px; font-weight: 800; line-height: 1; color: #a5b4fc; }
  #label { margin-left: 8px; font-size: 14px; color: #94a3b8; }
  #inc {
    padding: 12px 30px; border: 0; border-radius: 14px;
    font-size: 16px; font-weight: 600; color: #ffffff;
    background: linear-gradient(135deg, #6366f1, #8b5cf6);
    box-shadow: 0 10px 24px rgba(99, 102, 241, 0.45);
    transition: transform 0.15s, box-shadow 0.15s;
  }
  #inc:hover { transform: translateY(-2px); box-shadow: 0 14px 30px rgba(99, 102, 241, 0.6); }
  #inc:active { transform: translateY(1px); }
  .foot { margin-top: 18px; text-align: center; font-size: 12px; color: #64748b; }
</style>
</head>
<body>
  <div class="card">
    <div class="head">
      <span class="badge">KUI</span>
      <h1>Hello from KUI</h1>
      <p class="sub">A native window drawn from HTML and CSS.</p>
    </div>

    <div class="field">
      <label>NAME</label>
      <input id="name" type="text" value="Ada Lovelace">
    </div>
    <div class="pair">
      <div class="field">
        <label>PLAN</label>
        <select id="plan">
          <option>Free</option>
          <option selected>Pro</option>
          <option>Team</option>
        </select>
      </div>
      <div class="field">
        <label>START DATE</label>
        <input id="start" type="date" value="2026-12-01">
      </div>
    </div>
    <div class="opts">
      <label><input id="news" type="checkbox" checked>Email me updates</label>
      <div class="radios">
        <label><input type="radio" name="billing" checked>Monthly</label>
        <label><input type="radio" name="billing">Yearly</label>
      </div>
    </div>
    <div class="field">
      <label>VOLUME</label>
      <input id="volume" type="range" min="0" max="100" value="50">
    </div>

    <div class="action">
      <div class="counter"><div id="count">0</div><div id="label">clicks</div></div>
      <button id="inc">Click me</button>
    </div>
    <div class="foot">One engine · Windows, Linux, macOS</div>
  </div>
</body>
</html>
```

</details>

## 4. 运行库放哪

Windows 运行库在 `$KUI/runtime/windows-x64/bin`。`kui.Load("")` 在可执行文件旁边找 `kui_c.dll`，所以构建步骤把这个目录连同 `app.html` 一起复制到 `hello-kui.exe` 旁边。本程序只用到 `kui_c.dll` 和 `kui.dll`，其余文件是 3D、编辑器、PDF 扩展、媒体插件和 FFmpeg。

> 不想复制 DLL 时，改传完整路径：``kui.Load(`<克隆目录>\kui-go\runtime\windows-x64\bin\kui_c.dll`)``（把 `<克隆目录>` 换成你克隆到的文件夹），`kui.dll` 会在同一目录里找到。

## 5. 最小窗口

`kui.NewWindow` 创建窗口，`Width`、`Height` 是客户区大小，这里是 560×640。`doc.LoadFile` 加载 `app.html`，`kui.Run` 跑消息循环，直到窗口被关掉。

## 6. 一个 DOM 更新

`doc.GetElementByID("count")` 返回 `*Element`，`SetTextContent` 替换它的文字。本例在点击回调里调用它，屏幕上的数字跟着 `clicks` 走。

## 7. 一个事件监听

`button.AddEventListener("click", fn)` 在事件派发时于 UI 线程上调用 `fn`，回调里可以改 DOM。

## 8. 资源释放

- `kui.Init` 与 `kui.Shutdown` 在同一个 goroutine 里调用，UI 调用也留在这里；其他 goroutine 用 `kui.PostTask` 把工作投递过来。
- 窗口拥有自己的文档。`window.Close()` 销毁窗口，并释放 `window.Document()` 返回的文档；之后不要再用 `doc`。
- 每个 `*Element` 各持有一份引用：`Close` 一次，之后不能再用。元素的 `Close` 还会移除经它添加的监听，所以 `button` 一直留到 `Run` 返回。
- 传给监听回调的 `*kui.Event` 是借用的，只在回调内有效。
- `defer` 倒序执行：依次释放 `button`、`count`、窗口，最后关闭平台。

## 9. 构建与运行

Windows PowerShell，在 `hello-kui` 里：

```powershell
$env:CGO_ENABLED = '1'
$env:CC = "python `"$KUI\tools\clang_go_cgo.py`""
go build -o build\hello-kui.exe .
Copy-Item $KUI\runtime\windows-x64\bin\*, app.html build\
.\build\hello-kui.exe
```

窗口显示 0，每点一次加一。关掉窗口程序就结束。

> 本页只在 Windows x64 上验证过。Linux 与 macOS 上 cgo 用系统 C 编译器（不要设 `CC`），运行库在 `$KUI/runtime/linux-x64/lib` 或 `$KUI/runtime/macos-arm64/lib`，程序没有 `.exe` 后缀。macOS 上 UI 必须留在主 goroutine，也就是 `main` 里调 `kui.Init` 的地方。这两个平台验证过的命令见 [`README.md`](README.md) 的「编译并运行一个示例」。

## 10. 常见问题

- **程序退出并提示 `app.html has no #count or #inc`**：`app.html` 不在可执行文件旁边。本版 `LoadFile` 不会报告文件不存在（文档只是保持为空），而且相对路径按当前目录解析，不按可执行文件所在目录；所以 `main.go` 用 `os.Executable()` 拼出完整路径。
- **`cgo: C compiler "gcc" not found`**：当前终端没设 `CC`，按第 9 步设置。
- **编译报 `undefined: Document` 之类**：`CGO_ENABLED` 是 `0`。绑定需要 cgo，设成 `1`。
- **`kui: Load: LoadLibraryExW failed (win32=126)`**：可执行文件旁边没有 `kui_c.dll` 或 `kui.dll`；复制运行库目录，或给 `kui.Load` 传完整路径。
- **`go build` 时出现一串 `lld-link: warning: section name .debug_... is longer than 8 characters`**：无害，可执行文件照常生成、能运行。
