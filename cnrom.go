package nes

import "log"

type cnROM struct {
	mirrorMode byte
	prg        []byte
	chr        []byte

	chrBank byte
}

func (n *cnROM) readByte(address uint16) byte {
	switch {
	case address < 0x2000:
		return n.chr[int(n.chrBank)*0x2000+int(address)]
	case address >= 0x8000:
		index := int(address - 0x8000)
		return n.prg[index%len(n.prg)]
	default:
		log.Fatalf("invalid read address %04X", address)
	}
	return 0
}

func (n *cnROM) write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		n.chr[int(n.chrBank)*0x2000+int(address)] = value
	case address >= 0x8000:
		n.chrBank = value & 3
	default:
		log.Fatalf("invalid write address %04X", address)
	}
}

func (n *cnROM) mirror(address uint16) uint16 {
	return mirror(n.mirrorMode, address)
}
