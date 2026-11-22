package main

import (
	"log"

	kui "github.com/kui-engine/kui-go/v3"
	"github.com/kui-engine/kui-go/v3/examples/internal/exampleutil"
)

func main() {
	if err := kui.Load(exampleutil.RuntimePath()); err != nil {
		log.Fatal(err)
	}
	if err := kui.Init(); err != nil {
		log.Fatal(err)
	}
	defer kui.Shutdown()

	window, err := kui.NewWindow(kui.WindowProps{Title: "KUI Go SDK", Width: 960, Height: 640, Centered: true})
	if err != nil {
		log.Fatal(err)
	}
	defer window.Close()
	document, err := window.Document()
	if err != nil {
		log.Fatal(err)
	}
	defer document.Close()
	if err := document.LoadHTML(`<main style="padding:48px;font-family:sans-serif"><h1>Go SDK 已接入 KUI</h1><p>此程序仅通过公开 C ABI 工作。</p></main>`); err != nil {
		log.Fatal(err)
	}
	if err := window.Show(); err != nil {
		log.Fatal(err)
	}
	if _, err := kui.Run(); err != nil {
		log.Fatal(err)
	}
}
