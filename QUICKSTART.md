# KUI Go quickstart

**English** · [简体中文](QUICKSTART.zh-CN.md)

Build a native window from an HTML file in an empty Go project: a card with a small form and a counter that goes up each time you click the button.

<p align="center"><img src="docs/media/quickstart.gif" width="360" alt="The quickstart window, recorded: typing a name, picking a plan and a date, switching options, dragging the slider and clicking the button three times"></p>
<p align="center"><sub>A real window, recorded: edit the form, then click the button three times.</sub></p>

The page is ordinary HTML and CSS. Next to the counter it has a small form (text input, select, date picker, checkbox, radio buttons and a slider). KUI draws and handles those controls itself, so they work without any code: the recording above types a name, picks a plan and a date, switches the options, drags the slider and clicks the button.

The full API, the build scripts and all examples are in [`README.md`](README.md) and [`GUIDE.md`](GUIDE.md).

## 1. What you need

This page was verified on Windows 10 x64 with these versions:

- Go 1.26.7 (the SDK's `go.mod` declares `go 1.22`; older versions were not tried)
- LLVM clang 22.1.6 as cgo's C compiler (no MinGW needed) and Python 3.13.3 to run the clang wrapper
- Git 2.45.1

## 2. Get the SDK

The runtime libraries live in the repository, so clone it even if you take the Go module from elsewhere. In an empty folder:

```powershell
mkdir kui-quickstart; cd kui-quickstart
git clone --depth 1 https://github.com/kui-engine/kui-go.git
$KUI = "$PWD\kui-go"
mkdir hello-kui; cd hello-kui
```

To pin a version: `git clone --depth 1 --branch v3.1.0 https://github.com/kui-engine/kui-go.git`.

> Keep `--depth 1`. Every version commits its runtime libraries, so the full history is large.

## 3. Create the project

`hello-kui` holds three files: `go.mod`, `main.go` and `app.html`.

`go.mod`:

```go-mod
module example.com/hello-kui

go 1.22

require github.com/kui-engine/kui-go/v3 v3.1.0

replace github.com/kui-engine/kui-go/v3 => ../kui-go
```

The `replace` line builds against the clone next to the project: nothing is downloaded and no `go.sum` is needed. This page was tested with this form. The alternative is to delete the `replace` line and run `go get github.com/kui-engine/kui-go/v3@v3.1.0`, which takes the module from the Go proxy and writes `go.sum`.

> The module zip that `go get` downloads has no `runtime/` and no `tools/`, so you need the clone (or the SDK package on the Release) for the DLLs and the clang wrapper either way.

`main.go`:

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

`app.html` is the interface. The program only relies on two ids in it: `#count` (the number) and `#inc` (the button).

<details>
<summary>app.html (the interface, copy as is)</summary>

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

## 4. Where the runtime goes

The Windows runtime is `$KUI/runtime/windows-x64/bin`. `kui.Load("")` looks for `kui_c.dll` next to the executable, so the build step copies that folder, together with `app.html`, next to `hello-kui.exe`. This program needs only `kui_c.dll` and `kui.dll`; the other files are the 3D, editor and PDF extensions, the media plugins and FFmpeg.

> To leave the DLLs where they are, pass the full path instead: ``kui.Load(`<clone>\kui-go\runtime\windows-x64\bin\kui_c.dll`)`` (with `<clone>` replaced by the folder you cloned into). `kui.dll` is then found in the same folder.

## 5. Minimal window

`kui.NewWindow` creates the window; `Width` and `Height` are the client area, 560×640 here. `doc.LoadFile` loads `app.html`, and `kui.Run` runs the message loop until the window is closed.

## 6. A DOM update

`doc.GetElementByID("count")` returns an `*Element`; `SetTextContent` replaces its text. Here it runs inside the click handler, so the number on screen follows `clicks`.

## 7. An event listener

`button.AddEventListener("click", fn)` calls `fn` on the UI thread while the event is dispatched. Changing the DOM inside `fn` is allowed.

## 8. Releasing resources

- Call `kui.Init` and `kui.Shutdown` on the same goroutine and keep UI calls there; other goroutines hand work over with `kui.PostTask`.
- The window owns its document. `window.Close()` destroys the window and releases the document returned by `window.Document()`; do not use `doc` after that.
- Each `*Element` holds its own reference: `Close` it once, then do not use it again. `Close` on an element also removes the listeners added through it, so `button` stays open until `Run` returns.
- The `*kui.Event` passed to a listener is borrowed and valid only inside the callback.
- `defer` runs in reverse: `button`, `count`, the window, then the platform.

## 9. Build and run

Windows PowerShell, inside `hello-kui`:

```powershell
$env:CGO_ENABLED = '1'
$env:CC = "python `"$KUI\tools\clang_go_cgo.py`""
go build -o build\hello-kui.exe .
Copy-Item $KUI\runtime\windows-x64\bin\*, app.html build\
.\build\hello-kui.exe
```

The window shows 0; every click adds one. Closing the window ends the program.

> This page was only verified on Windows x64. On Linux and macOS cgo uses the system C compiler (do not set `CC`), the runtime is `$KUI/runtime/linux-x64/lib` or `$KUI/runtime/macos-arm64/lib`, and the program has no `.exe`. On macOS the UI must stay on the main goroutine, which is where `main` calls `kui.Init`. Commands verified on those platforms are in "Building and running an example" in [`README.md`](README.md).

## 10. Common problems

- **The program exits with `app.html has no #count or #inc`**: `app.html` is not next to the executable. `LoadFile` does not report a missing file in this version (the document just stays empty), and a relative path is resolved against the current directory, not the executable's folder; that is why `main.go` builds the full path from `os.Executable()`.
- **`cgo: C compiler "gcc" not found`**: `CC` is not set in this shell; set it as in step 9.
- **`undefined: Document` and similar when building**: `CGO_ENABLED` is `0`. The binding needs cgo; set it to `1`.
- **`kui: Load: LoadLibraryExW failed (win32=126)`**: `kui_c.dll` or `kui.dll` is not next to the executable; copy the runtime folder or pass the full path to `kui.Load`.
- **Many `lld-link: warning: section name .debug_... is longer than 8 characters` lines during `go build`**: harmless; the executable is built and runs.
