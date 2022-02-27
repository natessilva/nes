package nes

import "log"

type unROM struct {
	mirrorMode byte
	prg        []byte
	chr        []byte

	prgBank byte
}

func (n *unROM) readByte(address uint16) byte {
	switch {
	case address < 0x2000:
		return n.chr[address]
	case address >= 0xC000:
		// fixed to the last bank
		index := int(address-0xC000) + (len(n.prg) - 0x4000)
		return n.prg[index]
	case address >= 0x8000:
		index := int(address-0x8000) + int(n.prgBank)*0x4000
		return n.prg[index]
	default:
		log.Fatalf("invalid read address %04X", address)
	}
	return 0
}

func (n *unROM) write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		n.chr[address] = value
	case address >= 0x8000:
		numPrg := byte(len(n.prg) / 0x4000)
		n.prgBank = value % numPrg
	default:
		log.Fatalf("invalid write address %04X", address)
	}
}

func (n *unROM) mirror(address uint16) uint16 {
	return mirror(n.mirrorMode, address)
}
