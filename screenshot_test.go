//go:build (windows && amd64) || (linux && amd64) || (darwin && arm64)

package kui

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// Window.SaveScreenshot 写出的 PNG 是页面画面（T-BUG-WINDOW-SAVE-SCREENSHOT-BLANK：以前对已上屏的
// GL 表面做快照，红色方块在图里是白的）。
func TestWindowSaveScreenshotContent(t *testing.T) {
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
		window, err := NewWindow(WindowProps{Title: "kui-go-screenshot", Width: 640, Height: 480})
		if err != nil {
			t.Fatal(err)
		}
		defer window.Close()
		document, err := window.Document()
		if err != nil {
			t.Fatal(err)
		}
		defer document.Close()
		if err := document.LoadHTML(`<html><body style="margin:0;background:#fff"><div style="position:absolute;left:100px;top:80px;width:120px;height:60px;background:#f00"></div></body></html>`); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			_ = WaitEvents(0.02)
		}
		scale, err := window.DevicePixelRatio()
		if err != nil || scale <= 0 {
			scale = 1
		}
		shot := filepath.Join(t.TempDir(), "shot.png")
		if saved, err := window.SaveScreenshot(shot); err != nil || !saved {
			t.Fatalf("SaveScreenshot: %v %v", saved, err)
		}
		file, err := os.Open(shot)
		if err != nil {
			t.Fatal(err)
		}
		picture, err := png.Decode(file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		at := func(x, y float32) (uint32, uint32, uint32) {
			r, g, b, _ := picture.At(int(x*scale), int(y*scale)).RGBA()
			return r >> 8, g >> 8, b >> 8
		}
		if r, g, b := at(160, 110); r < 200 || g > 60 || b > 60 {
			t.Fatalf("red block expected at (160,110), got %d,%d,%d (image %v)", r, g, b, picture.Bounds())
		}
		if r, g, b := at(400, 300); r < 240 || g < 240 || b < 240 {
			t.Fatalf("white page expected at (400,300), got %d,%d,%d", r, g, b)
		}
	})
}
