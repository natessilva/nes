package nes

import (
	"fmt"
	"log"
)

// PPUCTRL bits
const (
	CTRL_N1 byte = 1 << iota
	CTRL_N2
	CTRL_I // VRAM address increment per CPU read/write of PPUDATA (0: add 1, going across; 1: add 32, going down)
	CTRL_S // Sprite pattern table address for 8x8 sprites (0: $0000; 1: $1000; ignored in 8x16 mode)
	CTRL_B // Background pattern table address (0: $0000; 1: $1000)
	CTRL_H // Sprite size (0: 8x8 pixels; 1: 8x16 pixels)
	CTRL_P // PPU master/slave select (0: read backdrop from EXT pins; 1: output color on EXT pins)
	CTRL_V // Generate an NMI at the start of the vertical blanking interval (0: off; 1: on)
)

const FLAG_NMI_OUTPUT = CTRL_V

const CTRL_N = CTRL_N1 | CTRL_N2 // Base nametable address

// PPUMASK bits
const (
	MASK_GRAY byte = 1 << iota // Grayscale
	MASK_BG_L                  // Show background in leftmost 8 pixels of screen
	MASK_SP_L                  // Show sprites in leftmost 8 pixels of screen
	MASK_BG                    // Show background
	MASK_SP                    // Show sprites
	MASK_R                     // Emphasize red
	MASK_G                     // Emphasize green
	MASK_B                     // Emphasize blue
)

const RENDER_ENABLED = MASK_BG | MASK_SP

// PPUSTATUS bits
const (
	STATUS_O byte = 1 << (iota + 5)
	STATUS_S
	STATUS_V
)

const FLAG_NMI_OCCURRED = STATUS_V

type PPU struct {
	cart *Cart

	// Cycle counting
	cycle    int // There are 341 cycles per scanline
	scanline int // There are 262 scanlines
	Frame    int // 341*262 cycles per frame

	// Internal registers
	v   uint16 // Current VRAM address (15 bits)
	t   uint16 // Temporary VRAM address (15 bits); can also be thought of as the address of the top left onscreen tile.
	x   byte   // Fine X scroll (3 bits)
	w   byte   // First or second write toggle (1 bit)
	odd bool   // is the frame odd?

	ctrl   byte
	mask   byte
	status byte

	// 2kb of nameTable stores the nameTable
	nameTable [0x800]byte
}

func NewPPU(cart *Cart) *PPU {
	return &PPU{
		cart: cart,
	}
}

// registers are repeated every 8 bytes
func (p *PPU) ReadRegister(address uint16) byte {
	switch address % 8 {
	// PPUSTATUS
	case 2:
		status := p.status
		p.status = resetBits(p.status, STATUS_V)
		p.w = 0
		return status
	case 4:
	case 7:
	}
	return 0
}

func (p *PPU) WriteRegister(address uint16, value byte) {
	switch address % 8 {
	case 0:
		fmt.Println("write status")
		p.ctrl = value
		p.t |= ((uint16(value) & 3) << 10)
		p.nmi()
	case 1:
		fmt.Println("write mask")
		p.mask = value
	case 5:
		fmt.Println("write scroll")
		if p.w == 0 {
			p.t |= (uint16(value) >> 3)
			p.x = value & 7
			p.w = 1
		} else {
			p.t |= ((uint16(value) & 0x07) << 12)
			p.t |= ((uint16(value) & 0xF8) << 2)
			p.w = 0
		}
	case 6:
		fmt.Println("write address")
		if p.w == 0 {
			p.t = ((uint16(value) & 0x3F) << 8)
			p.w = 1
		} else {
			p.t |= uint16(value)
			p.v = p.t
			p.w = 0
		}
	case 7:
		p.Write(p.v, value)
		if !isAnySet(p.ctrl, CTRL_I) {
			p.v += 1
		} else {
			p.v += 32
		}
	default:
		log.Fatalf("write invalid register %04x\n", address)
	}
}

func (p *PPU) Write(address uint16, value byte) {
	address %= 0x4000
	switch {
	// pattern table
	case address < 0x2000:
		log.Fatalf("pattern table write not implemented yet")
	// name table
	case address < 0x3F00:
		p.nameTable[mirror(p.v, p.cart.Mirror)] = value
	// palette ram
	case address < 0x4000:
		log.Fatalf("palette  RAM not implemented yet")
	default:
		log.Fatalf("invalid write address: %02x\n", address)
	}
}

// Step the PPU forward 1 cycle
func (p *PPU) Step() {
	p.cycle++

	renderingEnabled := isAnySet(p.mask, RENDER_ENABLED)

	// skip the last cycle in odd scanlines when rendering is enabled
	if p.odd && p.scanline == 261 && p.cycle == 340 && renderingEnabled {
		p.cycle = 0
		p.scanline = 0
		p.Frame++
		p.odd = !p.odd
	}

	// have we finished a scanline?
	if p.cycle > 340 {
		p.cycle = 0
		p.scanline++

		// have we finished a frame?
		if p.scanline > 261 {
			p.scanline = 0
			p.Frame++
			p.odd = !p.odd
		}
	}

	// vblank starts on the second cycle of scanline 241
	if p.scanline == 241 && p.cycle == 1 {
		p.setVBlank()
	}

	// vblank ends on the second cycle of scanline 241
	if p.scanline == 261 && p.cycle == 1 {
		p.clearVBlank()
	}
}

func (p *PPU) setVBlank() {
	fmt.Println("setVBlank")
	p.status = setBits(p.status, STATUS_V)
	p.nmi()
}

func (p *PPU) clearVBlank() {
	fmt.Println("clearVBlank")
	p.status = resetBits(p.status, STATUS_V)
	p.nmi()
}

// generate NMI if necessary
func (p *PPU) nmi() {
	if isAnySet(p.status, STATUS_V) && isAnySet(p.ctrl, CTRL_V) {
		log.Fatal("NMI not yet implemented")
	} else {
		fmt.Println("no NMI")
	}
}

// there are 2 1k name tables mirrored across a 4k address space
// starting a $2000
func mirror(address uint16, mode byte) uint16 {
	// translate the address onto the 4k space
	address = (address - 0x2000) % 0x1000
	// which of the 4 tables?
	table := address / 0x400
	// where at within the table
	tableAddress := address % 0x400

	// horizontal 0011
	if mode == H_MIRROR {
		table /= 2
		return table*0x400 + tableAddress
	} else {
		log.Fatalln("vertical mirroring not yet implemented")
	}
	return 0
}
