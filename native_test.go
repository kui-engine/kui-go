//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func currentThreadID() uintptr { return uintptr(osThreadID()) }

func TestHTTPBodyUsesABIBufferCapacity(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Content-Type", "application/octet-stream")
			_, _ = writer.Write([]byte("body-without-spare-capacity"))
		}))
		defer server.Close()

		client, err := NewHTTPClient()
		if err != nil {
			t.Fatal(err)
		}
		defer client.Close()
		response, err := client.Fetch(server.URL, 5)
		if err != nil {
			t.Fatal(err)
		}
		if response.Error != "" || response.Status != 200 {
			t.Fatalf("HTTP fetch failed: %+v", response)
		}
		if got, want := string(response.Body), "body-without-spare-capacity"; got != want {
			t.Fatalf("HTTP body = %q, want %q", got, want)
		}
	})
}

func loadTestRuntime(t testing.TB) {
	t.Helper()
	runtimePath := os.Getenv("KUI_C_TEST_RUNTIME")
	if runtimePath == "" {
		packaged := filepath.Join("native", "bin", RuntimeLibraryName)
		if _, err := os.Stat(packaged); err == nil {
			runtimePath = packaged
		} else if runtime.GOOS == "windows" {
			runtimePath = filepath.Join("..", "c", "build", "Release", RuntimeLibraryName)
		} else {
			// 单配置生成器（Ninja / Makefiles）的 sdk/c 构建把库放在构建目录根下。
			runtimePath = filepath.Join("..", "c", "build", RuntimeLibraryName)
		}
	}
	runtimePath, err := filepath.Abs(runtimePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := Load(runtimePath); err != nil {
		t.Fatalf("加载 C ABI 测试运行时失败：%v", err)
	}
}

func TestRuntimeExportsAndVersion(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		if !RuntimeLoaded() {
			t.Fatal("全部导出解析完成后 RuntimeLoaded 仍为 false")
		}
		version, err := RuntimeVersion()
		if err != nil {
			t.Fatal(err)
		}
		if version.Major != 3 || version.Minor < 1 {
			t.Fatalf("运行时版本异常：%+v", version)
		}
	})
}

func TestDocumentDOMAndLua(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		if err := Init(); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := Shutdown(); err != nil {
				t.Errorf("Shutdown: %v", err)
			}
		}()

		document, err := NewDocument()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		documentClone, err := document.Clone()
		if err != nil {
			t.Fatal(err)
		}
		defer documentClone.Close()
		if err := document.LoadHTML(`<html><body><div id="probe" class="a">ready</div></body></html>`); err != nil {
			t.Fatal(err)
		}
		probe, err := document.GetElementByID("probe")
		if err != nil {
			t.Fatal(err)
		}
		defer probe.Close()
		probeClone, err := probe.Clone()
		if err != nil {
			t.Fatal(err)
		}
		defer probeClone.Close()
		if got, err := probe.TextContent(); err != nil || got != "ready" {
			t.Fatalf("TextContent=(%q, %v)", got, err)
		}
		if err := probe.SetStyleProperty("color", "rgb(1, 2, 3)", false); err != nil {
			t.Fatal(err)
		}
		if got, err := probe.StyleProperty("color"); err != nil || got == "" {
			t.Fatalf("StyleProperty=(%q, %v)", got, err)
		}

		engine, err := NewScriptEngine()
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		if err := engine.BindDocument(document); err != nil {
			t.Fatal(err)
		}
		result, err := engine.Eval("local answer = 6 * 7", "go-sdk-smoke.lua")
		if err != nil {
			t.Fatal(err)
		}
		if !result.Success || result.Error != "" {
			t.Fatalf("Lua 执行失败：%+v", result)
		}
	})
}

func TestHTTPAsyncCloseCompletesCallback(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		client, err := NewHTTPClient()
		if err != nil {
			t.Fatal(err)
		}

		done := make(chan error, 1)
		if err := client.FetchAsync("not-a-url", 1, func(response HTTPResponse, callbackErr error) {
			if callbackErr != nil {
				done <- callbackErr
				return
			}
			if response.Error == "" {
				done <- &Error{Status: ErrInternal, Context: "TestHTTPAsyncCloseCompletesCallback", Message: "invalid URL unexpectedly succeeded"}
				return
			}
			done <- nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := client.Close(); err != nil {
			t.Fatal(err)
		}

		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("HTTPClient.Close 后异步完成回调超时")
		}
	})
}

func TestHTTPCallbackPanicDoesNotCrossCABI(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		client, err := NewHTTPClient()
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan struct{})
		if err := client.FetchAsync("not-a-url", 1, func(HTTPResponse, error) {
			close(done)
			panic("callback panic probe")
		}); err != nil {
			t.Fatal(err)
		}
		if err := client.Close(); err != nil {
			t.Fatal(err)
		}
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("panic 回调未被调用")
		}
	})
}

