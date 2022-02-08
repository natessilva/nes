package main

import (
	"log"
	"os"

	"github.com/natessilva/nes"
)

func main() {
	file := os.Args[1]
	cart, err := nes.LoadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	cpu := nes.NewCPU(cart)
	ppu := nes.NewPPU()
	// todo this is a weird cyclical dedendancy
	cpu.PPU = ppu
	for {
		cycles := cpu.Step()
		cycles *= 3
		for ; cycles > 0; cycles-- {
			ppu.Step()
		}
		if ppu.Frame == 2 {
			break
		}
	}
}
