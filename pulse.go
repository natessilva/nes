package nes

var dutyTable = [4][8]byte{
	{0, 1, 0, 0, 0, 0, 0, 0},
	{0, 1, 1, 0, 0, 0, 0, 0},
	{0, 1, 1, 1, 1, 0, 0, 0},
	{1, 0, 0, 1, 1, 1, 1, 1},
}

type pulse struct {
	enabled bool
	channel byte // 1 or 2

	duty      byte
	dutyIndex byte

	timerPeriod uint16
	timerValue  uint16

	lengthHalt    bool
	lengthCounter byte

	envelopeStart    bool
	envelopeLoop     bool
	envelopeConstant bool
	envelopeVolume   byte
	envelopeDecay    byte
	envelopeDivider  byte

	sweepEnabled bool
	sweepPeriod  byte
	sweepNegate  bool
	sweepShift   byte
	sweepDivider byte
	sweepReload  bool
}

func (p *pulse) stepTimer() {
	if p.timerValue == 0 {
		p.timerValue = p.timerPeriod
		p.dutyIndex = (p.dutyIndex + 1) % 8
	} else {
		p.timerValue--
	}
}

func (p *pulse) stepEnvelope() {
	if p.envelopeStart {
		p.envelopeStart = false
		p.envelopeDecay = 15
		p.envelopeDivider = p.envelopeVolume
	} else {
		if p.envelopeDivider == 0 {
			p.envelopeDivider = p.envelopeVolume
			if p.envelopeDecay > 0 {
				p.envelopeDecay--
			} else if p.envelopeLoop {
				p.envelopeDecay = 15
			}
		} else {
			p.envelopeDivider--
		}
	}
}

func (p *pulse) stepLength() {
	if !p.lengthHalt && p.lengthCounter > 0 {
		p.lengthCounter--
	}
}

func (p *pulse) sweepTarget() uint16 {
	change := p.timerPeriod >> p.sweepShift
	if p.sweepNegate {
		if p.channel == 1 {
			return p.timerPeriod - change - 1
		}
		return p.timerPeriod - change
	}
	return p.timerPeriod + change
}

func (p *pulse) stepSweep() {
	if p.sweepDivider == 0 && p.sweepEnabled && p.sweepShift > 0 && p.timerPeriod >= 8 && p.sweepTarget() <= 0x7FF {
		p.timerPeriod = p.sweepTarget()
	}
	if p.sweepDivider == 0 || p.sweepReload {
		p.sweepDivider = p.sweepPeriod
		p.sweepReload = false
	} else {
		p.sweepDivider--
	}
}

func (p *pulse) output() byte {
	if !p.enabled {
		return 0
	}
	if dutyTable[p.duty][p.dutyIndex] == 0 {
		return 0
	}
	if p.lengthCounter == 0 {
		return 0
	}
	if p.timerPeriod < 8 {
		return 0
	}
	if p.sweepTarget() > 0x7FF {
		return 0
	}
	if p.envelopeConstant {
		return p.envelopeVolume
	}
	return p.envelopeDecay
}
