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
	extensionDir := exampleutil.Argument(2, exampleutil.ExecutableSibling("."))
	count, err := doc.InitExtensions(extensionDir)
	if err != nil {
		log.Fatal(err)
	}
	loaded, err := doc.ExtensionLoaded("kui-3d-view")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("extensions=%d, kui-3d-view=%v\n", count, loaded)
}
