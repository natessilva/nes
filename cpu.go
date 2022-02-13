package nes

import (
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
	cycles uint64 // total cycle counter
	pc     uint16 // 16 bit program counter
	sp     byte   // 8 bit stack pointer
	a      byte   // 8 bit Accumulator
	x      byte   // 8 bit register
	y      byte   // 8 bit register

	// Status bits NV_BDIZC
	status byte

	ram [2048]byte

	cart *Cart
	ppu  *PPU

	nmiTriggered bool
}

func NewCPU(cart *Cart, ppu *PPU) *CPU {
	cpu := &CPU{
		cart: cart,
		ppu:  ppu,
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

func (c *CPU) nmi() int {
	c.pushWord(c.pc)
	c.php(0)
	c.pc = c.readWord(0xFFFA)
	c.status = setBits(c.status, cpuFlagI)
	c.nmiTriggered = false
	return 7
}

func (c *CPU) TriggerNMI() {
	c.nmiTriggered = true
}

// readByte reads a byte from the memory map
func (c *CPU) readByte(address uint16) byte {
	switch {
	case address < 0x2000:
		return c.ram[address%0x800]
	case address < 0x4000:
		return c.ppu.readRegister((address - 0x4000) % 8)
	case address < 0x4020:
		// todo APU
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

// reads a word by in the same page
// if the high byte would be on another page
// wrap aroudn to the beginning of the page
func (c *CPU) readWordPageWrap(address uint16) uint16 {
	low := uint16(c.readByte(address))
	highAddress := (address & 0xFF00) | uint16(byte(address+1))
	high := uint16(c.readByte(highAddress))
	return (high << 8) | low
}

func (c *CPU) write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		c.ram[address%0x800] = value
	case address < 0x4000:
		c.ppu.writeRegister((address-0x4000)%8, value)
	case address < 0x4020:
		// TODO implement APU
	default:
		log.Fatalf("invalid write address %04X", address)
	}
}

// Step steps the CPU forward one instruction returning
// the number of cyles it took.
func (c *CPU) Step() int {
	if c.nmiTriggered {
		return c.nmi()
	}
	opcode := c.readByte(c.pc)
	var inst func(address uint16)
	var mode int
	cycles := c.cycles
	var bytes uint16
	var address uint16
	noPageCrossed := false
	switch opcode {
	case 0x4C:
		inst = c.jmp
		mode = modeAbsolute
		bytes = 3
		c.cycles += 3
	case 0xA2:
		inst = c.ldx
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0x86:
		inst = c.stx
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x20:
		inst = c.jsr
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
	case 0xEA:
		inst = c.nop
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x38:
		inst = c.sec
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xB0:
		inst = c.bcs
		mode = modeRelative
		bytes = 2
		c.cycles += 2
	case 0x18:
		inst = c.clc
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x90:
		inst = c.bcc
		mode = modeRelative
		bytes = 2
		c.cycles += 2
	case 0xA9:
		inst = c.lda
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xF0:
		inst = c.beq
		mode = modeRelative
		bytes = 2
		c.cycles += 2
	case 0xD0:
		inst = c.bne
		mode = modeRelative
		bytes = 2
		c.cycles += 2
	case 0x85:
		inst = c.sta
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x24:
		inst = c.bit
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x70:
		inst = c.bvs
		mode = modeRelative
		bytes = 2
		c.cycles += 2
	case 0x50:
		inst = c.bvc
		mode = modeRelative
		bytes = 2
		c.cycles += 2
	case 0x10:
		inst = c.bpl
		mode = modeRelative
		bytes = 2
		c.cycles += 2
	case 0x60:
		inst = c.rts
		mode = modeImplied
		bytes = 1
		c.cycles += 6
	case 0x78:
		inst = c.sei
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xF8:
		inst = c.sed
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x08:
		inst = c.php
		mode = modeImplied
		bytes = 1
		c.cycles += 3
	case 0x68:
		inst = c.pla
		mode = modeImplied
		bytes = 1
		c.cycles += 4
	case 0x29:
		inst = c.and
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xC9:
		inst = c.cmp
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xD8:
		inst = c.cld
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x48:
		inst = c.pha
		mode = modeImplied
		bytes = 1
		c.cycles += 3
	case 0x28:
		inst = c.plp
		mode = modeImplied
		bytes = 1
		c.cycles += 4
	case 0x30:
		inst = c.bmi
		mode = modeRelative
		bytes = 2
		c.cycles += 2
	case 0x09:
		inst = c.ora
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xB8:
		inst = c.clv
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x49:
		inst = c.eor
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0x69:
		inst = c.adc
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xA0:
		inst = c.ldy
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xC0:
		inst = c.cpy
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xE0:
		inst = c.cpx
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xE9:
		inst = c.sbc
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xC8:
		inst = c.iny
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xE8:
		inst = c.inx
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x88:
		inst = c.dey
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xCA:
		inst = c.dex
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xA8:
		inst = c.tay
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xAA:
		inst = c.tax
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x98:
		inst = c.tya
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x8A:
		inst = c.txa
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xBA:
		inst = c.tsx
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x8E:
		inst = c.stx
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x9A:
		inst = c.txs
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xAE:
		inst = c.ldx
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0xAD:
		inst = c.lda
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x40:
		inst = c.rti
		mode = modeImplied
		bytes = 1
		c.cycles += 6
	case 0x4A:
		inst = c.lsra
		mode = modeAccumulator
		bytes = 1
		c.cycles += 2
	case 0x0A:
		inst = c.asla
		mode = modeAccumulator
		bytes = 1
		c.cycles += 2
	case 0x6A:
		inst = c.rora
		mode = modeAccumulator
		bytes = 1
		c.cycles += 2
	case 0x2A:
		inst = c.rola
		mode = modeAccumulator
		bytes = 1
		c.cycles += 2
	case 0xA5:
		inst = c.lda
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x8D:
		inst = c.sta
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0xA1:
		inst = c.lda
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
	case 0x81:
		inst = c.sta
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
	case 0x01:
		inst = c.ora
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
	case 0x21:
		inst = c.and
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
	case 0x41:
		inst = c.eor
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
	case 0x61:
		inst = c.adc
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
	case 0xC1:
		inst = c.cmp
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
	case 0xE1:
		inst = c.sbc
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
	case 0xA4:
		inst = c.ldy
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x84:
		inst = c.sty
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0xA6:
		inst = c.ldx
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x05:
		inst = c.ora
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x25:
		inst = c.and
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x45:
		inst = c.eor
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x65:
		inst = c.adc
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0xC5:
		inst = c.cmp
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0xE5:
		inst = c.sbc
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0xE4:
		inst = c.cpx
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x46:
		inst = c.lsr
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
	case 0xC4:
		inst = c.cpy
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x06:
		inst = c.asl
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
	case 0x66:
		inst = c.ror
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
	case 0x26:
		inst = c.rol
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
	case 0xE6:
		inst = c.inc
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
	case 0xC6:
		inst = c.dec
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
	case 0xAC:
		inst = c.ldy
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x8C:
		inst = c.sty
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x2C:
		inst = c.bit
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x0D:
		inst = c.ora
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x2D:
		inst = c.and
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x4D:
		inst = c.eor
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x6D:
		inst = c.adc
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0xCD:
		inst = c.cmp
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0xED:
		inst = c.sbc
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0xEC:
		inst = c.cpx
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0xCC:
		inst = c.cpy
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x4E:
		inst = c.lsr
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
	case 0x0E:
		inst = c.asl
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
	case 0x6E:
		inst = c.ror
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
	case 0x2E:
		inst = c.rol
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
	case 0xEE:
		inst = c.inc
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
	case 0xCE:
		inst = c.dec
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
	case 0xB1:
		inst = c.lda
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 5
	case 0x11:
		inst = c.ora
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 5
	case 0x31:
		inst = c.and
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 5
	case 0x51:
		inst = c.eor
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 5
	case 0x71:
		inst = c.adc
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 5
	case 0xD1:
		inst = c.cmp
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 5
	case 0xF1:
		inst = c.sbc
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 5
	case 0x91:
		inst = c.sta
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0x6C:
		inst = c.jmp
		mode = modeIndirect
		bytes = 3
		c.cycles += 5
	case 0xB9:
		inst = c.lda
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0x19:
		inst = c.ora
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0x39:
		inst = c.and
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0x59:
		inst = c.eor
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0x79:
		inst = c.adc
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0xD9:
		inst = c.cmp
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0xF9:
		inst = c.sbc
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0x99:
		inst = c.sta
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 5
		noPageCrossed = true
	case 0xB4:
		inst = c.ldy
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x94:
		inst = c.sty
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x15:
		inst = c.ora
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x35:
		inst = c.and
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x55:
		inst = c.eor
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x75:
		inst = c.adc
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0xD5:
		inst = c.cmp
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0xF5:
		inst = c.sbc
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0xB5:
		inst = c.lda
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x95:
		inst = c.sta
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x56:
		inst = c.lsr
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
	case 0x16:
		inst = c.asl
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
	case 0x76:
		inst = c.ror
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
	case 0x36:
		inst = c.rol
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
	case 0xF6:
		inst = c.inc
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
	case 0xD6:
		inst = c.dec
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
	case 0xB6:
		inst = c.ldx
		mode = modeZeroPageY
		bytes = 2
		c.cycles += 4
	case 0x96:
		inst = c.stx
		mode = modeZeroPageY
		bytes = 2
		c.cycles += 4
	case 0xBC:
		inst = c.ldy
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0x1D:
		inst = c.ora
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0x3D:
		inst = c.and
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0x5D:
		inst = c.eor
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0x7D:
		inst = c.adc
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0xDD:
		inst = c.cmp
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0xFD:
		inst = c.sbc
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0xBD:
		inst = c.lda
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0x9D:
		inst = c.sta
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 5
		noPageCrossed = true
	case 0x5E:
		inst = c.lsr
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0x1E:
		inst = c.asl
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0x7E:
		inst = c.ror
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0x3E:
		inst = c.rol
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0xFE:
		inst = c.inc
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0xDE:
		inst = c.dec
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0xBE:
		inst = c.ldx
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0x04:
		inst = c.nop
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x44:
		inst = c.nop
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x64:
		inst = c.nop
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x0C:
		inst = c.nop
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x14:
		inst = c.nop
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x34:
		inst = c.nop
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x54:
		inst = c.nop
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x74:
		inst = c.nop
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0xD4:
		inst = c.nop
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0xF4:
		inst = c.nop
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 4
	case 0x1A:
		inst = c.nop
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x3A:
		inst = c.nop
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x5A:
		inst = c.nop
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x7A:
		inst = c.nop
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xDA:
		inst = c.nop
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0xFA:
		inst = c.nop
		mode = modeImplied
		bytes = 1
		c.cycles += 2
	case 0x80:
		inst = c.nop
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0x1C:
		inst = c.nop
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0x3C:
		inst = c.nop
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0x5C:
		inst = c.nop
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0x7C:
		inst = c.nop
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0xDC:
		inst = c.nop
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0xFC:
		inst = c.nop
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 4
	case 0xA3:
		inst = c.lax
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0xA7:
		inst = c.lax
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0xAF:
		inst = c.lax
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0xB3:
		inst = c.lax
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 5
	case 0xB7:
		inst = c.lax
		mode = modeZeroPageY
		bytes = 2
		c.cycles += 4
	case 0xBF:
		inst = c.lax
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 4
	case 0x83:
		inst = c.sax
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0x87:
		inst = c.sax
		mode = modeZeroPage
		bytes = 2
		c.cycles += 3
	case 0x8F:
		inst = c.sax
		mode = modeAbsolute
		bytes = 3
		c.cycles += 4
	case 0x97:
		inst = c.sax
		mode = modeZeroPageY
		bytes = 2
		c.cycles += 4
	case 0xEB:
		inst = c.sbc
		mode = modeImmediate
		bytes = 2
		c.cycles += 2
	case 0xC3:
		inst = c.dcp
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0xC7:
		inst = c.dcp
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
		noPageCrossed = true
	case 0xCF:
		inst = c.dcp
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0xD3:
		inst = c.dcp
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0xD7:
		inst = c.dcp
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0xDB:
		inst = c.dcp
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0xDF:
		inst = c.dcp
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0xE3:
		inst = c.isc
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0xE7:
		inst = c.isc
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
		noPageCrossed = true
	case 0xEF:
		inst = c.isc
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0xF3:
		inst = c.isc
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0xF7:
		inst = c.isc
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0xFB:
		inst = c.isc
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0xFF:
		inst = c.isc
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0x03:
		inst = c.slo
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0x07:
		inst = c.slo
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
		noPageCrossed = true
	case 0x0F:
		inst = c.slo
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0x13:
		inst = c.slo
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0x17:
		inst = c.slo
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0x1B:
		inst = c.slo
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0x1F:
		inst = c.slo
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0x23:
		inst = c.rla
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0x27:
		inst = c.rla
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
		noPageCrossed = true
	case 0x2F:
		inst = c.rla
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0x33:
		inst = c.rla
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0x37:
		inst = c.rla
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0x3B:
		inst = c.rla
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0x3F:
		inst = c.rla
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0x43:
		inst = c.sre
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0x47:
		inst = c.sre
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
		noPageCrossed = true
	case 0x4F:
		inst = c.sre
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0x53:
		inst = c.sre
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0x57:
		inst = c.sre
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0x5B:
		inst = c.sre
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0x5F:
		inst = c.sre
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 7
		noPageCrossed = true
	case 0x63:
		inst = c.rra
		mode = modeIndexedIndirect
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0x67:
		inst = c.rra
		mode = modeZeroPage
		bytes = 2
		c.cycles += 5
		noPageCrossed = true
	case 0x6F:
		inst = c.rra
		mode = modeAbsolute
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0x73:
		inst = c.rra
		mode = modeIndirectIndexed
		bytes = 2
		c.cycles += 8
		noPageCrossed = true
	case 0x77:
		inst = c.rra
		mode = modeZeroPageX
		bytes = 2
		c.cycles += 6
		noPageCrossed = true
	case 0x7B:
		inst = c.rra
		mode = modeAbsoluteY
		bytes = 3
		c.cycles += 6
		noPageCrossed = true
	case 0x7F:
		inst = c.rra
		mode = modeAbsoluteX
		bytes = 3
		c.cycles += 6
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
	case modeRelative:
		// address is a relative offset signed byte
		offset := uint16(c.readByte(c.pc + 1))
		if offset < 0x80 {
			address = c.pc + 2 + offset
		} else {
			address = c.pc + 2 + offset - 0x100
		}
	case modeIndexedIndirect:
		address = c.readWordPageWrap(uint16(c.readByte(c.pc+1) + c.x))
	case modeIndirectIndexed:
		offset := c.readWordPageWrap(uint16(c.readByte(c.pc + 1)))
		address = offset + uint16(c.y)
		if pageCrossed(offset, address) && !noPageCrossed {
			c.cycles++
		}
	case modeIndirect:
		address = c.readWordPageWrap(c.readWord(c.pc + 1))
	case modeAbsoluteY:
		offset := c.readWord(c.pc + 1)
		address = offset + uint16(c.y)
		if pageCrossed(offset, address) {
			c.cycles++
		}
	case modeAbsoluteX:
		offset := c.readWord(c.pc + 1)
		address = offset + uint16(c.x)
		if pageCrossed(offset, address) && !noPageCrossed {
			c.cycles++
		}
	case modeZeroPageX:
		address = uint16(c.readByte(c.pc+1)+c.x) & 0xff
	case modeZeroPageY:
		address = uint16(c.readByte(c.pc+1)+c.y) & 0xff
	}
	c.pc += bytes
	inst(address)
	return int(c.cycles - cycles)
}

// a page is 256 bytes. The high byte is the page
// the low byte is the index within the page
// check if the pages differ
func pageCrossed(a, b uint16) bool {
	return a>>8 != b>>8
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

func (c *CPU) push(value byte) {
	c.write(0x100|uint16(c.sp), value)
	c.sp--
}

func (c *CPU) pushWord(value uint16) {
	high := byte(value >> 8)
	low := byte(value & 0xFF)
	c.push(high)
	c.push(low)
}

func (c *CPU) pull() byte {
	c.sp++
	return c.readByte(0x100 | uint16(c.sp))
}

func (c *CPU) pullWord() uint16 {
	low := uint16(c.pull())
	high := uint16(c.pull())
	return high<<8 | low
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

func (c *CPU) sty(address uint16) {
	c.write(address, c.y)
}

func (c *CPU) jsr(address uint16) {
	c.pushWord(c.pc - 1)
	c.pc = address
}

func (c *CPU) nop(address uint16) {
}

func (c *CPU) sec(address uint16) {
	c.status = setBits(c.status, cpuFlagC)
}

func (c *CPU) addBranchCycles(address uint16) {
	crossed := pageCrossed(c.pc, address)
	c.cycles++
	if crossed {
		c.cycles++
	}
}

func (c *CPU) bcs(address uint16) {
	if isAnySet(c.status, cpuFlagC) {
		c.addBranchCycles(address)
		c.pc = address
	}
}

func (c *CPU) bcc(address uint16) {
	if !isAnySet(c.status, cpuFlagC) {
		c.addBranchCycles(address)
		c.pc = address
	}
}

func (c *CPU) beq(address uint16) {
	if isAnySet(c.status, cpuFlagZ) {
		c.addBranchCycles(address)
		c.pc = address
	}
}

func (c *CPU) bne(address uint16) {
	if !isAnySet(c.status, cpuFlagZ) {
		c.addBranchCycles(address)
		c.pc = address
	}
}

func (c *CPU) bvs(address uint16) {
	if isAnySet(c.status, cpuFlagV) {
		c.addBranchCycles(address)
		c.pc = address
	}
}

func (c *CPU) bvc(address uint16) {
	if !isAnySet(c.status, cpuFlagV) {
		c.addBranchCycles(address)
		c.pc = address
	}
}

func (c *CPU) bpl(address uint16) {
	if !isAnySet(c.status, cpuFlagN) {
		c.addBranchCycles(address)
		c.pc = address
	}
}

func (c *CPU) bmi(address uint16) {
	if isAnySet(c.status, cpuFlagN) {
		c.addBranchCycles(address)
		c.pc = address
	}
}

func (c *CPU) clc(address uint16) {
	c.status = resetBits(c.status, cpuFlagC)
}

func (c *CPU) lda(address uint16) {
	c.a = c.readByte(address)
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) ldy(address uint16) {
	c.y = c.readByte(address)
	c.setZ(c.y)
	c.setN(c.y)
}

func (c *CPU) sta(address uint16) {
	c.write(address, c.a)
}

func (c *CPU) bit(address uint16) {
	value := c.readByte(address)
	if (value>>6)&1 == 1 {
		c.status = setBits(c.status, cpuFlagV)
	} else {
		c.status = resetBits(c.status, cpuFlagV)
	}
	c.setZ(value & c.a)
	c.setN(value)
}

func (c *CPU) rts(address uint16) {
	c.pc = c.pullWord() + 1
}

func (c *CPU) sei(address uint16) {
	c.status = setBits(c.status, cpuFlagI)
}

func (c *CPU) sed(address uint16) {
	c.status = setBits(c.status, cpuFlagD)
}

func (c *CPU) cld(address uint16) {
	c.status = resetBits(c.status, cpuFlagD)
}

func (c *CPU) php(address uint16) {
	// bit 5 is always set on push
	c.push(c.status | 0x10)
}

func (c *CPU) pha(address uint16) {
	c.push(c.a)
}

func (c *CPU) pla(address uint16) {
	c.a = c.pull()
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) plp(address uint16) {
	// ignore bit 5
	c.status = c.pull()&0xEF | 0x20
}

func (c *CPU) and(address uint16) {
	c.a &= c.readByte(address)
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) compare(a, b byte) {
	c.setZ(a - b)
	c.setN(a - b)
	if a >= b {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}

}

func (c *CPU) cmp(address uint16) {
	c.compare(c.a, c.readByte(address))
}

func (c *CPU) cpy(address uint16) {
	c.compare(c.y, c.readByte(address))
}

func (c *CPU) cpx(address uint16) {
	c.compare(c.x, c.readByte(address))
}

func (c *CPU) ora(address uint16) {
	c.a |= c.readByte(address)
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) clv(address uint16) {
	c.status = resetBits(c.status, cpuFlagV)
}

func (c *CPU) eor(address uint16) {
	c.a ^= c.readByte(address)
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) adc(address uint16) {
	a := c.a
	b := c.readByte(address)
	carry := c.status & 1
	c.a = a + b + carry
	c.setZ(c.a)
	c.setN(c.a)
	// if overflow set the carry bit
	if int(a)+int(b)+int(carry) > 0xFF {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}

	// if both positive and result negative
	// or both negative and result positive
	// then we've overflowed
	if (a^b)&0x80 == 0 && (a^c.a)&0x80 != 0 {
		c.status = setBits(c.status, cpuFlagV)
	} else {
		c.status = resetBits(c.status, cpuFlagV)
	}
}

func (c *CPU) sbc(address uint16) {
	a := c.a
	b := c.readByte(address)
	carry := c.status & 1
	c.a = a - b - (1 - carry)
	c.setZ(c.a)
	c.setN(c.a)
	if int(a)-int(b)-int(1-carry) >= 0 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}

	if (a^b)&0x80 != 0 && (a^c.a)&0x80 != 0 {
		c.status = setBits(c.status, cpuFlagV)
	} else {
		c.status = resetBits(c.status, cpuFlagV)
	}
}

func (c *CPU) iny(address uint16) {
	c.y++
	c.setZ(c.y)
	c.setN(c.y)
}

func (c *CPU) inx(address uint16) {
	c.x++
	c.setZ(c.x)
	c.setN(c.x)
}

func (c *CPU) dey(address uint16) {
	c.y--
	c.setZ(c.y)
	c.setN(c.y)
}

func (c *CPU) dex(address uint16) {
	c.x--
	c.setZ(c.x)
	c.setN(c.x)
}

func (c *CPU) tay(address uint16) {
	c.y = c.a
	c.setZ(c.y)
	c.setN(c.y)
}

func (c *CPU) tax(address uint16) {
	c.x = c.a
	c.setZ(c.x)
	c.setN(c.x)
}

func (c *CPU) tya(address uint16) {
	c.a = c.y
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) txa(address uint16) {
	c.a = c.x
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) tsx(address uint16) {
	c.x = c.sp
	c.setZ(c.x)
	c.setN(c.x)
}

func (c *CPU) txs(address uint16) {
	c.sp = c.x
}

func (c *CPU) rti(address uint16) {
	c.status = c.pull()&0xEF | 0x20
	c.pc = c.pullWord()
}

func (c *CPU) lsra(address uint16) {
	if c.a&1 == 1 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}
	c.a >>= 1
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) rora(address uint16) {
	carry := c.status & 1
	if c.a&1 == 1 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}
	c.a = (c.a >> 1) | (carry << 7)
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) ror(address uint16) {
	value := c.readByte(address)
	carry := c.status & 1
	if value&1 == 1 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}
	value = (value >> 1) | (carry << 7)
	c.write(address, value)
	c.setZ(value)
	c.setN(value)
}

