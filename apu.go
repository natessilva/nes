package nes

import (
	"encoding/binary"
	"math"
)

var lengthCounterTable = [32]byte{
	10, 254, 20, 2, 40, 4, 80, 6, 160, 8, 60, 10, 14, 12, 26, 14,
	12, 16, 24, 18, 48, 20, 96, 22, 192, 24, 72, 26, 16, 28, 32, 30,
}

type apu struct {
	pulse1   pulse
	pulse2   pulse
	triangle triangle
	noise    *noise

	frameMode       byte // 0 = 4-step, 1 = 5-step
	frameIRQInhibit bool
	frameCycle      int

	cycle uint64 // total CPU cycles, used for even/odd

	sampleAccumulator float64
	cyclesPerSample   float64

	sampleBuffer [736]float32
	sampleIndex  int
	sampleBytes  [2944]byte
}

func newAPU() *apu {
	a := &apu{
		noise:           newNoise(),
		cyclesPerSample: 1789773.0 / 44100.0, // NTSC CPU clock / sample rate
	}
	a.pulse1.channel = 1
	a.pulse2.channel = 2
	return a
}

func (a *apu) step() {
	a.cycle++

	// Triangle timer clocks every CPU cycle
	a.triangle.stepTimer()

	// Pulse and noise timers clock every APU cycle (every 2 CPU cycles)
	if a.cycle%2 == 0 {
		a.pulse1.stepTimer()
		a.pulse2.stepTimer()
		a.noise.stepTimer()
	}

	// Frame counter
	a.frameCycle++
	if a.frameMode == 0 {
		// 4-step mode
		switch a.frameCycle {
		case 3729:
			a.quarterFrame()
		case 7457:
			a.quarterFrame()
			a.halfFrame()
		case 11186:
			a.quarterFrame()
		case 14915:
			a.quarterFrame()
			a.halfFrame()
			a.frameCycle = 0
		}
	} else {
		// 5-step mode
		switch a.frameCycle {
		case 3729:
			a.quarterFrame()
		case 7457:
			a.quarterFrame()
			a.halfFrame()
		case 11186:
			a.quarterFrame()
		case 18641:
			a.quarterFrame()
			a.halfFrame()
			a.frameCycle = 0
		}
	}

	// Downsampling
	a.sampleAccumulator++
	if a.sampleAccumulator >= a.cyclesPerSample {
		a.sampleAccumulator -= a.cyclesPerSample
		if a.sampleIndex < len(a.sampleBuffer) {
			a.sampleBuffer[a.sampleIndex] = a.output()
			a.sampleIndex++
		}
	}
}

func (a *apu) quarterFrame() {
	a.pulse1.stepEnvelope()
	a.pulse2.stepEnvelope()
	a.triangle.stepLinearCounter()
	a.noise.stepEnvelope()
}

func (a *apu) halfFrame() {
	a.pulse1.stepLength()
	a.pulse2.stepLength()
	a.triangle.stepLength()
	a.noise.stepLength()
	a.pulse1.stepSweep()
	a.pulse2.stepSweep()
}

func (a *apu) output() float32 {
	p1 := float64(a.pulse1.output())
	p2 := float64(a.pulse2.output())
	t := float64(a.triangle.output())
	n := float64(a.noise.output())

	var pulseOut float64
	if p1+p2 > 0 {
		pulseOut = 95.88 / ((8128.0 / (p1 + p2)) + 100.0)
	}

	var tndOut float64
	tndSum := t/8227.0 + n/12241.0
	if tndSum > 0 {
		tndOut = 159.79 / (1.0/tndSum + 100.0)
	}

	return float32(pulseOut + tndOut)
}

