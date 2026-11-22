// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/repro_kui011.go; edit that file, not this copy.
// Host logic of examples/repro_kui011/main.cpp (KUI-011 crash repro): a 16 ms interval rebuilds
// the main window's #list with setInnerHTML (24 rows every 8th tick, otherwise 6) while a
// frameless 380x200 "modal" window is created, pumped by a nested message loop, destroyed and
// re-created, round after round. Prints "NO CRASH" and exits 0 after KUI011_ROUNDS rounds
// (default 400); KUI011_TIMEOUT_SEC (default 120) exits 3 instead.
//
// Like the original, everything runs before the regular event loop, pumping messages itself:
// PollEvents / WaitEvents stand for pumpFor's PeekMessage loop, and each nested-loop step is one
// WaitEvents call (GetMessageW returns one message, so KUI011_NESTED_MSGS counts batches here,
// not single messages). Not ported: the Win32 crash handler / minidump (a crash simply kills the
// process), KUI011_SIMPLE (would mean reloading the generated main page) and InvalidateRect
// (setInnerHTML already schedules the repaint). The main window is disabled during the modal with
// Window.SetEnabled (EnableWindow, Windows only) and the dialog is created hidden (WindowProps.Hidden),
// as the original does. The runner's placeholder second window is hidden: the original creates its
// dialog afresh every round.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

const dialogHTML = `<html><head><style>
  body { margin:0; background:#11111b; color:#cdd6f4;
         font-family:'Segoe UI',sans-serif; }
  .box { padding:18px; transform:translateZ(0); opacity:0.99;
         box-shadow:0 12px 40px rgba(0,0,0,0.6); }
  .btn { margin-top:14px; padding:8px 18px; background:#89b4fa; color:#11111b;
         border-radius:6px; display:inline-block; transform:scale(1); }
</style></head><body>
  <div class="box">
    <div>确认要执行该操作吗？（自绘模态消息框）</div>
    <div class="btn" id="ok">确定</div>
  </div>
</body></html>
`

func envInt(name string, fallback int) int {
	if value, err := strconv.Atoi(os.Getenv(name)); err == nil && value > 0 {
		return value
	}
	return fallback
}

func pumpFor(d time.Duration) {
	start := time.Now()
	for time.Since(start) < d {
		_ = kui.WaitEvents(0.001)
	}
}

func ready(ctx *runner.Context) error {
	doc := ctx.Document
	rounds := envInt("KUI011_ROUNDS", 400)
	nestedMsgs := envInt("KUI011_NESTED_MSGS", 80)
	watchdogSec := envInt("KUI011_TIMEOUT_SEC", 120)
	started := time.Now()
	if ctx.Second != nil {
		_ = ctx.Second.Hide()
	}

	tick := 0
	timer, err := kui.SetInterval(func() {
		tick++
		t := tick
		list, err := doc.GetElementByID("list")
		if err != nil || list == nil {
			return
		}
		defer list.Close()
		n := 6
		if t%8 == 0 {
			n = 24
		}
		var html strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&html, `<div class="row layerA"><span class="pill">#%d.%d</span> 扫描扇区 %d</div>`,
				t, i, (t*31+i)%100000)
		}
		_ = list.SetInnerHTML(html.String())
	}, 16)
	if err != nil {
		return err
	}

	pumpFor(300 * time.Millisecond) // let the main window settle
	fmt.Fprintf(os.Stderr, "[repro] start: rounds=%d nestedMsgs=%d timeout=%ds\n", rounds, nestedMsgs, watchdogSec)

	round := 0
	for ; round < rounds; round++ {
		if time.Since(started) > time.Duration(watchdogSec)*time.Second {
			fmt.Fprintf(os.Stderr, "[repro] WATCHDOG TIMEOUT (%ds) — exiting (no crash)\n", watchdogSec)
			return kui.Quit(3)
		}
		dialog, err := kui.NewWindow(kui.WindowProps{Title: "modal", Width: 380, Height: 200,
			Centered: true, Frameless: true, NoDecorations: true, Hidden: true}) // visible = false, as the original
		if err != nil {
			fmt.Fprintf(os.Stderr, "[repro] dlg create failed @round %d\n", round)
			break
		}
		if dialogDoc, err := dialog.Document(); err == nil {
			_ = dialogDoc.LoadHTML(dialogHTML)
		}
		_ = ctx.Window.SetEnabled(false) // standard modal: disable the parent (Windows only)
		_ = dialog.Show()

		// Nested loop: dispatches every window's messages (main timer, paints) while the dialog
		// is up; after nestedMsgs steps the "user" closes it.
		for step := 0; step < nestedMsgs; step++ {
			_ = kui.WaitEvents(0.1)
		}

		_ = ctx.Window.SetEnabled(true)
		_ = dialog.Close()             // dlg.reset(): destroys the window and its document
		pumpFor(20 * time.Millisecond) // back to the main loop for a few frames

		if round%25 == 0 {
			fmt.Fprintf(os.Stderr, "[repro] round %d / %d (tick=%d)\n", round, rounds, tick)
		}
	}

	fmt.Fprintf(os.Stderr, "[repro] NO CRASH after %d rounds (tick=%d)\n", round, tick)
	_ = kui.ClearTimer(timer)
	return kui.Quit(0)
}
