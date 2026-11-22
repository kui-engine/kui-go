// Package runner 是全部 canonical Go examples（产品清单 + API 巡览）共用的 Full Profile 宿主。
package runner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	kui "github.com/kui-engine/kui-go/v3"
)

const (
	flagExtensions  = 1 << 0
	flagDevTools    = 1 << 1
	flagFrameless   = 1 << 2
	flagTransparent = 1 << 3
	flagAlwaysOnTop = 1 << 4
	// 1 << 5 以前是「圆角」；没有原版设置窗口圆角（见生成器）。
	flagMultiWindow   = 1 << 6
	flagLottie        = 1 << 7
	flagPDF           = 1 << 8
	flagTray          = 1 << 9
	flagShadow        = 1 << 10
	flagSizeable      = 1 << 11
	flagCloseTogether = 1 << 12
)

// Spec 是生成入口携带的不可变产品元数据。
type Spec struct {
	Target, Title string
	Width, Height int
	Kind          string
	Variant       int
	Flags         int
	Asset         string
	// Ready 是示例自己的宿主逻辑（取元素、登记监听、定时器……）；nil 表示这个示例只是页面。
	// 在页面和各 kind 的初始化做完、事件循环开始之前，在 UI 线程上调用一次；返回错误则示例退出。
	Ready func(*Context) error
	// Prepare（可选）在读完命令行之后、加载运行库与建窗口之前调用一次：args 是命令行参数（不含程序名，
	// 也不含开头可选的运行库路径），props 是将要用来建窗口的属性，可以改。可以在这里设 KUI 启动时读的
	// 环境变量、解析示例自己的命令行。返回 exit=true 则示例立即以 code 退出（例如打印了 --help / --list）。
	Prepare func(args []string, props *kui.WindowProps) (exit bool, code int)
	// BeforeLoad（可选）在窗口、文档、脚本引擎（已绑定）都有了、页面还没加载时调用一次，可在这里定义
	// 自定义元素、设样式配置、注册组件。此时 Context.Second 为 nil。返回错误则示例退出。
	BeforeLoad func(*Context) error
}

// Context 是交给示例宿主逻辑的句柄，生命周期由运行器管理：只用，不要 Close。
type Context struct {
	Window         *kui.Window
	Document       *kui.Document
	Script         *kui.ScriptEngine // 已绑定到 Document：页面脚本、读取表单控件的当前值
	Second         *kui.Window       // 只有多窗口示例才有，否则为 nil
	SecondDocument *kui.Document
	host           *host
}

// Keep 让一个元素包装对象活到示例退出。Go SDK 的 Element.Close 会移除它登记过的监听，
// 所以登记了监听的元素必须交给运行器保管（在处理函数里临时取用的元素自己 Close 即可）。
func (c *Context) Keep(element *kui.Element) {
	c.host.kept = append(c.host.kept, element)
}

// OnClose 登记一个在示例退出时调用的清理函数（例如关闭 VirtualList / ListView / TreeView）。
// 按登记的相反顺序、在窗口关闭之前执行。
func (c *Context) OnClose(fn func()) {
	c.host.closers = append(c.host.closers, fn)
}

type host struct {
	closers        []func()
	spec           Spec
	window         *kui.Window
	document       *kui.Document
	script         *kui.ScriptEngine
	http           *kui.HTTPClient
	kept           []*kui.Element
	second         *kui.Window
	secondDocument *kui.Document
	executable     string
}

// Main 运行一个生成的 canonical example。
func Main(spec Spec) {
	runtime.LockOSThread()
	if err := run(spec); err != nil {
		log.Fatalf("[%s] %v", spec.Target, err)
	}
}

