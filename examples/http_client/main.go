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
	client, err := kui.NewHTTPClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	response, err := client.Fetch(exampleutil.Argument(2, "https://example.com"), 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("HTTP %d, %d bytes, %s\n", response.Status, len(response.Body), response.ContentType)
}
