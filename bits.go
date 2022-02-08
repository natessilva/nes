package nes

func setBits(value, flags byte) byte {
	return value | flags
}

func resetBits(value, flags byte) byte {
	return value & mask(flags)
}

func mask(flags byte) byte {
	return ^flags
}

func isAnySet(value, flags byte) bool {
	return value&flags > 0
}
