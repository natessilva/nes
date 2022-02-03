package nes

import (
	"fmt"
	"log"
)

const (
	C = iota
	Z
	I
	D
	B
	_
	V
	N
)

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
	fmt.Printf("status bits before: %08b\n", c.Status)

	opcode := c.ReadByte(c.PC)
	fmt.Printf("opcode: %x\n", opcode)
	c.PC += 1

	switch opcode {
	case 0x78:
		c.SEI()
	default:
		fmt.Println("unknown opcode")
	}

	fmt.Printf("status bits after: %08b\n", c.Status)
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

func (c *CPU) CLD() {
	log.Fatal("Not implemented yet")
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

func (c *CPU) LDX() {
	log.Fatal("Not implemented yet")
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
	c.Status |= (1 << I)
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
