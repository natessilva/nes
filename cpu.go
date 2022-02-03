package nes

import (
	"fmt"
	"log"
)

const (
	FLAG_C byte = 1 << iota
	FLAG_Z
	FLAG_I
	FLAG_D
	FLAG_B
	FLAG_UNUSED
	FLAG_V
	FLAG_N
)

func setBits(value, flags byte) byte {
	return value | flags
}

func resetBits(value, flags byte) byte {
	return value & mask(flags)
}

func mask(flag byte) byte {
	return ^flag
}

type CPU struct {
	PC  uint16 // 16 bit program counter
	SP  byte   // 8 bit stack pointer
	Acc byte   // 8 bit accumulator
	X   byte   // 8 bit register
	Y   byte   // 8 bit register

	// Status bits NV_BDIZC
	Status byte

	RAM [2048]byte

	Cart *Cart
}

func NewCPU(cart *Cart) *CPU {
	cpu := &CPU{
		Cart: cart,
	}
	cpu.Reset()
	return cpu
}

func (c *CPU) Reset() {
	// Program counter always starts at 0xFFFC
	c.PC = c.ReadWord(0xFFFC)
}

// ReadByte reads a single byte from either
// RAM, ROM, or IO.
func (c *CPU) ReadByte(address uint16) byte {
	// There is only 2KB of memory, but
	// it is mirrored 4 times over the
	// first 8KB
	if address < 0x2000 {
		return c.RAM[address%0x800]
	}
	// TODO eventually implement memory mapper
	if address >= 0x8000 {
		return c.Cart.ReadByte(address)
	}

	// TODO implement ROM and IO
	return 0
}

// ReadWord read two bytes from the given
// address.
func (c *CPU) ReadWord(address uint16) uint16 {
	b1 := uint16(c.ReadByte(address))
	b2 := uint16(c.ReadByte(address + 1))

	// The CPU is little endian so least significant byte
	// first.
	return b2<<8 | b1
}

// Step looks up the opcode at the PC, executes
// the instructionm, and steps the PC forward
func (c *CPU) Step() {

	opcode := c.ReadByte(c.PC)
	fmt.Printf("opcode: %02x\n", opcode)
	c.PC += 1

	switch opcode {
	case 0x78:
		c.SEI()
	case 0xD8:
		c.CLD()
	case 0xA2:
		c.LDX(c.ImmediateMode())
	default:
		fmt.Println("unknown opcode")
	}
	fmt.Printf("status bits before: %08b\n", c.Status)
	fmt.Printf("X: %02x\n", c.X)
}

// Addressing modes

// ImmediateMode simply reads the value after the opcode
// and increments PC again
func (c *CPU) ImmediateMode() byte {
	b := c.ReadByte(c.PC)
	c.PC++
	return b
}

func (c *CPU) ADC() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) AND() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) ASL() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BCC() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BCS() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BEQ() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BIT() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BMI() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BNE() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BPL() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BRK() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BVC() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) BVS() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) CLC() {
	log.Fatal("Not implemented yet")
}

// Clear decimal flag. Not sure if I need this
// because the NES doesn't support decimal
// mode anyway
func (c *CPU) CLD() {
	c.Status = resetBits(c.Status, FLAG_D)
}

func (c *CPU) CLI() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) CLV() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) CMP() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) CPX() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) CPY() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) DEC() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) DEX() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) DEY() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) EOR() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) INC() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) INX() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) INY() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) JMP() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) JSR() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) LDA() {
	log.Fatal("Not implemented yet")
}

// Load X with memory address
// sets N and Z flags
func (c *CPU) LDX(value byte) {
	c.X = value
	c.Status = setBits(c.Status, FLAG_N|FLAG_Z)
}

func (c *CPU) LDY() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) LSR() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) NOP() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) ORA() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) PHA() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) PHP() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) PLA() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) PLP() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) ROL() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) ROR() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) RTI() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) RTS() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) SBC() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) SEC() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) SED() {
	log.Fatal("Not implemented yet")
}

// Set interrupt disable flag
func (c *CPU) SEI() {
	c.Status = setBits(c.Status, FLAG_I)
}

func (c *CPU) STA() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) STX() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) STY() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) TAX() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) TAY() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) TSX() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) TXA() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) TXS() {
	log.Fatal("Not implemented yet")
}

func (c *CPU) TYA() {

	log.Fatal("Not implemented yet")
}