func run(spec Spec) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	executable, _ = filepath.Abs(executable)
	runtimePath, arguments := runtimeAndArguments(filepath.Dir(executable), os.Args[1:])
	h := &host{spec: spec, executable: filepath.Dir(executable)}
	path, width, height, port := h.arguments(arguments)
	props := kui.WindowProps{
		Title: spec.Title, Width: width, Height: height, Centered: true,
		NoDecorations:   spec.Flags&flagFrameless != 0,
		Frameless:       spec.Flags&flagFrameless != 0,
		Transparent:     spec.Flags&flagTransparent != 0,
		AlwaysOnTop:     spec.Flags&flagAlwaysOnTop != 0,
		FramelessShadow: spec.Flags&flagShadow != 0,
	}
	if spec.Prepare != nil {
		if exit, code := spec.Prepare(arguments, &props); exit {
			os.Exit(code) // 运行库还没加载，没有要释放的东西
		}
	}
	if err := kui.Load(runtimePath); err != nil {
		return fmt.Errorf("加载 %s: %w", runtimePath, err)
	}
	if err := kui.Init(); err != nil {
		return err
	}
	defer kui.Shutdown()

	defer h.close()
	h.window, err = kui.NewWindow(props)
	if err != nil {
		return err
	}
	h.document, err = h.window.Document()
	if err != nil {
		return err
	}
	if spec.Flags&flagSizeable != 0 {
		_ = h.window.SetFramelessSizeable(true)
	}
	if spec.Kind == "html_loader" {
		h.prepareHTMLLoader()
	}
	h.script, err = kui.NewScriptEngine()
	if err != nil {
		return err
	}
	if err := h.script.BindDocument(h.document); err != nil {
		return err
	}
	if err := h.window.SetScriptEngine(h.script); err != nil {
		return err
	}
	if spec.Flags&flagExtensions != 0 {
		count, initErr := h.document.InitExtensions(h.executable)
		if initErr != nil {
			return initErr
		}
		fmt.Fprintf(os.Stderr, "[%s] extensions=%d\n", spec.Target, count)
	}
	if spec.BeforeLoad != nil {
		if err := spec.BeforeLoad(&Context{
			Window: h.window, Document: h.document, Script: h.script, host: h,
		}); err != nil {
			return fmt.Errorf("示例加载前逻辑: %w", err)
		}
	}
	asset, err := h.findAsset(path)
	if err != nil {
		return err
	}
	if err := h.document.LoadFile(asset); err != nil {
		return err
	}
	if err := h.script.ExecuteScripts(); err != nil {
		return err
	}
	// 环境变量 KUI_EXAMPLE_DEVTOOLS_PORT 是与 C / Rust 运行器共用的测试开关：设了就启动 DevTools
	// （自己启动 DevTools 的示例也改用这个端口），scripts/verify_example_ports.py 靠它真实点击并读回 DOM。
	devtoolsPort, fromEnv := port, false
	if value := os.Getenv("KUI_EXAMPLE_DEVTOOLS_PORT"); value != "" {
		if _, scanErr := fmt.Sscanf(value, "%d", &devtoolsPort); scanErr == nil {
			fromEnv = true
		}
	}
	if spec.Flags&flagDevTools != 0 || fromEnv {
		started, startErr := h.window.StartDevTools(uint16(devtoolsPort))
		if startErr != nil {
			return startErr
		}
		fmt.Fprintf(os.Stderr, "[%s] devtools=%v port=%d\n", spec.Target, started, devtoolsPort)
	}
	if spec.Flags&flagMultiWindow != 0 {
		if err := h.setupSecondWindow(); err != nil {
			return err
		}
		if spec.Flags&flagCloseTogether != 0 {
			h.closeTogether()
		}
	}
	if spec.Flags&flagTray != 0 {
		if err := h.setupTray(); err != nil {
			return err
		}
	}
	if err := h.setupKind(); err != nil {
		return err
	}
	if spec.Flags&flagLottie != 0 {
		h.setupLottie()
	}
	if spec.Flags&flagPDF != 0 {
		written, pdfErr := h.document.PrintToPDF("kui-go-example.pdf", kui.PDFOptions{
			Title: spec.Title, Author: "KUI Go SDK canonical example",
			PageWidthPT: 612, PageHeightPT: 792, GenerateOutline: true,
		})
		if pdfErr != nil {
			return pdfErr
		}
		fmt.Fprintf(os.Stderr, "[%s] PDF written=%v\n", spec.Target, written)
	}
	if spec.Ready != nil {
		if err := spec.Ready(&Context{
			Window: h.window, Document: h.document, Script: h.script,
			Second: h.second, SecondDocument: h.secondDocument, host: h,
		}); err != nil {
			return fmt.Errorf("示例宿主逻辑: %w", err)
		}
	}
	if err := h.window.Show(); err != nil {
		return err
	}
	if h.second != nil {
		_ = h.second.Show()
	}
	_, err = kui.Run()
	return err
}

func runtimeAndArguments(directory string, args []string) (string, []string) {
	if len(args) != 0 && strings.EqualFold(filepath.Ext(args[0]), filepath.Ext(kui.RuntimeLibraryName)) {
		return args[0], args[1:]
	}
	if value := os.Getenv("KUI_C_RUNTIME"); value != "" {
		return value, args
	}
	return filepath.Join(directory, kui.RuntimeLibraryName), args
}

