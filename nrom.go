package nes

import "log"

type nROM struct {
	mirrorMode byte
	prg        []byte
	chr        []byte
}

func (n *nROM) readByte(address uint16) byte {
	switch {
	case address < 0x2000:
		return n.chr[address]
	case address >= 0xC000:
		index := int(address - 0x8000)
		return n.prg[index%len(n.prg)]
	case address >= 0x8000:
		index := int(address - 0x8000)
		return n.prg[index%len(n.prg)]
	default:
		log.Fatalf("invalid address %04x\n", address)
	}
	return 0
}

func (n *nROM) write(address uint16, value byte) {
	log.Fatal("NROM doesn't support writes")
}

func (n *nROM) mirror(address uint16) uint16 {
	return mirror(n.mirrorMode, address)
}
