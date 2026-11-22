//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"os"
	"runtime"
	"testing"
)

// 测试体在主线程上执行。
//
// macOS 只允许在主线程初始化平台、创建窗口、泵事件，而 testing 把每个测试放在自己的 goroutine 里，那个
// goroutine 不一定在主线程上——Mac 上 5 次里约 2 次 NewWindow / WaitEvents 报 internal error（AppKit 抛出的
// Objective-C 异常被 C ABI 接住）。包初始化在主 goroutine、主线程上执行，init 把主 goroutine 锁在主线程上；
// TestMain 在别的 goroutine 里跑 m.Run，主线程只执行 onMain 投递来的测试体。三个平台同一套。

func init() {
	runtime.LockOSThread()
}

var mainCalls = make(chan func())

func TestMain(m *testing.M) {
	code := make(chan int, 1)
	go func() {
		code <- m.Run()
		close(mainCalls)
	}()
	for call := range mainCalls {
		call()
	}
	os.Exit(<-code)
}

// abortTest 结束主线程上的测试体（mainT 的 Fatal / FailNow）。
type abortTest struct{}

// mainT 是主线程上的测试体拿到的 t。testing 规定 FailNow / Fatal 只能在测试自己的 goroutine 里调用，这里改为
// 「记下失败、panic 结束测试体」，onMain 回到测试 goroutine 后再 FailNow。Errorf、Log、Helper、TempDir 等用内嵌的
// *testing.T，它们可以在别的 goroutine 里调用。Skip 系列同样只能在测试 goroutine 里调用，测试体里不要用。
type mainT struct{ *testing.T }

func (t *mainT) FailNow() {
	t.T.Fail()
	panic(abortTest{})
}

func (t *mainT) Fatal(args ...any) {
	t.T.Helper()
	t.T.Error(args...)
	panic(abortTest{})
}

func (t *mainT) Fatalf(format string, args ...any) {
	t.T.Helper()
	t.T.Errorf(format, args...)
	panic(abortTest{})
}

// onMain 在主线程上执行 body，等它结束。
func onMain(t *testing.T, body func(t *mainT)) {
	t.Helper()
	outcome := make(chan any, 1)
	mainCalls <- func() {
		defer func() { outcome <- recover() }()
		body(&mainT{T: t})
	}
	switch p := (<-outcome).(type) {
	case nil:
	case abortTest:
		t.FailNow()
	default:
		panic(p)
	}
}