func (h *host) arguments(args []string) (string, int, int, int) {
	// html_loader 的命令行归示例自己的 Prepare 解析，运行器总是先加载打包的页面。
	path, width, height, port := h.spec.Asset, h.spec.Width, h.spec.Height, 9300
	if h.spec.Kind != "file_host" {
		return path, width, height, port
	}
	// 与 C 的 parse_file_host 一致：--file 总是生效；不以 - 开头的参数只在还没有页面时才当页面。
	file := ""
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--file":
			if index+1 < len(args) {
				index++
				file = args[index]
			}
		case "--width":
			if index+1 < len(args) {
				index++
				fmt.Sscanf(args[index], "%d", &width)
			}
		case "--height":
			if index+1 < len(args) {
				index++
				fmt.Sscanf(args[index], "%d", &height)
			}
		case "--port":
			if index+1 < len(args) {
				index++
				fmt.Sscanf(args[index], "%d", &port)
			}
		default:
			if file == "" && !strings.HasPrefix(args[index], "-") {
				file = args[index]
			}
		}
	}
	if file == "" {
		fmt.Fprintf(os.Stderr, "[%s] no --file supplied; opening the packaged Go host page. "+
			"Use --file <html> for ADVEX automation.\n", h.spec.Target)
		return path, width, height, port
	}
	return file, width, height, port
}

