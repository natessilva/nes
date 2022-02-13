package nes

const (
	ButtonA byte = 1 << iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
)

type joypad struct {
	// strobe enabled
	// if strobe is enabled, reads will
	// only return the status of the A
	// button.
	// if strong is disabled, reads will
	// cycle through all buttons and then
	// begin continually returning 1s
	// until the strobe is enabled again
	strobe bool

	// button currently being read
	buttonIndex byte

	// the state of all 8 buttons
	buttonState byte
}

// Read the state of a single button
func (j *joypad) read() byte {
	if j.buttonIndex > 7 {
		return 1
	}
	// if strong enabled, return the state of A
	if j.strobe {
		return j.buttonState & 1
	}
	value := j.buttonState & (1 << j.buttonIndex) >> j.buttonIndex
	j.buttonIndex++
	return value
}

// writing 1 to the joypad will enable strobe
// mode and reset the buttonIndex = 0
func (j *joypad) write(value byte) {
	j.strobe = value&1 == 1
	if j.strobe {
		j.buttonIndex = 0
	}
}
