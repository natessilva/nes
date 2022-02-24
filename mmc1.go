package nes

import (
	"log"
)

type mmc1 struct {
	mirrorMode byte
	prg        []byte
	chr        []byte
	sram       [0x2000]byte

	// registers are written to by
	// first write to the shift register
	// 5 times, shifting 1 bit at a time
	shift    byte
	ctrl     byte
	chrBank0 byte
	chrBank1 byte
	prgBank  byte

	// after writing to the registers
	// figure out our bank offsets
	prgOffsets [2]int
	chrOffsets [2]int
}

func newMMC1(mirror byte, prg, chr []byte) *mmc1 {
	m := &mmc1{
		mirrorMode: mirror,
		prg:        prg,
		chr:        chr,
		shift:      0x10,
	}
	m.prgOffsets[1] = len(prg) - 0x4000
	return m
}

func (n *mmc1) readByte(address uint16) byte {
	switch {
	case address < 0x2000:
		bank := address / 0x1000
		bankOffset := address % 0x1000
		return n.chr[n.chrOffsets[bank]+int(bankOffset)]
	case address >= 0x8000:
		bank := (address - 0x8000) / 0x4000
		bankOffset := address % 0x4000
		return n.prg[n.prgOffsets[bank]+int(bankOffset)]
	case address >= 0x6000:
		index := int(address - 0x6000)
		return n.sram[index]
	default:
		log.Fatalf("mmc1 invalid read address %04X", address)
	}
	return 0
}

func (n *mmc1) write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		bank := address / 0x1000
		bankOffset := address % 0x1000
		n.chr[n.chrOffsets[bank]+int(bankOffset)] = value
	case address >= 0x8000:
		n.loadRegister(address, value)
	case address >= 0x6000:
		index := int(address-0x6000) % len(n.sram)
		n.sram[index] = value
	default:
		log.Fatalf("mmc1 invalid write address %04X", address)
	}
}

func (n *mmc1) loadRegister(address uint16, value byte) {
	// if bit 7 is hi, reset the shift register
	if value&0x80 == 0x80 {
		n.shift = 0x10
		n.ctrl |= 0x0C
		n.evaluateRegisters()
		return
	}
	// sr is full when the 1 bit is set
	full := n.shift&1 == 1

	//load the lo bit into the fifth position
	n.shift = (n.shift >> 1) | ((value & 1) << 4)

	if full {
		n.writeRegister(address)
		n.evaluateRegisters()
		n.shift = 0x10
	}
}

func (n *mmc1) writeRegister(address uint16) {
	switch {
	case address < 0xA000:
		n.ctrl = n.shift
	case address < 0xC000:
		n.chrBank0 = n.shift
	case address < 0xE000:
		n.chrBank1 = n.shift
	case address >= 0xE000:
		n.prgBank = n.shift & 0x0F
	default:
		log.Fatalf("invalid register address %04X", address)
	}
}

func (n *mmc1) evaluateRegisters() {
	mirror := n.ctrl & 3
	prgMode := (n.ctrl >> 2) & 3
	chrMode := (n.ctrl >> 4) & 1
	numPrg := len(n.prg) / 0x4000
	numChr := len(n.chr) / 0x1000

	switch prgMode {
	case 0, 1:
		// switch 32 KB at $8000, ignoring low bit of bank number
		// n.prgOffsets
		bank := int(n.prgBank&0x0E) % numPrg
		n.prgOffsets[0] = bank * 0x4000
		n.prgOffsets[1] = (bank + 1) * 0x4000
	case 2:
		// fix first bank at $8000 and switch 16 KB bank at $C000
		bank := int(n.prgBank) % numPrg
		n.prgOffsets[0] = 0
		n.prgOffsets[1] = bank * 0x4000
	case 3:
		bank := int(n.prgBank) % numPrg
		n.prgOffsets[0] = bank * 0x4000
		n.prgOffsets[1] = len(n.prg) - 0x4000
	}

	if chrMode == 0 {
		// 8k mode
		bank := int(n.chrBank0&0x1E) % numChr
		n.chrOffsets[0] = bank * 0x1000
		n.chrOffsets[1] = n.chrOffsets[0] + 0x1000
	} else {
		// 4k mode
		n.chrOffsets[0] = (int(n.chrBank0) % numChr) * 0x1000
		n.chrOffsets[1] = (int(n.chrBank1) % numChr) * 0x1000
	}

	switch mirror {
	case 0:
		n.mirrorMode = mirrorSingle0
	case 1:
		n.mirrorMode = mirrorSingle1
	case 2:
		n.mirrorMode = mirrorVertical
	case 3:
		n.mirrorMode = mirrorHorizontal
	}
}

func (n *mmc1) mirror(address uint16) uint16 {
	return mirror(n.mirrorMode, address)
}
