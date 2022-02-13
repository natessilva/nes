package main

import (
	"bytes"
	"image"
	"log"
	"syscall/js"

	"github.com/natessilva/nes"
)

func loadROM(this js.Value, inputs []js.Value) interface{} {
	fileArr := inputs[0]
	inBuf := make([]byte, fileArr.Get("byteLength").Int())
	js.CopyBytesToGo(inBuf, fileArr)
	r := bytes.NewReader(inBuf)
	cart, err := nes.ReadFile(r)
	if err != nil {
		log.Fatal("err", err)
	}
	width := 256
	height := 240
	image := image.NewRGBA(image.Rect(0, 0, width, height))
	ppu := nes.NewPPU(cart, image)
	cpu := nes.NewCPU(cart, ppu)
	nes := nes.NewNES(cpu, ppu)

	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")

	copyBuf := js.Global().Get("Uint8Array").New(len(image.Pix))
	imgData := ctx.Call("createImageData", width, height)

	render := func(this js.Value, inputs []js.Value) interface{} {
		nes.StepFrame()
		js.CopyBytesToJS(copyBuf, image.Pix)
		imgData.Get("data").Call("set", copyBuf)
		ctx.Call("putImageData", imgData, 0, 0)
		return nil
	}

	js.Global().Set("render", js.FuncOf(render))
	js.Global().Call("renderOuter")

	return nil
}

func main() {
	c := make(chan bool)
	js.Global().Set("loadROM", js.FuncOf(loadROM))
	<-c
}
