package nes

type NES struct {
	ppu *PPU
	cpu *CPU
}

func NewNES(cpu *CPU, ppu *PPU) *NES {
	return &NES{
		cpu: cpu,
		ppu: ppu,
	}
}

func (n *NES) StepFrame() {
	for {
		cycles := n.cpu.Step()
		cycles *= 3
		beforeNMI := n.ppu.NMITriggered()
		for ; cycles > 0; cycles-- {
			n.ppu.Step()
		}
		afterNMI := n.ppu.NMITriggered()
		if !beforeNMI && afterNMI {
			n.cpu.TriggerNMI()
			break
		}
	}
}