// RequestIdle：UI 线程上登记，事件循环空闲时在 UI 线程执行且只执行一次；nil 回调与 Init 之前都报错。
func TestRequestIdleRunsOnceOnTheUIThread(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		if err := RequestIdle(nil); err == nil {
			t.Fatal("nil 回调应返回错误")
		}
		if err := Init(); err != nil {
			t.Fatal(err)
		}
		uiThread := currentThreadID()
		var ran, wrongThread int32
		if err := RequestIdle(func() {
			atomic.AddInt32(&ran, 1)
			if currentThreadID() != uiThread {
				atomic.StoreInt32(&wrongThread, 1)
			}
		}); err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(5 * time.Second)
		for atomic.LoadInt32(&ran) == 0 && time.Now().Before(deadline) {
			if err := WaitEvents(0.01); err != nil {
				t.Fatal(err)
			}
		}
		for i := 0; i < 5; i++ { // 再转几轮，确认只执行一次
			if err := WaitEvents(0.01); err != nil {
				t.Fatal(err)
			}
		}
		if got := atomic.LoadInt32(&ran); got != 1 {
			t.Fatalf("RequestIdle 执行了 %d 次，期望 1", got)
		}
		if atomic.LoadInt32(&wrongThread) != 0 {
			t.Fatal("RequestIdle 的回调没有在 UI 线程执行")
		}
		if err := Shutdown(); err != nil {
			t.Fatal(err)
		}
	})
}

// PostTask 以前走 request_idle → SetTimer(nullptr)，从别的 goroutine 调用时定时器挂在调用线程上：
// 回调不触发或跑在错误的线程上。现在必须从任意 goroutine 投递、只在 UI 线程执行。
func TestPostTaskFromOtherGoroutinesRunsOnUIThread(t *testing.T) {
	onMain(t, func(t *mainT) {
		loadTestRuntime(t)
		if err := PostTask(func() {}); err == nil {
			t.Fatal("Init 之前 PostTask 应返回错误")
		}
		if err := Init(); err != nil {
			t.Fatal(err)
		}
		uiThread := currentThreadID()
		const goroutines, perGoroutine = 4, 32
		var ran, wrongThread int32
		var wg sync.WaitGroup
		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < perGoroutine; i++ {
					if err := PostTask(func() {
						atomic.AddInt32(&ran, 1)
						if currentThreadID() != uiThread {
							atomic.StoreInt32(&wrongThread, 1)
						}
					}); err != nil {
						t.Error(err)
					}
				}
			}()
		}
		wg.Wait()
		deadline := time.Now().Add(5 * time.Second)
		for atomic.LoadInt32(&ran) < goroutines*perGoroutine && time.Now().Before(deadline) {
			if err := WaitEvents(0.01); err != nil {
				t.Fatal(err)
			}
		}
		if got := atomic.LoadInt32(&ran); got != goroutines*perGoroutine {
			t.Fatalf("PostTask 执行了 %d 个，期望 %d", got, goroutines*perGoroutine)
		}
		if atomic.LoadInt32(&wrongThread) != 0 {
			t.Fatal("PostTask 的回调没有在 UI 线程执行")
		}

		// 主线程版 HTTP：回调落在 UI 线程。
		client, err := NewHTTPClient()
		if err != nil {
			t.Fatal(err)
		}
		httpDone := make(chan uintptr, 1)
		if err := client.FetchAsyncOnMainThread("not-a-url", 1, func(response HTTPResponse, callbackErr error) {
			httpDone <- currentThreadID()
		}); err != nil {
			t.Fatal(err)
		}
		var httpThread uintptr
		for httpThread == 0 && time.Now().Before(deadline.Add(5*time.Second)) {
			select {
			case httpThread = <-httpDone:
			default:
				if err := WaitEvents(0.01); err != nil {
					t.Fatal(err)
				}
			}
		}
		if httpThread != uiThread {
			t.Fatalf("FetchAsyncOnMainThread 的回调线程 %d，UI 线程 %d", httpThread, uiThread)
		}
		if err := client.Close(); err != nil {
			t.Fatal(err)
		}

		// Shutdown 时还在排队的任务被丢弃（不执行，handle 由丢弃回调释放）。
		var late int32
		for i := 0; i < 8; i++ {
			if err := PostTask(func() { atomic.AddInt32(&late, 1) }); err != nil {
				t.Fatal(err)
			}
		}
		if err := Shutdown(); err != nil {
			t.Fatal(err)
		}
		if atomic.LoadInt32(&late) != 0 {
			t.Fatal("Shutdown 不应执行排队中的任务")
		}
		if err := PostTask(func() {}); err == nil {
			t.Fatal("Shutdown 之后 PostTask 应返回错误")
		}
	})
}
