package nes

import "log"

type cartridge struct {
	mirror byte
	prg    []byte
	chr    []byte
}

func (c *cartridge) readByte(address uint16) byte {
	switch {
	case address < 0x2000:
		return c.chr[address]
	case address >= 0xC000:
		index := int(address-0xC000) % len(c.prg)
		return c.prg[index]
	case address >= 0x8000:
		index := int(address-0x8000) % len(c.prg)
		return c.prg[index]
	default:
		log.Fatalf("invalid address %04x\n", address)
	}
	return 0
}