func (a *apu) writeRegister(address uint16, value byte) {
	switch {
	case address == 0x4000:
		a.pulse1.duty = (value >> 6) & 3
		a.pulse1.lengthHalt = value&0x20 != 0
		a.pulse1.envelopeLoop = value&0x20 != 0
		a.pulse1.envelopeConstant = value&0x10 != 0
		a.pulse1.envelopeVolume = value & 0x0F
	case address == 0x4001:
		a.pulse1.sweepEnabled = value&0x80 != 0
		a.pulse1.sweepPeriod = (value >> 4) & 7
		a.pulse1.sweepNegate = value&0x08 != 0
		a.pulse1.sweepShift = value & 7
		a.pulse1.sweepReload = true
	case address == 0x4002:
		a.pulse1.timerPeriod = (a.pulse1.timerPeriod & 0xFF00) | uint16(value)
	case address == 0x4003:
		a.pulse1.timerPeriod = (a.pulse1.timerPeriod & 0x00FF) | (uint16(value&7) << 8)
		if a.pulse1.enabled {
			a.pulse1.lengthCounter = lengthCounterTable[value>>3]
		}
		a.pulse1.dutyIndex = 0
		a.pulse1.envelopeStart = true
	case address == 0x4004:
		a.pulse2.duty = (value >> 6) & 3
		a.pulse2.lengthHalt = value&0x20 != 0
		a.pulse2.envelopeLoop = value&0x20 != 0
		a.pulse2.envelopeConstant = value&0x10 != 0
		a.pulse2.envelopeVolume = value & 0x0F
	case address == 0x4005:
		a.pulse2.sweepEnabled = value&0x80 != 0
		a.pulse2.sweepPeriod = (value >> 4) & 7
		a.pulse2.sweepNegate = value&0x08 != 0
		a.pulse2.sweepShift = value & 7
		a.pulse2.sweepReload = true
	case address == 0x4006:
		a.pulse2.timerPeriod = (a.pulse2.timerPeriod & 0xFF00) | uint16(value)
	case address == 0x4007:
		a.pulse2.timerPeriod = (a.pulse2.timerPeriod & 0x00FF) | (uint16(value&7) << 8)
		if a.pulse2.enabled {
			a.pulse2.lengthCounter = lengthCounterTable[value>>3]
		}
		a.pulse2.dutyIndex = 0
		a.pulse2.envelopeStart = true
	case address == 0x4008:
		a.triangle.controlFlag = value&0x80 != 0
		a.triangle.lengthHalt = value&0x80 != 0
		a.triangle.linearReload = value & 0x7F
	case address == 0x400A:
		a.triangle.timerPeriod = (a.triangle.timerPeriod & 0xFF00) | uint16(value)
	case address == 0x400B:
		a.triangle.timerPeriod = (a.triangle.timerPeriod & 0x00FF) | (uint16(value&7) << 8)
		if a.triangle.enabled {
			a.triangle.lengthCounter = lengthCounterTable[value>>3]
		}
		a.triangle.linearReloadFlag = true
	case address == 0x400C:
		a.noise.lengthHalt = value&0x20 != 0
		a.noise.envelopeLoop = value&0x20 != 0
		a.noise.envelopeConstant = value&0x10 != 0
		a.noise.envelopeVolume = value & 0x0F
	case address == 0x400E:
		a.noise.mode = value&0x80 != 0
		a.noise.timerPeriod = noisePeriodTable[value&0x0F]
	case address == 0x400F:
		if a.noise.enabled {
			a.noise.lengthCounter = lengthCounterTable[value>>3]
		}
		a.noise.envelopeStart = true
	case address == 0x4015:
		a.pulse1.enabled = value&1 != 0
		if !a.pulse1.enabled {
			a.pulse1.lengthCounter = 0
		}
		a.pulse2.enabled = value&2 != 0
		if !a.pulse2.enabled {
			a.pulse2.lengthCounter = 0
		}
		a.triangle.enabled = value&4 != 0
		if !a.triangle.enabled {
			a.triangle.lengthCounter = 0
		}
		a.noise.enabled = value&8 != 0
		if !a.noise.enabled {
			a.noise.lengthCounter = 0
		}
	case address == 0x4017:
		a.frameMode = (value >> 7) & 1
		a.frameIRQInhibit = value&0x40 != 0
		if a.frameMode == 1 {
			a.quarterFrame()
			a.halfFrame()
		}
	}
}

func (a *apu) readStatus() byte {
	var status byte
	if a.pulse1.lengthCounter > 0 {
		status |= 1
	}
	if a.pulse2.lengthCounter > 0 {
		status |= 2
	}
	if a.triangle.lengthCounter > 0 {
		status |= 4
	}
	if a.noise.lengthCounter > 0 {
		status |= 8
	}
	return status
}

func (a *apu) flushSamples() ([]byte, int) {
	count := a.sampleIndex
	for i := 0; i < count; i++ {
		bits := math.Float32bits(a.sampleBuffer[i])
		binary.LittleEndian.PutUint32(a.sampleBytes[i*4:], bits)
	}
	a.sampleIndex = 0
	return a.sampleBytes[:count*4], count
}
