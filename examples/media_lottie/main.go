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
	if err := doc.LoadHTML(`<audio id="audio"></audio><lottie-player id="lottie"></lottie-player>`); err != nil {
		log.Fatal(err)
	}
	audio, err := doc.GetElementByID("audio")
	if err != nil {
		log.Fatal(err)
	}
	defer audio.Close()
	available, err := audio.MediaAvailable()
	if err != nil {
		log.Fatal(err)
	}
	lottie, err := doc.GetElementByID("lottie")
	if err != nil {
		log.Fatal(err)
	}
	defer lottie.Close()
	if err := lottie.SetLottieSource(exampleutil.ExecutableSibling("assets/native_animation/pulse.json")); err != nil {
		log.Fatal(err)
	}
	if err := lottie.LottiePlay(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("media available=%v, lottie source/play PASS\n", available)
}
