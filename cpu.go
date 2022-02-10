package nes

import (
	"fmt"
	"log"
)

const (
	cpuFlagC byte = 1 << iota
	cpuFlagZ
	cpuFlagI
	cpuFlagD
	cpuFlagB
	cpuFlagU
	cpuFlagV
	cpuFlagN
)

const (
	modeAccumulator = iota
	modeAbsolute
	modeAbsoluteX
	modeAbsoluteY
	modeImmediate
	modeImplied
	modeIndexedIndirect
	modeIndirect
	modeIndirectIndexed
	modeRelative
	modeZeroPage
	modeZeroPageX
	modeZeroPageY
)

type CPU struct {
	pc uint16 // 16 bit program counter
	sp byte   // 8 bit stack pointer
	a  byte   // 8 bit Accumulator
	x  byte   // 8 bit register
	y  byte   // 8 bit register

	// Status bits NV_BDIZC
	status byte

	ram [2048]byte

	cart *Cart
}

func NewCPU(cart *Cart) *CPU {
	cpu := &CPU{
		cart: cart,
	}
	cpu.reset()
	return cpu
}

func (c *CPU) reset() {
	// Program counter always starts at 0xFFFC
	c.pc = c.readWord(0xFFFC)
	c.sp = 0xFD
	c.status = 0x24
}

// readByte reads a byte from the memory map
func (c *CPU) readByte(address uint16) byte {
	switch {
	case address < 0x2000:
		return c.ram[address%0x800]
	case address >= 0x6000:
		return c.cart.ReadByte(address)
	default:
		log.Fatalf("invalid read address %02X", address)
	}
	return 0
}

// readWord reads a 16 bit word from the memory map
// low byte first.
func (c *CPU) readWord(address uint16) uint16 {
	low := uint16(c.readByte(address))
	high := uint16(c.readByte(address + 1))
	return (high << 8) | low
}

func (c *CPU) write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		fmt.Printf("%02X = %02X\n", address%0x800, value)
		c.ram[address%0x800] = value
	default:
		log.Fatalf("invalid write address %04X", address)
	}
}

// Step steps the CPU forward one instruction returning
// the number of cyles it took.
func (c *CPU) Step() int {
	opcode := c.readByte(c.pc)
	var inst func(address uint16)
	var mode int
	var cycles int
	var bytes uint16
	var address uint16
	switch opcode {
	case 0x4C:
		inst = c.jmp
		mode = modeAbsolute
		bytes = 3
		cycles = 3
	case 0xA2:
		inst = c.ldx
		mode = modeImmediate
		bytes = 2
		cycles = 2
	case 0x86:
		inst = c.stx
		mode = modeZeroPage
		bytes = 2
		cycles = 3
	default:
		log.Fatalf("unknown opcode %02X", opcode)
	}
	switch mode {
	case modeAbsolute:
		address = c.readWord(c.pc + 1)
	case modeImmediate:
		address = c.pc + 1
	case modeZeroPage:
		address = uint16(c.readByte(c.pc + 1))
	}
	c.pc += bytes
	inst(address)
	return cycles
}

func (c *CPU) setZ(value byte) {
	if value == 0 {
		c.status = setBits(c.status, cpuFlagZ)
	} else {
		c.status = resetBits(c.status, cpuFlagZ)
	}
}

func (c *CPU) setN(value byte) {
	if value >= 0x80 {
		c.status = setBits(c.status, cpuFlagN)
	} else {
		c.status = resetBits(c.status, cpuFlagN)
	}
}

func (c *CPU) jmp(address uint16) {
	c.pc = address
}

func (c *CPU) ldx(address uint16) {
	c.x = c.readByte(address)
	c.setZ(c.x)
	c.setN(c.x)
}

func (c *CPU) stx(address uint16) {
	c.write(address, c.x)
}
