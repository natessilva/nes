package nes

import "log"

const (
	H_MIRROR = iota
	V_MIRROR
)

type Cart struct {
	Mirror byte
	PRG    []byte
	CHR    []byte
}

func (c *Cart) ReadByte(address uint16) byte {
	switch {
	case address >= 0xC000:
		index := int(address-0xC000) % len(c.PRG)
		return c.PRG[index]
	case address >= 0x8000:
		index := int(address-0x8000) % len(c.PRG)
		return c.PRG[index]
	default:
		log.Fatalf("invalid address %04x\n", address)
	}
	return 0
}
