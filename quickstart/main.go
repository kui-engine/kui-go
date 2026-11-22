package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	kui "github.com/kui-engine/kui-go/v3"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	check(kui.Load("")) // "" = kui_c.dll next to this executable
	check(kui.Init())   // locks this goroutine to the UI thread
	defer kui.Shutdown()

	window, err := kui.NewWindow(kui.WindowProps{
		Title: "KUI Quick Start", Width: 560, Height: 640, NoResize: true, Centered: true,
	})
	check(err)
	defer window.Close() // destroys the window and releases its document

	doc, err := window.Document() // owned by the window, do not Close it yourself
	check(err)
	exe, err := os.Executable()
	check(err)
	check(doc.LoadFile(filepath.Join(filepath.Dir(exe), "app.html")))

	count, err := doc.GetElementByID("count")
	check(err)
	defer count.Close()
	button, err := doc.GetElementByID("inc")
	check(err)
	defer button.Close()               // also removes the click listener below
	if count == nil || button == nil { // GetElementByID returns nil, nil for an unknown id
		log.Fatal("app.html has no #count or #inc")
	}

	clicks := 0
	_, err = button.AddEventListener("click", func(*kui.Event) {
		clicks++
		_ = count.SetTextContent(strconv.Itoa(clicks)) // the DOM update
	})
	check(err)

	check(window.Show())
	_, err = kui.Run() // returns when the window is closed
	check(err)
}
