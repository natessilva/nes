package main

import (
	"image"
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
	ppu := nes.NewPPU(cart, image.NewRGBA(image.Rect(0, 0, 256, 240)))
	cpu := nes.NewCPU(cart, ppu)
	for {
		cycles := cpu.Step()
		cycles *= 3
		beforeNMI := ppu.NMITriggered()
		for ; cycles > 0; cycles-- {
			ppu.Step()
		}
		afterNMI := ppu.NMITriggered()
		if !beforeNMI && afterNMI {
			cpu.TriggerNMI()
			break
		}
	}
}
