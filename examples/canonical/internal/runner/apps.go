package runner

import (
	"fmt"
	"time"

	kui "github.com/kui-engine/kui-go/v3"
)

// 个别 canonical 示例自带的应用行为，照 C++ 原版移植（与 C runtime 的同名部分一一对应）。

// ---------------------------------------------------------------------------
// kui_tray_icon_example：examples/tray_icon_example 的托盘行为
// ---------------------------------------------------------------------------

const (
	trayMenuToggle uint32 = iota + 1
	trayMenuInfo
	trayMenuWarn
	trayMenuError
	trayMenuAbout
	trayMenuQuit
)

var trayTooltips = [3]string{
	"KUI Tray Icon Example — Idle",
	"KUI Tray Icon Example — Syncing…",
	"KUI Tray Icon Example — All done ✓",
}
var trayLabels = [3]string{"Idle", "Busy", "OK"}

type trayApp struct {
	h                                               *host
	clicks, doubles, rights, balloonClicks, variant int
	trayVisible, closeHinted                        bool
}

func (t *trayApp) byID(id string) *kui.Element {
	element, err := t.h.document.GetElementByID(id)
	if err != nil {
		return nil
	}
	return element
}

func (t *trayApp) setText(id, text string) {
	if element := t.byID(id); element != nil {
		_ = element.SetTextContent(text)
		_ = element.Close()
	}
}

// log 把最新一条放在最前，最多 10 条（与原版相同）。
func (t *trayApp) log(message string) {
	list := t.byID("logList")
	if list == nil {
		return
	}
	defer list.Close()
	item, err := t.h.document.CreateElement("li")
	if err != nil {
		return
	}
	_ = item.SetTextContent(time.Now().Format("15:04:05.000") + "  " + message)
	first, _ := list.FirstChild()
	_ = list.InsertBefore(item, first)
	if first != nil {
		_ = first.Close()
	}
	_ = item.Close()
	for {
		count, err := list.ChildCount()
		if err != nil || count <= 10 {
			break
		}
		last, err := list.LastChild()
		if err != nil || last == nil {
			break
		}
		_ = list.RemoveChild(last)
		_ = last.Close()
	}
}

func (t *trayApp) visible() bool {
	visible, _ := t.h.window.IsVisible()
	return visible
}

func (t *trayApp) toast(level kui.ToastLevel, message string) {
	if !t.visible() {
		return
	}
	_, _ = t.h.document.ShowToast(kui.ToastOptions{Message: message, Level: level, Position: "top-right", DurationMS: 2500})
}

func (t *trayApp) showWindow() {
	_ = t.h.window.Show()
	_ = t.h.window.Focus()
}

func (t *trayApp) balloon(icon int, title, body string) {
	t.log("[menu] Balloon: " + title)
	_ = t.h.window.ShowTrayBalloon(title, body, icon, 4000)
}

func (t *trayApp) onTray(event kui.TrayEvent) {
	switch event {
	case kui.TrayClick:
		t.clicks++
		t.setText("cntClick", fmt.Sprint(t.clicks))
		t.log("[tray] left-click")
		t.toast(kui.ToastInfo, "Tray: left-click")
		if t.visible() {
			_ = t.h.window.Hide()
		} else {
			t.showWindow()
		}
	case kui.TrayDoubleClick:
		t.doubles++
		t.setText("cntDouble", fmt.Sprint(t.doubles))
		t.log("[tray] double-click (show + focus)")
		t.showWindow()
		t.toast(kui.ToastSuccess, "Window restored")
	case kui.TrayRightClick:
		t.rights++
		t.setText("cntRight", fmt.Sprint(t.rights))
		t.log("[tray] right-click (context menu)")
	case kui.TrayBalloonClick:
		t.balloonClicks++
		t.setText("cntBalloon", fmt.Sprint(t.balloonClicks))
		t.log("[tray] balloon-click (restore window)")
		t.showWindow()
		t.toast(kui.ToastInfo, "Balloon clicked")
	}
}

func (t *trayApp) onMenu(id uint32) {
	switch id {
	case trayMenuToggle:
		if t.visible() {
			t.log("[menu] Hide window")
			_ = t.h.window.Hide()
		} else {
			t.log("[menu] Show window")
			t.showWindow()
		}
	case trayMenuInfo:
		t.balloon(1, "KUI", "Hello from system tray!")
	case trayMenuWarn:
		t.balloon(2, "Heads up", "Something may need attention.")
	case trayMenuError:
		t.balloon(3, "Error", "An example error has occurred.")
	case trayMenuAbout:
		t.log("[menu] About")
		_ = t.h.window.ShowTrayBalloon("About KUI Tray Example", "KUI Framework v1.x · ENH-P1-003 System Tray Icon · © 2026", 1, 5000)
	case trayMenuQuit:
		t.log("[menu] Quit requested")
		t.closeHinted = true // 之后的关闭真正退出
		_ = t.h.window.DestroyTray()
		_ = t.h.window.CloseRequest()
		_ = kui.Quit(0)
	}
}

