package main

import (
	"bytes"
	"log"
	"syscall/js"

	"github.com/natessilva/nes"
)

func loadROM(this js.Value, inputs []js.Value) interface{} {
	fileArr := inputs[0]
	inBuf := make([]byte, fileArr.Get("byteLength").Int())
	js.CopyBytesToGo(inBuf, fileArr)
	r := bytes.NewReader(inBuf)
	_, err := nes.ReadFile(r)
	if err != nil {
		log.Fatal("err", err)
	}
	return nil
}

func main() {
	c := make(chan bool)
	js.Global().Set("loadROM", js.FuncOf(loadROM))
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	dvp := js.Global().Get("devicePixelRatio").Float()
	canvas.Set("width", 256*dvp)
	canvas.Set("height", 240*dvp)
	ctx := canvas.Call("getContext", "2d")
	ctx.Set("fillStyle", "black")
	ctx.Call("fillRect", 10, 10, 10, 10)
	<-c
}