func (h *host) findAsset(relative string) (string, error) {
	if filepath.IsAbs(relative) {
		if _, err := os.Stat(relative); err == nil {
			return relative, nil
		}
	}
	candidates := []string{
		relative,
		filepath.Join(h.executable, "assets", relative),
		filepath.Join(h.executable, relative),
		filepath.Join(h.executable, "..", "assets", relative),
		filepath.Join("examples", "canonical", "assets", relative),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", fmt.Errorf("asset not found: %s", relative)
}

func (h *host) setupKind() error {
	switch h.spec.Kind {
	case "three_d":
		return h.setup3D()
	case "media":
		return h.setupMedia()
	case "network":
		return h.setupNetwork()
	case "hello_extension":
		return h.setupHelloExtension()
	}
	return nil
}

func (h *host) setup3D() error {
	methods := []string{"createTriangle", "createBox", "createSphere", "createBox", "createSphere", "createBox", "createBox", "createBox", "createSphere", "createSphere", "createBox", "createSphere", "createBox", "createBox", "createSphere", "createBox", "createBox", "createBox", "createSphere", "createSphere", "createPlane", "createBox", "createBox", "createSphere", "createBox", "createSphere"}
	if h.spec.Variant < 1 || h.spec.Variant > len(methods) {
		return nil
	}
	loaded, err := h.document.ExtensionLoaded("kui-3d-view")
	if err != nil || !loaded {
		return err
	}
	view, err := h.document.GetElementByID("view")
	if err != nil {
		return err
	}
	defer view.Close()
	index := h.spec.Variant - 1
	values, err := h.document.InvokeExtension(view, methods[index], []kui.ExtensionValue{
		{Type: kui.ExtensionFloat, Float: 0.18 + float64(index%3)*0.20},
		{Type: kui.ExtensionFloat, Float: 0.45 + float64(index%2)*0.25},
		{Type: kui.ExtensionFloat, Float: 0.92 - float64(index%4)*0.13},
	}, 1)
	if err != nil || len(values) == 0 {
		return err
	}
	_, err = h.document.InvokeExtension(view, "addToScene", []kui.ExtensionValue{values[0]}, 0)
	return err
}

func (h *host) setupMedia() error {
	media, err := h.document.GetElementByID("media")
	if err != nil {
		return err
	}
	h.kept = append(h.kept, media)
	_ = media.SetMediaControls(true)
	_ = media.SetVolume(0.75)
	name := "native_media/sample.mp4"
	if h.spec.Variant == 1 {
		name = "native_media/sample.mp3"
	}
	if path, findErr := h.findAsset(name); findErr == nil {
		_ = media.SetMediaSource(path)
	}
	if status, statusErr := h.document.GetElementByID("status"); statusErr == nil {
		available, _ := media.MediaAvailable()
		_ = status.SetTextContent(fmt.Sprintf("Media backend available=%v through Go SDK.", available))
		status.Close()
	}
	for id, play := range map[string]bool{"play": true, "pause": false} {
		button, buttonErr := h.document.GetElementByID(id)
		if buttonErr != nil {
			continue
		}
		h.kept = append(h.kept, button)
		_, _ = button.AddEventListener("click", func(*kui.Event) {
			if play {
				_ = media.MediaPlay()
			} else {
				_ = media.MediaPause()
			}
		})
	}
	return nil
}

func (h *host) setupNetwork() error {
	client, err := kui.NewHTTPClient()
	if err != nil {
		return err
	}
	h.http = client
	_ = client.SetUserAgent("kui-go-canonical/3.1.0")
	button, err := h.document.GetElementByID("send")
	if err != nil {
		return err
	}
	h.kept = append(h.kept, button)
	_, err = button.AddEventListener("click", func(*kui.Event) {
		url := "https://example.com/"
		if input, inputErr := h.document.GetElementByID("url"); inputErr == nil {
			if value, valueErr := input.Attribute("value"); valueErr == nil && value != "" {
				url = value
			}
			input.Close()
		}
		if h.spec.Variant == 2 {
			_ = client.FetchAsync(url, 15, func(response kui.HTTPResponse, fetchErr error) {
				_ = kui.PostTask(func() { h.showResponse(response, fetchErr) })
			})
		} else {
			response, fetchErr := client.Fetch(url, 15)
			h.showResponse(response, fetchErr)
		}
	})
	return err
}

func (h *host) showResponse(response kui.HTTPResponse, err error) {
	text := response.Error
	if err != nil {
		text = err.Error()
	} else if len(response.Body) != 0 {
		body := response.Body
		if len(body) > 8192 {
			body = body[:8192]
		}
		text = string(body)
	}
	if output, outputErr := h.document.GetElementByID("response"); outputErr == nil {
		_ = output.SetTextContent(text)
		output.Close()
	}
	if status, statusErr := h.document.GetElementByID("status"); statusErr == nil {
		_ = status.SetTextContent(fmt.Sprintf("HTTP %d via Go SDK", response.Status))
		status.Close()
	}
}

func (h *host) setupHelloExtension() error {
	loaded, err := h.document.ExtensionLoaded("kui-hello")
	if err != nil {
		return err
	}
	if status, statusErr := h.document.GetElementByID("status"); statusErr == nil {
		_ = status.SetTextContent(fmt.Sprintf("kui_ext_hello loaded=%v through Go SDK", loaded))
		status.Close()
	}
	button, err := h.document.GetElementByID("action")
	if err != nil {
		return nil
	}
	h.kept = append(h.kept, button)
	_, err = button.AddEventListener("click", func(*kui.Event) {
		element, elementErr := h.document.GetElementByID("hello")
		if elementErr != nil {
			return
		}
		defer element.Close()
		values, invokeErr := h.document.InvokeExtension(element, "greet", []kui.ExtensionValue{{Type: kui.ExtensionString, String: "Go SDK"}}, 1)
		if invokeErr == nil && len(values) != 0 {
			if status, statusErr := h.document.GetElementByID("status"); statusErr == nil {
				_ = status.SetTextContent(values[0].String)
				status.Close()
			}
		}
	})
	return err
}

func (h *host) setupLottie() {
	path, err := h.findAsset("native_animation/pulse.json")
	if err != nil {
		return
	}
	element, err := h.document.QuerySelector("lottie-player")
	if err != nil {
		return
	}
	defer element.Close()
	_ = element.SetLottieSource(path)
	_ = element.LottiePlay()
}

func (h *host) setupSecondWindow() error {
	window, err := kui.NewWindow(kui.WindowProps{Title: "KUI Go — second window", Width: h.spec.Width, Height: h.spec.Height})
	if err != nil {
		return err
	}
	h.second = window
	document, err := window.Document()
	if err != nil {
		return err
	}
	h.secondDocument = document
	return document.LoadHTML("<html><body style='padding:32px;background:#172554;color:white'><h2>Second window through Go SDK</h2></body></html>")
}

func (h *host) close() {
	for index := len(h.closers) - 1; index >= 0; index-- {
		h.closers[index]()
	}
	if h.http != nil {
		_ = h.http.Close()
	}
	for _, element := range h.kept {
		_ = element.Close()
	}
	if h.script != nil {
		_ = h.script.Close()
	}
	if h.second != nil {
		_ = h.second.Close()
	}
	if h.window != nil {
		_ = h.window.Close()
	}
}
