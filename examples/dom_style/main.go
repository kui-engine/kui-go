package main

import (
	"fmt"
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
	doc, err := kui.NewDocument()
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()
	if err := doc.LoadHTML(`<div id="probe">DOM</div>`); err != nil {
		log.Fatal(err)
	}
	el, err := doc.GetElementByID("probe")
	if err != nil {
		log.Fatal(err)
	}
	defer el.Close()
	if err := el.SetStyleProperty("color", "red", false); err != nil {
		log.Fatal(err)
	}
	value, err := el.StyleProperty("color")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DOM/style PASS:", value)
}
