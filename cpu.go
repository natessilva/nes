package nes

import (
	"log"
)

const (
	CPU_FLAG_C byte = 1 << iota
	CPU_FLAG_Z
	CPU_FLAG_I
	CPU_FLAG_D
	CPU_FLAG_B
	CPU_FLAG_UNUSED
	CPU_FLAG_V
	CPU_FLAG_N
)

type CPU struct {
	PC uint16 // 16 bit program counter
	SP byte   // 8 bit stack pointer
	A  byte   // 8 bit Accumulator
	X  byte   // 8 bit register
	Y  byte   // 8 bit register

	// Status bits NV_BDIZC
	Status byte

	RAM [2048]byte

	Cart *Cart

	PPU *PPU
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
	if address < 0x4000 {
		val := c.PPU.ReadRegister(address)
		return val
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

func (c *CPU) Write(address uint16, value byte) {
	switch {
	case address < 0x2000:
		c.RAM[address%0x0800] = value
	case address < 0x4000:
		c.PPU.WriteRegister(address, value)
	default:
		log.Fatalf("invalid write address %04x\n", address)
	}
}

// Set the zero flag
func (c *CPU) SetZ(value byte) {
	if value == 0 {
		c.Status = setBits(c.Status, CPU_FLAG_Z)
	} else {
		c.Status = resetBits(c.Status, CPU_FLAG_Z)
	}
}

// Set the negative flag
func (c *CPU) SetN(value byte) {
	if int8(value) < 0 {
		c.Status = setBits(c.Status, CPU_FLAG_N)
	} else {
		c.Status = resetBits(c.Status, CPU_FLAG_N)
	}
}

// Step looks up the opcode at the PC, executes
// the instruction, and steps the PC forward.
// Returns the number of CPU cycles the opcode
// took
func (c *CPU) Step() int {

	opcode := c.ReadByte(c.PC)
	c.PC += 1

	switch opcode {
	case 0x78:
		c.SEI()
		return 2
	case 0xD8:
		c.CLD()
		return 2
	case 0xA2:
		c.LDX(c.ImmediateMode())
		return 2
	case 0x9A:
		c.TXS()
		return 2
	case 0xAD:
		c.LDA(c.AbsoluteMode())
		return 4
	case 0x10:
		return c.BPL(c.RelativeMode())
	case 0xA9:
		c.LDA(uint16(c.ImmediateMode()))
		return 2
	case 0x8D:
		c.STA(c.AbsoluteMode())
		return 4
	case 0x8E:
		c.STX(c.AbsoluteMode())
		return 4
	case 0xA0:
		c.LDY(c.ImmediateMode())
		return 2
	default:
		log.Fatalf("unknown opcode %02x\n", opcode)
		return 0
	}
}

// Addressing modes

// ImmediateMode reads the byte after the opcode
// and increments PC past it
func (c *CPU) ImmediateMode() byte {
	b := c.ReadByte(c.PC)
	c.PC++
	return b
}

// AbsoluteMode reads the word after the opcode
// and increments PC past it
func (c *CPU) AbsoluteMode() uint16 {
	b := c.ReadWord(c.PC)
	c.PC += 2
	return b
}

// RelativeMode reads the offset after the opcode
// and adds it to the PC
// Branch offsets are signed 8-bit values, -128 ... +127
// negative offsets in two's complement.
func (c *CPU) RelativeMode() uint16 {

	offset := uint16(c.ReadByte(c.PC))
	c.PC += 1

	return offset
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

// The page is the high byte. If the high bytes
// differ then we've crossed a page boundary
func pageCrossed(address1, address2 uint16) bool {
	return address1&0xFF00 != address2&0xFF00
}

// Branch on N=0
func (c *CPU) BPL(offset uint16) int {
	if !isAnySet(c.Status, CPU_FLAG_N) {
		pc := c.PC
		c.PC += offset
		// negative offset
		if offset >= 0x80 {
			c.PC -= 0x100
		}
		// Add a cycle for branching
		// Add another for crossing pages
		if pageCrossed(pc, c.PC) {
			return 4
		}
		return 3
	}
	return 2
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
	c.Status = resetBits(c.Status, CPU_FLAG_D)
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

// Load A with the value at address
// sets N and Z flags
func (c *CPU) LDA(address uint16) {
	c.A = c.ReadByte(address)
	c.SetN(c.A)
	c.SetZ(c.A)
}

// Load X with value
// sets N and Z flags
func (c *CPU) LDX(value byte) {
	c.X = value
	c.SetN(c.X)
	c.SetZ(c.X)
}

// Load X with value
// sets N and Z flags
func (c *CPU) LDY(value byte) {
	c.Y = value
	c.SetN(c.Y)
	c.SetZ(c.Y)
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
	c.Status = setBits(c.Status, CPU_FLAG_I)
}

// Store accumulator in memory
func (c *CPU) STA(address uint16) {
	c.Write(address, c.A)
}

// Store X in memoory
func (c *CPU) STX(address uint16) {
	c.Write(address, c.X)
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

// Transfer X to Stack Pointer
func (c *CPU) TXS() {
	c.SP = c.X
}

func (c *CPU) TYA() {

	log.Fatal("Not implemented yet")
}
