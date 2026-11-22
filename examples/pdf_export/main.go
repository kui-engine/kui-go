package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	if err := doc.LoadHTML(`<h1>KUI Go PDF</h1>`); err != nil {
		log.Fatal(err)
	}
	output := exampleutil.Argument(2, filepath.Join(os.TempDir(), "kui-go-sdk-example.pdf"))
	written, err := doc.PrintToPDF(output, kui.PDFOptions{
		Title:           "KUI Go PDF",
		Author:          "KUI",
		PageWidthPT:     612,
		PageHeightPT:    792,
		GenerateOutline: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PDF written:", written)
}
