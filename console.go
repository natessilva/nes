package nes

import (
	"image"
	"io"
)

type Console struct {
	ppu *ppu
	cpu *cpu
}

func NewConsole() *Console {
	return &Console{}
}

func (n *Console) LoadROM(r io.Reader) error {
	cart, err := readFile(r)
	if err != nil {
		return err
	}
	n.ppu = newPPU(cart)
	n.cpu = newCPU(cart, n.ppu)
	return nil
}

func (n *Console) RenderFrame(image *image.RGBA) {
	for {
		cycles := n.cpu.Step()
		cycles *= 3
		beforeNMI := n.ppu.NMITriggered()
		for ; cycles > 0; cycles-- {
			n.ppu.Step()
		}
		afterNMI := n.ppu.NMITriggered()
		if !beforeNMI && afterNMI {
			n.ppu.render(image)
			n.cpu.TriggerNMI()
			break
		}
	}
}
