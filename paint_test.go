//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"testing"
)

// C ABI 第 17 节（3.61.8）：自定义元素自绘、tick 与输入。截图解码后核对画出来的像素。
func TestCustomElementPaintTickInput(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-custom-paint", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()

		var (
			paints, ticks, mouseDowns int
			animating                 = true
			paintW, paintH, dpr       float32
			drawErr                   error
			mouse                     CustomMouseEvent
			wheel                     CustomWheelEvent
			key                       string
			update, committed         string
		)
		err = document.DefineCustomElement("x-go-canvas", CustomElementDefinition{
			Paint: func(element *Element, p *PaintContext) {
				paints++
				paintW, paintH, dpr = p.Width, p.Height, p.DevicePixelRatio
				steps := []error{
					p.FillRect(0, 0, p.Width, p.Height, 0xFF0000FF),
					p.FillRect(80, 30, 30, 20, 0x0000FFFF),
					p.Save(),
					p.ClipRoundedRect(0, 0, 50, 50, UniformCornerRadii(6)),
					p.Translate(2, 2),
					p.SetOpacity(0.5),
					p.FillRoundedRect(0, 0, 20, 20, UniformCornerRadii(4), 0x00FF00FF),
					p.StrokeRoundedRect(0, 0, 20, 20, UniformCornerRadii(4), 1, 0x000000FF),
					p.DrawText("hi", 4, 40, &TextStyle{FontSize: 12, FontWeight: 700}, 0x000000FF),
					p.DrawPixels([]uint32{0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00}, 2, 2, 30, 30, 8, 8),
					p.DrawDecorationLine(0, 45, 40, 1, DecorationWavy, 0x00FF00FF),
					p.Restore(),
				}
				for _, step := range steps {
					if step != nil && drawErr == nil {
						drawErr = step
					}
				}
			},
			Tick: func(element *Element, deltaSeconds float64) {
				ticks++
				if animating {
					_ = element.RequestRepaint()
				}
			},
			NeedsTick: func(*Element) bool { return animating },
			Mouse: func(_ *Element, event CustomMouseEvent) bool {
				if event.Type == CustomMouseDown {
					mouseDowns++
					mouse = event
				}
				return false
			},
			Wheel: func(_ *Element, event CustomWheelEvent) bool { wheel = event; return true },
			Key: func(_ *Element, event CustomKeyEvent) bool {
				if event.Type == CustomKeyDown {
					key = event.Key
				}
				return true
			},
			IME: func(_ *Element, event CustomIMEEvent) bool {
				switch event.Type {
				case IMEUpdate:
					update = event.Text
				case IMECommit:
					committed = event.CommittedText
				}
				return true
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := document.LoadHTML(`<html><body style="margin:0;background:#fff"><x-go-canvas id="cv" tabindex="0" style="display:block;position:absolute;left:10px;top:10px;width:120px;height:60px"></x-go-canvas></body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 30 && (paints == 0 || ticks < 3); i++ {
			_ = WaitEvents(0.02)
		}
		if paints == 0 || paintW != 120 || paintH != 60 || dpr <= 0 {
			t.Fatalf("paint: %d %vx%v dpr %v", paints, paintW, paintH, dpr)
		}
		if drawErr != nil {
			t.Fatalf("drawing inside paint: %v", drawErr)
		}
		if ticks < 3 {
			t.Fatalf("ticks: %d", ticks)
		}
		animating = false
		for i := 0; i < 3; i++ {
			_ = WaitEvents(0.02)
		}
		stopped := ticks
		for i := 0; i < 5; i++ {
			_ = WaitEvents(0.02)
		}
		if ticks != stopped {
			t.Fatalf("NeedsTick false should stop ticks: %d -> %d", stopped, ticks)
		}

		// 像素核对：内容框 (100,5) 是红色底，(95,40) 落在蓝色矩形里（窗口坐标再加 10 的偏移）。
		// 走 DevTools 截图（现画一帧再读回）；Window.SaveScreenshot 读的是已上屏后的窗口表面，见
		// T-BUG-WINDOW-SAVE-SCREENSHOT-BLANK。
		response, err := document.DispatchDevToolsCommand(`{"id":1,"method":"Page.captureScreenshot"}`)
		if err != nil {
			t.Fatal(err)
		}
		var capture struct {
			Result struct {
				Data string `json:"data"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(response), &capture); err != nil || capture.Result.Data == "" {
			t.Fatalf("captureScreenshot: %v %.200s", err, response)
		}
		encoded, err := base64.StdEncoding.DecodeString(capture.Result.Data)
		if err != nil {
			t.Fatal(err)
		}
		picture, err := png.Decode(bytes.NewReader(encoded))
		if err != nil {
			t.Fatal(err)
		}
		at := func(x, y float32) (uint32, uint32, uint32) {
			r, g, b, _ := picture.At(int(x*dpr), int(y*dpr)).RGBA()
			return r >> 8, g >> 8, b >> 8
		}
		if r, g, b := at(10+100, 10+5); r < 200 || g > 60 || b > 60 {
			t.Fatalf("red background expected at (110,15), got %d,%d,%d", r, g, b)
		}
		if r, g, b := at(10+95, 10+40); r > 60 || g > 60 || b < 200 {
			t.Fatalf("blue rectangle expected at (105,50), got %d,%d,%d", r, g, b)
		}
		if r, g, b := at(300, 300); r < 240 || g < 240 || b < 240 {
			t.Fatalf("white page outside the element, got %d,%d,%d", r, g, b)
		}

		element, err := document.GetElementByID("cv")
		if err != nil || element == nil {
			t.Fatalf("cv: %v", err)
		}
		defer element.Close()
		if metrics, err := element.MeasureText("Hello", nil); err != nil || metrics.Width <= 0 || metrics.Ascent <= 0 {
			t.Fatalf("MeasureText: %+v %v", metrics, err)
		}
		before := paints
		if err := element.InvalidateRect(0, 0, 10, 10); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 10 && paints == before; i++ {
			_ = WaitEvents(0.02)
		}
		if paints == before {
			t.Fatal("InvalidateRect should repaint")
		}

		_ = document.DispatchMouse(MousePressed, 40, 30, MouseButtonLeft, 0, 1, 0, 0)
		_ = document.DispatchMouse(MouseReleased, 40, 30, MouseButtonLeft, 0, 1, 0, 0)
		if mouseDowns != 1 || mouse.X != 30 || mouse.Y != 20 || mouse.Button != 0 {
			t.Fatalf("mouse: %d %+v", mouseDowns, mouse)
		}
		_ = document.DispatchMouse(MouseWheel, 40, 30, MouseButtonNone, 0, 0, 0, 40)
		if wheel.DeltaY != 40 {
			t.Fatalf("wheel: %+v", wheel)
		}
		if err := element.Focus(); err != nil {
			t.Fatal(err)
		}
		_ = document.DispatchKey(KeyDown, "a", "KeyA", 0, "a", false)
		if key != "a" {
			t.Fatalf("key: %q", key)
		}
		for _, step := range []struct {
			kind IMEEventType
			text string
		}{{IMEStart, ""}, {IMEUpdate, "zhong"}, {IMECommit, "中"}} {
			if err := document.DispatchIME(step.kind, step.text); err != nil {
				t.Fatal(err)
			}
		}
		if update != "zhong" || committed != "中" {
			t.Fatalf("ime: %q %q", update, committed)
		}
	})
}