func (t *trayApp) updateBadge() {
	if badge := t.byID("statusBadge"); badge != nil {
		class := "status-dot"
		if !t.trayVisible {
			class = "status-dot off"
		}
		_ = badge.SetClassName(class)
		_ = badge.Close()
	}
	if t.trayVisible {
		t.setText("statusText", "active")
	} else {
		t.setText("statusText", "hidden")
	}
}

func (t *trayApp) onButton(id string) {
	switch id {
	case "btnBalloonInfo":
		t.balloon(1, "Info from app", "This balloon was sent by the in-app button.")
	case "btnBalloonWarn":
		t.balloon(2, "Warning from app", "Remember to save your work.")
	case "btnBalloonError":
		t.balloon(3, "Error from app", "Simulated error for demonstration.")
	case "btnToggleTooltip":
		t.variant = (t.variant + 1) % 3
		_ = t.h.window.SetTrayTooltip(trayTooltips[t.variant])
		t.setText("tooltipLabel", trayLabels[t.variant])
		t.log("[ui] tooltip changed")
	case "btnToggleVisible":
		t.trayVisible = !t.trayVisible
		_ = t.h.window.SetTrayVisible(t.trayVisible)
		t.updateBadge()
		if t.trayVisible {
			t.setText("btnToggleVisible", "Hide Tray")
			t.log("[ui] tray shown")
		} else {
			t.setText("btnToggleVisible", "Show Tray")
			t.log("[ui] tray hidden")
		}
	}
}

// onClose：第一次关闭缩到托盘并弹气泡，第二次真正关闭。
func (t *trayApp) onClose(event *kui.Event) {
	if !t.closeHinted {
		t.closeHinted = true
		_ = event.PreventDefault()
		_ = t.h.window.Hide()
		_ = t.h.window.ShowTrayBalloon("Still running", "KUI Tray Example was minimized to the system tray. Double-click the icon to restore, or right-click for Quit.", 1, 5000)
		t.log("[close] minimized to tray (first close)")
		return
	}
	_ = t.h.window.DestroyTray()
}

func (h *host) setupTray() error {
	t := &trayApp{h: h, trayVisible: true}
	if err := h.window.CreateTray("", trayTooltips[0], true); err != nil {
		return fmt.Errorf("failed to register tray icon - system tray may be disabled: %w", err)
	}
	if err := h.window.SetTrayEventHandler(t.onTray); err != nil {
		return err
	}
	items := []kui.MenuItem{
		{Text: "Show / Hide Window", ID: trayMenuToggle}, {Separator: true},
		{Text: "Send Info Balloon", ID: trayMenuInfo},
		{Text: "Send Warning Balloon", ID: trayMenuWarn},
		{Text: "Send Error Balloon", ID: trayMenuError}, {Separator: true},
		{Text: "About…", ID: trayMenuAbout}, {Separator: true},
		{Text: "Quit", ID: trayMenuQuit},
	}
	if err := h.window.SetTrayContextMenu(items, t.onMenu); err != nil {
		return err
	}
	t.log("[init] tray icon installed")
	for _, id := range []string{"btnBalloonInfo", "btnBalloonWarn", "btnBalloonError", "btnToggleTooltip", "btnToggleVisible"} {
		if button := t.byID(id); button != nil {
			name := id
			_, _ = button.AddEventListener("click", func(*kui.Event) { t.onButton(name) })
			h.kept = append(h.kept, button)
		}
	}
	if _, err := h.window.AddEventListener("close", false, t.onClose); err != nil {
		return err
	}
	t.log("[init] ready")
	return nil
}

// ---------------------------------------------------------------------------
// kui_multi_window_animation：关掉任何一个窗口就关掉全部
// ---------------------------------------------------------------------------

func (h *host) closeTogether() {
	closing := false
	link := func(self *kui.Window) {
		_, _ = self.AddEventListener("close", false, func(*kui.Event) {
			if closing {
				return
			}
			closing = true
			for _, window := range []*kui.Window{h.window, h.second} {
				if window != nil && window != self {
					_ = window.CloseRequest()
				}
			}
		})
	}
	link(h.window)
	if h.second != nil {
		link(h.second)
	}
}

// ---------------------------------------------------------------------------
// kui_html_loader：浏览器默认样式、内置组件、原生弹出层（样例页的三个列表在示例自己的 port 里）
// ---------------------------------------------------------------------------

func (h *host) prepareHTMLLoader() {
	_ = h.document.SetStyleProfile(kui.StyleBrowserCompat)
	_ = h.document.RegisterBuiltinComponents()
	_ = h.document.Popups().SetMode(kui.PopupModeNativeWindow)
}
