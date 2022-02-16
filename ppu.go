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

type ppu struct {
	cart *cartridge

	// 2 screens worth of ram
	vram [2048]byte

	// 32 bytes, background color is repeated
	// every 4 bytes. 13 background colors and
	// 12 sprite colors
	paletteTable [32]byte

	// 64 sprites 4 bytes each
	oamAddr byte
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

	// data reads are buffered
	readBuffer byte

	// scroll
	xScroll byte
	yScroll byte
}

func newPPU(cart *cartridge) *ppu {
	return &ppu{
		cart: cart,
	}
}

func (p *ppu) readRegister(address uint16) byte {
	switch address {
	case 2:
		// STATUS 3 bits plus the remain bits filled by the latch
		value := (p.status & 0xE0) | p.latch
		p.status = resetBits(p.status, statusV)
		p.w = false
		return value
	case 7:
		buff := p.readBuffer
		value := p.readByte(p.addr)
		// 0-3EFF is buffered
		if p.addr < 0x3F00 {
			p.readBuffer = value
			value = buff
		} else {
			// Reading the palettes still updates the internal buffer though,
			//but the data placed in it is the mirrored nametable data that
			// would appear "underneath" the palette.
			p.readBuffer = p.readByte(p.addr - 0x1000)
		}
		if !isAnySet(p.ctrl, ctrlI) {
			p.addr += 1
		} else {
			p.addr += 32
		}
		return value
	default:
		log.Fatalf("invalid register read address %d", address)
	}
	return 0
}

