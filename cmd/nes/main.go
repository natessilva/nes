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
	ppu := nes.NewPPU(cart)
	cpu := nes.NewCPU(cart, ppu)
	for {
		cycles := cpu.Step()
		cycles *= 3
		for ; cycles > 0; cycles-- {
			// ppu.Step()
		}
	}
}
