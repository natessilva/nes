package nes

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
	index := (int(address) - 0x8000) % len(c.PRG)
	return c.PRG[index]
}
