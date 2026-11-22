// Copied by scripts/generate_go_rust_examples.py from examples/go/ports/native_multiwindow_03_window_messaging.go; edit that file, not this copy.
// Host logic of examples/native/multiwindow/03_window_messaging/main.cpp: Window A's Send button
// counts messages, logs "Sent message #N" in #log and "Received message #N from Window A" in
// Window B's #log2.
//
// The generated page only carries Window A; the runner's second window shows a placeholder, so
// this port loads the original Window B page (and title) into it first.
package main

import (
	"fmt"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/canonical/internal/runner"
)

const windowBHTML = `<!DOCTYPE html><html><head><style>
*{margin:0;padding:0;box-sizing:border-box}
body{padding:24px;font-family:"Segoe UI",sans-serif;background:#064e3b;color:#f8fafc}
h2{margin-bottom:16px}
.card{background:rgba(255,255,255,.1);border-radius:12px;padding:20px;border:1px solid rgba(255,255,255,.1)}
.card h3{font-size:14px;margin-bottom:8px}
.log{background:rgba(0,0,0,.3);border-radius:8px;padding:12px;font-size:12px;color:#a7f3d0;font-family:"Consolas",monospace;min-height:60px}
</style></head><body>
<h2>Window B — Receiver</h2>
<div class="card"><h3>Messages</h3><div class="log" id="log2">Waiting for messages...</div></div>
</body></html>`

func ready(ctx *runner.Context) error {
	d1, d2 := ctx.Document, ctx.SecondDocument
	if d2 == nil {
		return fmt.Errorf("没有第二个窗口")
	}
	if err := d2.LoadHTML(windowBHTML); err != nil {
		return err
	}
	_ = ctx.Second.SetTitle("KUI — Window B")

	send, err := d1.GetElementByID("sendBtn")
	if err != nil || send == nil {
		return fmt.Errorf("没有 #sendBtn: %v", err)
	}
	ctx.Keep(send)

	setText := func(doc *kui.Document, id, text string) {
		if el, err := doc.GetElementByID(id); err == nil && el != nil {
			_ = el.SetTextContent(text)
			el.Close()
		}
	}
	count := 0
	_, err = send.AddEventListener("click", func(*kui.Event) {
		count++
		setText(d1, "log", fmt.Sprintf("Sent message #%d", count))
		setText(d2, "log2", fmt.Sprintf("Received message #%d from Window A", count))
	})
	return err
}
