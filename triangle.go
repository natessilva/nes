package nes

var triangleSequence = [32]byte{
	15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
}

type triangle struct {
	enabled bool

	timerPeriod uint16
	timerValue  uint16

	sequenceIndex byte

	lengthHalt    bool
	lengthCounter byte

	linearReload     byte
	linearCounter    byte
	linearReloadFlag bool
	controlFlag      bool
}

func (t *triangle) stepTimer() {
	if t.timerValue == 0 {
		t.timerValue = t.timerPeriod
		if t.linearCounter > 0 && t.lengthCounter > 0 {
			t.sequenceIndex = (t.sequenceIndex + 1) % 32
		}
	} else {
		t.timerValue--
	}
}

func (t *triangle) stepLinearCounter() {
	if t.linearReloadFlag {
		t.linearCounter = t.linearReload
	} else if t.linearCounter > 0 {
		t.linearCounter--
	}
	if !t.controlFlag {
		t.linearReloadFlag = false
	}
}

func (t *triangle) stepLength() {
	if !t.lengthHalt && t.lengthCounter > 0 {
		t.lengthCounter--
	}
}

func (t *triangle) output() byte {
	if !t.enabled {
		return 0
	}
	return triangleSequence[t.sequenceIndex]
}
