package nes

import (
	"image"
	"io"
)

type Console struct {
	ppu     *ppu
	cpu     *cpu
	apu     *apu
	joypad1 *joypad
}

func NewConsole(r io.Reader) (*Console, error) {
	c := &Console{}
	err := c.loadROM(r)
	return c, err
}

func (c *Console) loadROM(r io.Reader) error {
	cart, err := readFile(r)
	if err != nil {
		return err
	}
	c.ppu = newPPU(cart)
	c.apu = newAPU()
	c.joypad1 = &joypad{}
	c.cpu = newCPU(cart, c.ppu, c.apu, c.joypad1)
	return nil
}

func (c *Console) RenderFrame(image *image.RGBA) {
	for {
		cpuCycles := c.cpu.Step()
		for i := 0; i < cpuCycles; i++ {
			c.apu.step()
		}
		ppuCycles := cpuCycles * 3
		beforeNMI := c.ppu.nmiTriggered()
		for ; ppuCycles > 0; ppuCycles-- {
			c.ppu.step(image)
		}
		afterNMI := c.ppu.nmiTriggered()
		if !beforeNMI && afterNMI {
			c.cpu.triggerNMI()
			break
		}
	}
}

func (c *Console) AudioSamples() ([]byte, int) {
	return c.apu.flushSamples()
}

func (c *Console) SetJoypad(button byte, pressed bool) {
	if pressed {
		c.joypad1.buttonState = setBits(c.joypad1.buttonState, button)
	} else {
		c.joypad1.buttonState = resetBits(c.joypad1.buttonState, button)
	}
}
