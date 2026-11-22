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
	name, err := kui.OSName()
	if err != nil {
		log.Fatal(err)
	}
	version, err := kui.OSVersion()
	if err != nil {
		log.Fatal(err)
	}
	monitor, err := kui.PrimaryMonitor()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s, primary=%+v\n", name, version, monitor)
}
