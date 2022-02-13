package main

import (
	"bytes"
	"image"
	"syscall/js"

	"github.com/natessilva/nes"
)

func main() {
	c := make(chan bool)

	var console *nes.Console
	width := 256
	height := 240
	image := image.NewRGBA(image.Rect(0, 0, width, height))
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")

	copyBuf := js.Global().Get("Uint8Array").New(len(image.Pix))
	imgData := ctx.Call("createImageData", width, height)

	loadROM := func(this js.Value, inputs []js.Value) interface{} {
		fileArr := inputs[0]
		inBuf := make([]byte, fileArr.Get("byteLength").Int())
		js.CopyBytesToGo(inBuf, fileArr)
		r := bytes.NewReader(inBuf)

		// TODO error handling
		console = nes.NewConsole()
		console.LoadROM(r)

		js.Global().Call("renderOuter")
		return nil
	}
	js.Global().Set("loadROM", js.FuncOf(loadROM))

	render := func(this js.Value, inputs []js.Value) interface{} {
		if console == nil {
			return nil
		}
		console.RenderFrame(image)
		js.CopyBytesToJS(copyBuf, image.Pix)
		imgData.Get("data").Call("set", copyBuf)
		ctx.Call("putImageData", imgData, 0, 0)
		return nil
	}

	js.Global().Set("render", js.FuncOf(render))
	<-c
}
