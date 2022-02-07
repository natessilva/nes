package main

import (
	"bytes"
	"image"
	"image/color"
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
	ctx := canvas.Call("getContext", "2d")

	width := 256
	height := 240
	image := image.NewRGBA(image.Rect(0, 0, width, height))
	copyBuf := js.Global().Get("Uint8Array").New(len(image.Pix))
	imgData := ctx.Call("createImageData", width, height)

	posX := 0
	posY := 0
	frame := 0
	colors := []color.Color{color.Black, color.White}
	render := func(this js.Value, inputs []js.Value) interface{} {
		colorIndex := (frame / 240) % 2
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				image.Set(x, y, colors[colorIndex])
			}
		}
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				image.Set(posX+i, posY+j, colors[(colorIndex+1)%2])

			}
		}
		frame++
		js.CopyBytesToJS(copyBuf, image.Pix)
		imgData.Get("data").Call("set", copyBuf)
		ctx.Call("putImageData", imgData, 0, 0)
		posX = (posX + 1) % 256
		posY = (posY + 1) % 240
		return nil
	}

	js.Global().Set("render", js.FuncOf(render))
	js.Global().Call("renderOuter")
	<-c
}
