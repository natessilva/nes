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

	// data reads are buffered
	readBuffer byte

	// vram address and registers

	// The 15 bit registers t and v are composed this way during rendering:
	// yyy NN YYYYY XXXXX
	// ||| || ||||| +++++-- coarse X scroll
	// ||| || +++++-------- coarse Y scroll
	// ||| ++-------------- nametable select
	// +++----------------- fine Y scroll

	v       uint16 // Current VRAM address (15 bits)
	t       uint16 // Temporary VRAM address (15 bits); can also be thought of as the address of the top left onscreen tile.
	x       byte   // Fine X scroll (3 bits)
	w       bool   // First or second write toggle (1 bit)
	xScroll byte
	yScroll byte

	// rendering shift registers
	nameTableByte        byte
	attributeTableByte   byte
	patternTableLowByte  byte
	patternTableHighByte byte
	// prepared 4 bit pixel data for 16 horizontal pixels
	pixelData uint64
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
		value := p.readByte(p.v)
		// 0-3EFF is buffered
		if p.v < 0x3F00 {
			p.readBuffer = value
			value = buff
		} else {
			// Reading the palettes still updates the internal buffer though,
			//but the data placed in it is the mirrored nametable data that
			// would appear "underneath" the palette.
			p.readBuffer = p.readByte(p.v - 0x1000)
		}
		if !isAnySet(p.ctrl, ctrlI) {
			p.v += 1
		} else {
			p.v += 32
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
		// copy nametable select into temp vram
		p.t = (p.t & 0xF3FF) | ((uint16(value) & ctrlN) << 10)
	case 1:
		p.mask = value
	case 3:
		p.oamAddr = value
	case 5:
		// scroll is a byte (0-255) and can be broken into 2 parts
		// the high 5 bits are the coarse scroll and represent the
		// tile index in vram. The low 3 bits are the fine scroll
		// and represent the pixel offset within the tile.
		if !p.w {
			// write coarseX into the temp address
			p.t = (p.t & 0xFFE0) | (uint16(value) >> 3)
			// fine x has its own register
			p.x = value & 7
			p.xScroll = value
			p.w = true
		} else {
			// write coarseY and fineY into the temp address
			p.t = (p.t & 0x8C1F) | ((uint16(value) & 0xF8) << 2) | ((uint16(value) & 3) << 12)
			p.yScroll = value
			p.w = false
		}
	case 6:
		if !p.w {
			p.t = (p.t & 0x80FF) | ((uint16(value) & 0x3F) << 8)
			p.w = true
		} else {
			p.t = (p.t & 0xFF00) | uint16(value)
			p.v = p.t
			p.w = false
		}
	case 7:
		p.write(p.v, value)
		if !isAnySet(p.ctrl, ctrlI) {
			p.v += 1
		} else {
			p.v += 32
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

func (p *ppu) Step(image *image.RGBA) {
	p.cycle++

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

	visibleScanLine := p.scanline < 240
	preRenderScanLine := p.scanline == 261
	fetchScanLine := preRenderScanLine || visibleScanLine

	visibleCycle := 1 <= p.cycle && p.cycle <= 256
	preRenderCycle := 321 <= p.cycle && p.cycle <= 336
	fetchCycle := preRenderCycle || visibleCycle

	copyYCycle := 280 <= p.cycle && p.cycle <= 304

	// cycle accurate vram address manipulation
	if renderingEnabled {
		if visibleCycle && visibleScanLine {
			p.renderBackgroundPixel(image)
		}

		// fetch the data for the background pixels
		// we get the 4 bytes in cycles 1,3,5,7
		// and then we combine them into 8 pixels
		// worth of 4 bit pixel data or 32 bits of
		// information.
		// Since we prefetch 2 tiles worth of data
		// our pixelData shift register is 64 bits
		// wide and we shift off 4 bits per pixel
		// rendered.
		if fetchScanLine && fetchCycle {
			p.pixelData <<= 4
			switch p.cycle % 8 {
			case 0:
				p.preparePixelData()
			case 1:
				p.getNameTableByte()
			case 3:
				p.getAttributeTableByte()
			case 5:
				p.getPatternTableLowByte()
			case 7:
				p.getPatternTableHighByte()
			}
		}
		if fetchScanLine {
			if p.cycle == 256 {
				p.incrementY()
			}
			if p.cycle == 257 {
				p.copyX()
			}
			if fetchCycle && p.cycle%8 == 0 {
				p.incrementX()
			}
		}
		if copyYCycle && preRenderScanLine {
			p.copyY()
		}
	}

	if renderingEnabled && p.cycle > 0 && p.cycle <= 256 && p.scanline < 240 {

		// sprite zero detection
		x := int(p.oamData[3]) + 8
		y := int(p.oamData[0]) + 8
		if y == p.scanline && x == p.cycle+1 && isAnySet(p.mask, maskSP) {
			p.status = setBits(p.status, statusS)
		}
	}

	// vblank
	if p.cycle == 1 && p.scanline == 241 {
		p.status = setBits(p.status, statusV)
	}

	if p.scanline == 261 && p.cycle == 1 {
		p.status = 0
		p.ctrl &= 0xFC
	}
}

func (p *ppu) getNameTableByte() {
	address := 0x2000 | (p.v & 0x0FFF)
	p.nameTableByte = p.readByte(address)
}

func (p *ppu) getAttributeTableByte() {
	address := 0x23C0 | (p.v & 0x0C00) | ((p.v >> 4) & 0x38) | ((p.v >> 2) & 0x07)
	p.attributeTableByte = p.readByte(address)
}

func (p *ppu) getPatternTableLowByte() {
	y := (p.v >> 12) & 7
	nameTable := uint16((p.ctrl & ctrlB) >> 4)
	tileIndex := uint16(p.nameTableByte)
	address := 0x1000*uint16(nameTable) + tileIndex*16 + y
	p.patternTableLowByte = p.readByte(address)
}

func (p *ppu) getPatternTableHighByte() {
	y := (p.v >> 12) & 7
	nameTable := uint16((p.ctrl & ctrlB) >> 4)
	tileIndex := uint16(p.nameTableByte)
	address := 0x1000*uint16(nameTable) + tileIndex*16 + y
	p.patternTableHighByte = p.readByte(address + 8)
}

// preparePixelData takes the information from the
// nameTableByte, attributeTableByte, low and high
// pattern table bytes and converts them into 8 pixels
// worth of 4 bits of palette index data
// Though we are preparing a single tile, we prefetch
// two tiles before starting to render a line, so we
// store the result in the low half of a 64 bit register.
// this is to account for the fineX scroll. A give 8
// pixel section might be split across two tiles.
func (p *ppu) preparePixelData() {
	attr := p.attributeTableByte
	shift := ((p.v >> 4) & 4) | (p.v & 2)

	paletteIndex := ((attr >> shift) & 3) << 2
	lo := p.patternTableLowByte
	hi := p.patternTableHighByte

	var pixelData uint64
	for x := 0; x < 8; x++ {
		// value is a number from 0-15 representing an
		// index into the background section of the palette
		value := paletteIndex | ((hi & 0x80) >> 6) | ((lo & 0x80) >> 7)
		hi <<= 1
		lo <<= 1
		pixelData = (pixelData << 4) | uint64(value)
	}

	// we now have the 32 pixels representing 8 palette indices
	// as we read the next 8 pixels we'll shift this left 32 bits
	p.pixelData |= pixelData
}

// using the prepared pixelData in combination with the fineX scroll
// figure out the palette index of a single pixel and render it.
func (p *ppu) renderBackgroundPixel(image *image.RGBA) {
	firstTilePixelData := p.pixelData >> 32

	// because we are always shifting our prepared pixel data by one pixel
	// we can always use the same offset to get the next pixel
	shift := (7 - p.x) * 4
	pixelData := byte((firstTilePixelData >> shift) & 0x0F)

	x := p.cycle - 1
	y := p.scanline

	// background color
	if pixelData%4 == 0 {
		pixelData = 0
	}
	image.Set(x, y, palette[p.paletteTable[pixelData]])
}

func (p *ppu) NMITriggered() bool {
	return isAnySet(p.status, statusV) && isAnySet(p.ctrl, ctrlV)
}

func (p *ppu) render(image *image.RGBA) {
	if !isAnySet(p.mask, maskBG|maskSP) {
		return
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

// incrementX during rendering after each 8 pixels is rendered
// Each nametable is 32 tiles wide. When scrolling, we might
// render parts of 2 nametables. If we've reached the end of a
// nametable, wrap around back to zero and switch to the next
// horizontal nametable.
func (p *ppu) incrementX() {
	if p.v&0x001F == 31 {
		// set coarseX to zero
		p.v &= 0xFFE0
		// toggle the low nametable select bit
		p.v ^= 0x400
	} else {
		p.v++
	}
}

// incrementY during rendering after each scanLine. We increment
// fineY (0-7), wrapping around to zero and incrementing coarseY
// if we wrap. When we increment coarseY (0-30), wrapping around
// to zero and switching to the next vertical nametable
func (p *ppu) incrementY() {
	// if fine Y < 7
	if (p.v & 0x7000) != 0x7000 {
		// increment fine Y
		p.v += 0x1000
	} else {
		// fine Y = 0
		p.v &= 0x8FFF
		// let y = coarse Y
		y := (p.v & 0x03E0) >> 5
		if y == 29 {
			y = 0
			// switch vertical nametable
			p.v ^= 0x0800
		} else if y == 31 {
			// coarse Y = 0, nametable not switched
			y = 0
		} else {
			// increment coarse Y
			y += 1
			// put coarse Y back into v
			p.v = (p.v & 0xFC1F) | (y << 5)
		}
	}
}

// copy the horizontal information from t after each scanline
// reseting the x position back to the left side of the screen
// and resetting the horizontal bit of the nametable
func (p *ppu) copyX() {
	p.v = (p.v & 0xFBE0) | (p.t & 0x041F)
}

// copy the vertical information from t after each frame
// resetting the y position back to the top of the screen
// and resetting the vertical bit of the nametable
func (p *ppu) copyY() {
	p.v = (p.v & 0x841F) | (p.t & 0x7BE0)
}
