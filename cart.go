package nes

// type cartridge struct {
// 	mirror byte
// 	prg    []byte
// 	chr    []byte
// }

// func (c *cartridge) readByte(address uint16) byte {
// 	switch {
// 	case address < 0x2000:
// 		return c.chr[address]
// 	case address >= 0xC000:
// 		index := int(address-0x8000) % len(c.prg)
// 		return c.prg[index]
// 	case address >= 0x8000:
// 		index := int(address-0x8000) % len(c.prg)
// 		return c.prg[index]
// 	default:
// 		log.Fatalf("invalid address %04x\n", address)
// 	}
// 	return 0
// }

import "log"

type cartridge interface {
	readByte(address uint16) byte
	write(address uint16, value byte)
	mirror(address uint16) uint16
}

func newCart(mapper, mirror byte, prg, chr []byte) cartridge {
	switch mapper {
	case 0:
		return &nROM{
			mirrorMode: mirror,
			prg:        prg,
			chr:        chr,
		}
	case 1:
		return newMMC1(mirror, prg, chr)
	case 2:
		return &unROM{
			mirrorMode: mirror,
			prg:        prg,
			chr:        chr,
		}
	case 3:
		return &cnROM{
			mirrorMode: mirror,
			prg:        prg,
			chr:        chr,
		}
	default:
		log.Fatalf("unsupported mapper %d", mapper)
	}
	return nil
}
