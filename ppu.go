package nes

import (
	"image"
	"image/color"
	"log"
)

const (
	statusO = 1 << (5 + iota) // sprite overflow
	statusS                   // sprite 0 hit
	statusV                   // vertical blanking has started
)

const (
	ctrlN = 3               //base nametable address
	ctrlI = 1 << (1 + iota) //vram address increment
	ctrlS                   // sprite pattern table
	ctrlB                   //background pattern table
	ctrlH                   // sprite size
	ctrlP                   // ppu master/slave select
	ctrlV                   // generate NMI on start of vblank
)

const (
	maskGr = 1 << iota
	maskBGL
	maskSPL
	maskBG
	maskSP
	maskR
	maskG
	maskB
)

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
	scanline int
	odd      bool

	// registers
	ctrl   byte
	mask   byte
	status byte
	latch  byte
	addr   uint16
	w      bool

	// render frame
	image *image.RGBA
}

func NewPPU(cart *Cart, image *image.RGBA) *PPU {
	return &PPU{
		cart:  cart,
		image: image,
	}
}

func (p *PPU) readRegister(address uint16) byte {
	switch address {
	case 2:
		// STATUS 3 bits plus the remain bits filled by the latch
		value := (p.status & 0xE0) | p.latch
		p.status = resetBits(p.status, statusV)
		p.w = false
		return value
	default:
		log.Fatalf("invalid register read address %d", address)
	}
	return 0
}

func (p *PPU) writeRegister(address uint16, value byte) {
	// the latch is always written to for every write
	p.latch = value
	switch address {
	case 0:
		p.ctrl = value
	case 1:
		p.mask = value
	case 5:
		// TODO implement scroll
	case 6:
		if !p.w {
			p.addr = (uint16(value) & 0x3f) << 8
			p.w = true
		} else {
			p.addr |= uint16(value)
			p.w = false
		}
	case 7:
		p.write(p.addr, value)
		if !isAnySet(p.ctrl, ctrlI) {
			p.addr += 1
		} else {
			p.addr += 32
		}
	default:
		log.Fatalf("invalid register write address %d", address)
	}
}

func (p *PPU) write(address uint16, value byte) {
	switch {
	case address < 0x3000:
		p.vram[mirror(address)] = value
	case address < 0x4000:
		p.paletteTable[address%32] = value
	default:
		log.Fatalf("invalid ppu write address %04X\n", address)
	}
}

// mirror the vram nameTables
func mirror(address uint16) uint16 {
	// map to the vram address space
	address = (address - 0x2000) % 0x1000
	// which of the 4 name tables
	table := address / 0x400

	//TODO horizontal is hardcoded for now
	table /= 2

	// where at within the table
	location := address % 0x400

	return table*0x400 + location
}

func (p *PPU) Step() {
	p.cycle++

	// TODO implement rendering
	renderingEnabled := isAnySet(p.mask, maskBG|maskSP)

	// On odd rendered frames there is one less tick
	if renderingEnabled && p.cycle > 339 && p.scanline > 261 && p.odd {
		p.cycle = 0
		p.scanline = 0
	}

	if p.cycle > 340 {
		p.cycle = 0
		p.scanline++

		if p.scanline > 261 {
			p.scanline = 0
		}
	}
	p.odd = !p.odd

	// vblank
	if p.cycle == 1 && p.scanline == 241 {
		p.status = setBits(p.status, statusV)

		if p.NMITriggered() && renderingEnabled {
			p.render()
		}
	}

	if p.scanline == 261 && p.cycle == 1 {
		p.status = 0
	}
}

func (p *PPU) NMITriggered() bool {
	return isAnySet(p.status, statusV) && isAnySet(p.ctrl, ctrlV)
}

func (p *PPU) render() {
	for tile := 0; tile < 960; tile++ {
		tileAddress := uint16(p.vram[tile]) * 16
		tileY := tile / 32
		tileX := tile % 32
		tileBytes := p.cart.CHR[tileAddress : tileAddress+16]

		colors := []color.RGBA{
			{0, 0, 0, 255},
			{85, 85, 85, 255},
			{170, 170, 170, 255},
			{255, 255, 255, 255},
		}

		for y := 0; y < 8; y++ {
			hi := tileBytes[y]
			lo := tileBytes[y+8]

			for x := 7; x >= 0; x-- {
				value := (hi&1)<<1 | (lo & 1)
				hi >>= 1
				lo >>= 1
				p.image.Set(tileX*8+x, tileY*8+y, colors[value])

			}
		}

	}
}
