//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"errors"
	"strings"
	"testing"
)

// C ABI 第 11 节（3.61.8）：路由、数据绑定、动画状态、Observer。
func TestFrameworkRouterBindingObservers(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-framework", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		pump := func(rounds int, done func() bool) {
			for i := 0; i < rounds && (done == nil || !done()); i++ {
				_ = WaitEvents(0.01)
			}
		}
		get := func(id string) *Element {
			element, err := document.GetElementByID(id)
			if err != nil || element == nil {
				t.Fatalf("GetElementByID(%s): %v", id, err)
			}
			return element
		}
		text := func(element *Element) string {
			value, err := element.TextContent()
			if err != nil {
				t.Fatal(err)
			}
			return value
		}

		// --- 路由 ---
		if err := document.LoadHTML(`<html><body><main id="view"></main><p id="label">old</p><input id="name"></body></html>`); err != nil {
			t.Fatal(err)
		}
		pump(5, nil)
		router := document.Router()
		view := get("view")
		defer view.Close()
		if err := router.RegisterView(view); err != nil {
			t.Fatal(err)
		}
		build := func(doc *Document, route RouteInfo) *Element {
			element, err := doc.CreateElement("p")
			if err != nil {
				return nil
			}
			_ = element.SetAttribute("id", "routed")
			_ = element.SetTextContent("view:" + route.Pattern + " id=" + route.Params["id"])
			return element
		}
		if err := router.AddRoute("/users/:id", "user", build); err != nil {
			t.Fatal(err)
		}
		if err := router.AddRoute("/admin", "", build); err != nil {
			t.Fatal(err)
		}
		if err := router.AddRoute("/admin", "", build); err == nil {
			t.Fatal("重复的模式应被拒绝")
		}
		if count, _ := router.RouteCount(); count != 2 {
			t.Fatalf("RouteCount: %d", count)
		}
		guardID, err := router.AddGuard(RouteBeforeEnter, func(from, to RouteInfo) bool {
			return !strings.HasPrefix(to.Path, "/admin")
		})
		if err != nil {
			t.Fatal(err)
		}
		var navigatedTab string
		navID, err := router.OnNavigated(func(route RouteInfo) { navigatedTab = route.QueryParams["tab"] })
		if err != nil {
			t.Fatal(err)
		}
		if ok, err := router.Push("/users/42?tab=info"); err != nil || !ok {
			t.Fatalf("Push: %v %v", ok, err)
		}
		if navigatedTab != "info" {
			t.Fatalf("OnNavigated 查询参数: %q", navigatedTab)
		}
		if value, ok, err := router.Param("id"); err != nil || !ok || value != "42" {
			t.Fatalf("Param: %q %v %v", value, ok, err)
		}
		if _, ok, _ := router.Param("missing"); ok {
			t.Fatal("不存在的参数")
		}
		if path, err := router.CurrentPath(); err != nil || path != "/users/42?tab=info" {
			t.Fatalf("CurrentPath: %q %v", path, err)
		}
		current, err := router.Current()
		if err != nil || current.Pattern != "/users/:id" || current.Query != "tab=info" {
			t.Fatalf("Current: %+v %v", current, err)
		}
		routed := get("routed")
		if got := text(routed); got != "view:/users/:id id=42" {
			t.Fatalf("视图内容: %q", got)
		}
		routed.Close()
		if ok, _ := router.Push("/admin"); ok {
			t.Fatal("守卫应取消 /admin")
		}
		_ = router.RemoveGuard(guardID)
		if ok, _ := router.Push("/admin"); !ok {
			t.Fatal("移除守卫后应可导航")
		}
		if ok, _ := router.Back(); !ok {
			t.Fatal("Back")
		}
		if index, size, _ := router.History(); index != 0 || size != 2 {
			t.Fatalf("History: %d %d", index, size)
		}
		_ = router.RemoveNavigated(navID)
		var kuiErr *Error
		if err := router.RemoveRoute("/missing"); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotFound {
			t.Fatalf("RemoveRoute 未知模式: %v", err)
		}
		if err := router.UnregisterView(view); err != nil {
			t.Fatal(err)
		}
		if err := router.Clear(); err != nil {
			t.Fatal(err)
		}

		// --- 数据绑定 ---
		label := get("label")
		defer label.Close()
		obs, err := NewObservable("hello")
		if err != nil {
			t.Fatal(err)
		}
		changes := 0
		subID, err := obs.Subscribe(func() { changes++ })
		if err != nil {
			t.Fatal(err)
		}
		if _, err := obs.Bind(label, "textContent", false); err != nil {
			t.Fatal(err)
		}
		if got := text(label); got != "hello" {
			t.Fatalf("绑定应立即写入: %q", got)
		}
		_ = obs.Set("world")
		_ = obs.Set("world")
		if changes != 1 {
			t.Fatalf("订阅次数: %d", changes)
		}
		_ = document.FlushBindings()
		if got := text(label); got != "world" {
			t.Fatalf("绑定后内容: %q", got)
		}
		if value, _ := obs.Get(); value != "world" {
			t.Fatalf("Get: %q", value)
		}
		// Clone 再取一个引用：关掉原句柄后值仍在，关掉最后一个才释放。
		shared, err := obs.Clone()
		if err != nil {
			t.Fatalf("Observable.Clone: %v", err)
		}
		if err := obs.Close(); err != nil {
			t.Fatal(err)
		}
		if err := shared.Set("shared"); err != nil {
			t.Fatalf("克隆句柄应仍可用: %v", err)
		}
		if value, _ := shared.Get(); value != "shared" {
			t.Fatalf("克隆句柄 Get: %q", value)
		}
		obs = shared
		if _, err := (*Observable)(nil).Clone(); err == nil {
			t.Fatal("已关闭的 Observable 不能 Clone")
		}
		if err := obs.Unsubscribe(subID); err != nil {
			t.Fatal(err)
		}
		if err := obs.Unsubscribe(subID); !errors.As(err, &kuiErr) || kuiErr.Status != ErrNotFound {
			t.Fatalf("重复退订: %v", err)
		}
		if err := obs.Close(); err != nil {
			t.Fatal(err)
		}

		// --- 动画状态与 Observer ---
		if err := document.LoadHTML(`<html><head><style>@keyframes pulse{from{opacity:0}to{opacity:1}}#box{animation:pulse 1s infinite}</style></head>` +
			`<body style="margin:0"><div id="box" title="a" style="width:120px;height:40px"></div></body></html>`); err != nil {
			t.Fatal(err)
		}
		box := get("box")
		defer box.Close()
		animating := false
		pump(100, func() bool { animating, _, _ = box.AnimationState(); return animating })
		if !animating {
			t.Fatal("应报告 CSS 动画")
		}
		if animations, _, _ := document.AnimationCounts(); animations < 1 {
			t.Fatalf("AnimationCounts: %d", animations)
		}

		var oldTitle string
		mutation, err := NewMutationObserver(func(records []MutationRecord) {
			for _, record := range records {
				if record.Type == MutationAttributes && record.AttributeName == "title" {
					oldTitle = record.OldValue
				}
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := mutation.ObserveMutations(box, MutationOptions{Attributes: true, AttributeOldValue: true, AttributeFilter: []string{"title"}}); err != nil {
			t.Fatal(err)
		}
		if err := mutation.Observe(box); err == nil {
			t.Fatal("MutationObserver 应使用 ObserveMutations")
		}
		_ = box.SetAttribute("title", "b")
		if oldTitle != "a" {
			t.Fatalf("旧值: %q", oldTitle)
		}
		// Clone 再取一个引用：关掉原句柄后观察仍有效，关掉最后一个才断开。
		sharedMutation, err := mutation.Clone()
		if err != nil {
			t.Fatalf("Observer.Clone: %v", err)
		}
		_ = mutation.Close()
		_ = box.SetAttribute("title", "c")
		if oldTitle != "b" {
			t.Fatalf("克隆句柄应继续观察: %q", oldTitle)
		}
		if _, err := mutation.Clone(); err == nil {
			t.Fatal("已关闭的 Observer 不能 Clone")
		}
		_ = sharedMutation.Close()

		var width float32
		resize, err := NewResizeObserver(func(entries []ResizeEntry) { width = entries[len(entries)-1].ContentWidth })
		if err != nil {
			t.Fatal(err)
		}
		if err := resize.Observe(box); err != nil {
			t.Fatal(err)
		}
		pump(100, func() bool { return width != 0 })
		if width != 120 {
			t.Fatalf("ResizeObserver 首次报告: %v", width)
		}
		_ = resize.Close()

		intersecting := -1
		intersection, err := NewIntersectionObserver(&IntersectionOptions{Thresholds: []float32{0, 1}}, func(entries []IntersectionEntry) {
			if entries[len(entries)-1].Intersecting {
				intersecting = 1
			} else {
				intersecting = 0
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := intersection.Observe(box); err != nil {
			t.Fatal(err)
		}
		pump(100, func() bool { return intersecting != -1 })
		if intersecting != 1 {
			t.Fatalf("IntersectionObserver: %d", intersecting)
		}
		_ = intersection.Close()
		if _, err := NewIntersectionObserver(&IntersectionOptions{Thresholds: []float32{2}}, func([]IntersectionEntry) {}); err == nil {
			t.Fatal("阈值越界应被拒绝")
		}
	})
}
