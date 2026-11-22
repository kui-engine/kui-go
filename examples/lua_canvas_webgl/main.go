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
	if err := doc.LoadHTML(`<canvas id="canvas" width="64" height="64"></canvas>`); err != nil {
		log.Fatal(err)
	}
	engine, err := kui.NewScriptEngine()
	if err != nil {
		log.Fatal(err)
	}
	defer engine.Close()
	if err := engine.BindDocument(doc); err != nil {
		log.Fatal(err)
	}
	result, err := engine.Eval(`local canvas=document:getElementById("canvas"); local ctx=canvas:getContext("2d"); ctx:fillRect(0,0,8,8)`, "go-canvas.lua")
	if err != nil || !result.Success {
		log.Fatalf("Lua/Canvas 执行失败：%v %+v", err, result)
	}
	fmt.Println("Lua/Canvas PASS；WebGL 使用同一 ScriptEngine 边界")
}
