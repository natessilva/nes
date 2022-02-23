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
		js.CopyBytesToJS(imgData.Get("data"), image.Pix)
		ctx.Call("putImageData", imgData, 0, 0)
		return nil
	}
	js.Global().Set("render", js.FuncOf(render))

	keydown := func(this js.Value, inputs []js.Value) interface{} {
		if console == nil {
			return nil
		}
		switch inputs[0].String() {
		case "Enter":
			console.SetJoypad(nes.ButtonStart, true)
		case "f":
			console.SetJoypad(nes.ButtonA, true)
		case "d":
			console.SetJoypad(nes.ButtonB, true)
		case " ":
			console.SetJoypad(nes.ButtonSelect, true)
		case "ArrowUp":
			console.SetJoypad(nes.ButtonUp, true)
		case "ArrowDown":
			console.SetJoypad(nes.ButtonDown, true)
		case "ArrowLeft":
			console.SetJoypad(nes.ButtonLeft, true)
		case "ArrowRight":
			console.SetJoypad(nes.ButtonRight, true)
		}
		return nil
	}
	js.Global().Set("keydown", js.FuncOf(keydown))

	keyup := func(this js.Value, inputs []js.Value) interface{} {
		if console == nil {
			return nil
		}
		switch inputs[0].String() {
		case "Enter":
			console.SetJoypad(nes.ButtonStart, false)
		case "f":
			console.SetJoypad(nes.ButtonA, false)
		case "d":
			console.SetJoypad(nes.ButtonB, false)
		case " ":
			console.SetJoypad(nes.ButtonSelect, false)
		case "ArrowUp":
			console.SetJoypad(nes.ButtonUp, false)
		case "ArrowDown":
			console.SetJoypad(nes.ButtonDown, false)
		case "ArrowLeft":
			console.SetJoypad(nes.ButtonLeft, false)
		case "ArrowRight":
			console.SetJoypad(nes.ButtonRight, false)
		}
		return nil
	}
	js.Global().Set("keyup", js.FuncOf(keyup))

	<-c
}
