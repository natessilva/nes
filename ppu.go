package nes

import "log"

type PPU struct {
	cart *Cart

	// 2 screens worth of ram
	vram [2048]byte

	// 32 bytes, background color is repeated
	// every 4 bytes. 13 background colors and
	// 12 sprite colors
	paletteTable [32]byte

	// 64 sprites 4 bytes each
	oamData [256]byte

	// render timing
	cycle    int
	scanLine int
	odd      bool
}

func NewPPU(cart *Cart) *PPU {
	return &PPU{
		cart: cart,
	}
}

func (p *PPU) readRegister(address uint16) byte {
	switch address {
	default:
		log.Fatalf("invalid register read address %d", address)
	}
	return 0
}

func (p *PPU) writeRegister(address uint16, value byte) {
	switch address {
	default:
		log.Fatalf("invalid register write address %d", address)
	}
}

func (p *PPU) Step() {
	p.cycle++

	// TODO implement rendering
	renderingEnabled := false

	// On odd rendered frames there is one less tick
	if renderingEnabled && p.cycle > 339 && p.scanLine > 261 && p.odd {
		p.cycle = 0
		p.scanLine = 0
	}

	if p.cycle > 340 {
		p.cycle = 0
		p.scanLine++

		if p.scanLine > 261 {
			p.scanLine = 0
		}
	}
	p.odd = !p.odd
}
