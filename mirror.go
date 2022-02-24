package nes

const (
	mirrorHorizontal = iota
	mirrorVertical
	mirrorSingle0
	mirrorSingle1
)

func mirror(mode byte, address uint16) uint16 {
	// map to the vram address space
	address = (address - 0x2000) % 0x1000
	// which of the 4 name tables
	table := address / 0x400

	// todo implement more mirroring modes
	switch mode {
	case mirrorHorizontal:
		table /= 2
	case mirrorVertical:
		table %= 2
	case mirrorSingle0:
		table = 0
	case mirrorSingle1:
		table = 1
	}

	// where at within the table
	location := address % 0x400

	return table*0x400 + location
}