func (p *ppu) writeRegister(address uint16, value byte) {
	// the latch is always written to for every write
	p.latch = value
	switch address {
	case 0:
		p.ctrl = value
	case 1:
		p.mask = value
	case 3:
		p.oamAddr = value
	case 5:
		if !p.w {
			p.xScroll = value
			p.w = true
		} else {
			p.yScroll = value
			p.w = false
		}
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

// writeDMA will be called 256 times in sequence by the CPU
func (p *ppu) writeDMA(value byte) {
	p.oamData[p.oamAddr] = value
	p.oamAddr++
}

func (p *ppu) readByte(address uint16) byte {
	switch {
	case address < 0x2000:
		return p.cart.readByte(address)
	case address < 0x3F00:
		return p.vram[mirror(address)]
	case address < 0x4000:
		return p.paletteTable[address%32]
	default:
		log.Fatalf("invalid ppu read address %04X\n", address)
	}
	return 0
}

func (p *ppu) write(address uint16, value byte) {
	switch {
	case address < 0x3F00:
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
	table %= 2

	// where at within the table
	location := address % 0x400

	return table*0x400 + location
}

func (p *ppu) Step() {
	p.cycle++

	renderingEnabled := isAnySet(p.mask, maskBG|maskSP)

	// On odd rendered frames there is one less tick
	if renderingEnabled && p.cycle > 339 && p.scanline > 261 && p.odd {
		p.cycle = 0
		p.scanline = 0
	}

	if p.cycle > 340 {
		// sprite zero detection
		x := int(p.oamData[0])
		y := int(p.oamData[3])
		if y == p.scanline && x < p.cycle && isAnySet(p.mask, maskSP) {
			p.status = setBits(p.status, statusS)
		}

		p.cycle = 0
		p.scanline++

		if p.scanline > 261 {
			p.scanline = 0
		}
	}
	p.odd = !p.odd

	// vblank
	if p.cycle == 1 && p.scanline == 241 {
		p.status = resetBits(p.status, statusS)
		p.status = setBits(p.status, statusV)
	}

	if p.scanline == 261 && p.cycle == 1 {
		p.status = 0
	}
}

func (p *ppu) NMITriggered() bool {
	return isAnySet(p.status, statusV) && isAnySet(p.ctrl, ctrlV)
}

func (p *ppu) render(image *image.RGBA) {
	if !isAnySet(p.mask, maskBG|maskSP) {
		return
	}

	baseNameTable := int((p.ctrl & ctrlN) % 2)

	for tile := 0; tile < 960; tile++ {
		tileAddress := uint16(p.vram[baseNameTable*1024+tile]) * 16
		if isAnySet(p.ctrl, ctrlB) {
			tileAddress += 0x1000
		}
		tileY := tile / 32
		tileX := tile % 32
		tileBytes := p.cart.chr[tileAddress : tileAddress+16]

		metaX := tileX / 4
		metaY := tileY / 4
		metaIndex := metaY*8 + metaX
		attr := p.vram[(baseNameTable*1024+960+metaIndex)%2048]
		shift := ((tile >> 4) & 4) | (tile & 2)

		paletteIndex := ((attr >> shift) & 3) << 2

		colors := [4]color.RGBA{
			palette[p.paletteTable[0]],
			palette[p.paletteTable[paletteIndex+1]],
			palette[p.paletteTable[paletteIndex+2]],
			palette[p.paletteTable[paletteIndex+3]],
		}

		for y := 0; y < 8; y++ {
			lo := tileBytes[y]
			hi := tileBytes[y+8]

			for x := 7; x >= 0; x-- {
				value := (hi&1)<<1 | (lo & 1)
				hi >>= 1
				lo >>= 1

				if isAnySet(p.mask, maskBG) {
					image.Set(tileX*8+x-int(p.xScroll), tileY*8+y, colors[value])
				} else {
					image.Set(tileX*8+x, tileY*8+y, color.Black)

				}

			}
		}

	}

	for tile := 0; tile < 960; tile++ {
		tileAddress := uint16(p.vram[((baseNameTable+1)*1024+tile)%2048]) * 16
		if isAnySet(p.ctrl, ctrlB) {
			tileAddress += 0x1000
		}
		tileY := tile / 32
		tileX := tile % 32
		tileBytes := p.cart.chr[tileAddress : tileAddress+16]

		metaX := tileX / 4
		metaY := tileY / 4
		metaIndex := metaY*8 + metaX
		attr := p.vram[((baseNameTable+1)*1024+960+metaIndex)%2048]
		shift := ((tile >> 4) & 4) | (tile & 2)

		paletteIndex := ((attr >> shift) & 3) << 2

		colors := [4]color.RGBA{
			palette[p.paletteTable[0]],
			palette[p.paletteTable[paletteIndex+1]],
			palette[p.paletteTable[paletteIndex+2]],
			palette[p.paletteTable[paletteIndex+3]],
		}

		for y := 0; y < 8; y++ {
			lo := tileBytes[y]
			hi := tileBytes[y+8]

			for x := 7; x >= 0; x-- {
				value := (hi&1)<<1 | (lo & 1)
				hi >>= 1
				lo >>= 1

				if isAnySet(p.mask, maskBG) {
					image.Set(tileX*8+x-int(p.xScroll)+256, tileY*8+y, colors[value])
				} else {
					image.Set(tileX*8+x, tileY*8+y, color.Black)

				}

			}
		}

	}
	if isAnySet(p.mask, maskSP) {
		for i := 256; i > 0; i -= 4 {
			tileX := int(p.oamData[i-1])
			tileY := int(p.oamData[i-4])
			tileIndex := int(p.oamData[i-3])
			attr := p.oamData[i-2]

			// behind background?
			behind := attr&32 == 32

			tileAddress := tileIndex * 16

			if isAnySet(p.ctrl, ctrlS) {
				tileAddress += 0x1000
			}
			tileBytes := p.cart.chr[tileAddress : tileAddress+16]

			paletteIndex := ((attr & 3) + 4) * 4

			colors := [3]color.RGBA{
				palette[p.paletteTable[paletteIndex+1]],
				palette[p.paletteTable[paletteIndex+2]],
				palette[p.paletteTable[paletteIndex+3]],
			}
			flipY := (attr>>7)&1 != 1
			flipX := (attr>>6)&1 != 1

			for y := 0; y < 8; y++ {
				lo := tileBytes[y]
				hi := tileBytes[y+8]
				for x := 7; x >= 0; x-- {
					value := (hi&1)<<1 | (lo & 1)
					hi >>= 1
					lo >>= 1

					fy := y
					fx := x
					if !flipY {
						fy = 7 - y
					}
					if !flipX {
						fx = 7 - x
					}

					// 0 is BG
					if value != 0 {
						if !behind || image.At(tileX+fx, tileY+fy) == palette[p.paletteTable[0]] {
							image.Set(tileX+fx, tileY+fy, colors[value-1])
						}
					}

				}
			}

		}
	}
}
