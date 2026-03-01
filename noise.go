package nes

var noisePeriodTable = [16]uint16{
	4, 8, 16, 32, 64, 96, 128, 160, 202, 254, 380, 508, 762, 1016, 2034, 4068,
}

type noise struct {
	enabled bool

	timerPeriod uint16
	timerValue  uint16

	mode     bool
	shiftReg uint16

	lengthHalt    bool
	lengthCounter byte

	envelopeStart    bool
	envelopeLoop     bool
	envelopeConstant bool
	envelopeVolume   byte
	envelopeDecay    byte
	envelopeDivider  byte
}

func newNoise() *noise {
	return &noise{shiftReg: 1}
}

func (n *noise) stepTimer() {
	if n.timerValue == 0 {
		n.timerValue = n.timerPeriod
		var bit2 uint16
		if n.mode {
			bit2 = (n.shiftReg >> 6) & 1
		} else {
			bit2 = (n.shiftReg >> 1) & 1
		}
		feedback := (n.shiftReg & 1) ^ bit2
		n.shiftReg >>= 1
		n.shiftReg |= feedback << 14
	} else {
		n.timerValue--
	}
}

func (n *noise) stepEnvelope() {
	if n.envelopeStart {
		n.envelopeStart = false
		n.envelopeDecay = 15
		n.envelopeDivider = n.envelopeVolume
	} else {
		if n.envelopeDivider == 0 {
			n.envelopeDivider = n.envelopeVolume
			if n.envelopeDecay > 0 {
				n.envelopeDecay--
			} else if n.envelopeLoop {
				n.envelopeDecay = 15
			}
		} else {
			n.envelopeDivider--
		}
	}
}

func (n *noise) stepLength() {
	if !n.lengthHalt && n.lengthCounter > 0 {
		n.lengthCounter--
	}
}

func (n *noise) output() byte {
	if !n.enabled {
		return 0
	}
	if n.shiftReg&1 == 1 {
		return 0
	}
	if n.lengthCounter == 0 {
		return 0
	}
	if n.envelopeConstant {
		return n.envelopeVolume
	}
	return n.envelopeDecay
}