func (c *CPU) rola(address uint16) {
	carry := c.status & 1
	if (c.a>>7)&1 == 1 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}
	c.a = (c.a << 1) | carry
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) rol(address uint16) {
	value := c.readByte(address)
	carry := c.status & 1
	if (value>>7)&1 == 1 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}
	value = (value << 1) | carry
	c.write(address, value)
	c.setZ(value)
	c.setN(value)
}

func (c *CPU) lsr(address uint16) {
	value := c.readByte(address)
	if value&1 == 1 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}
	value >>= 1
	c.write(address, value)
	c.setZ(value)
	c.setN(value)
}

func (c *CPU) asla(address uint16) {
	if (c.a>>7)&1 == 1 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}
	c.a <<= 1
	c.setZ(c.a)
	c.setN(c.a)
}

func (c *CPU) asl(address uint16) {
	value := c.readByte(address)
	if (value>>7)&1 == 1 {
		c.status = setBits(c.status, cpuFlagC)
	} else {
		c.status = resetBits(c.status, cpuFlagC)
	}
	value <<= 1
	c.write(address, value)
	c.setZ(value)
	c.setN(value)
}

func (c *CPU) inc(address uint16) {
	value := c.readByte(address)
	value++
	c.setZ(value)
	c.setN(value)
	c.write(address, value)
}

func (c *CPU) dec(address uint16) {
	value := c.readByte(address)
	value--
	c.setZ(value)
	c.setN(value)
	c.write(address, value)
}

func (c *CPU) lax(address uint16) {
	c.lda(address)
	c.ldx(address)
}

func (c *CPU) sax(address uint16) {
	c.write(address, c.a&c.x)
}

func (c *CPU) dcp(address uint16) {
	c.dec(address)
	c.cmp(address)
}

func (c *CPU) isc(address uint16) {
	c.inc(address)
	c.sbc(address)
}

func (c *CPU) slo(address uint16) {
	c.asl(address)
	c.ora(address)
}

func (c *CPU) rla(address uint16) {
	c.rol(address)
	c.and(address)
}

func (c *CPU) sre(address uint16) {
	c.lsr(address)
	c.eor(address)
}

func (c *CPU) rra(address uint16) {
	c.ror(address)
	c.adc(address)
}
