package nes

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestLogFile(t *testing.T) {
	file, err := os.Open("./nestest.nes")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	cart, err := readFile(file)
	if err != nil {
		t.Fatal(err)
	}
	ppu := newPPU(cart)
	cpu := newCPU(cart, ppu)
	// nestest automation mode starts at 0xC000
	cpu.pc = 0xC000

	logFile, err := os.Open("./nestest.log")
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()
	scanner := bufio.NewScanner(logFile)

	// CPU starts at cycle 7 and PPU starts a cycle 21 in the log file
	totalCycles := 7
	for i := 0; i < 21; i++ {
		ppu.Step()
	}

	for scanner.Scan() {
		text := scanner.Text()
		expectedPC := text[0:4]
		actualPC := fmt.Sprintf("%04X", cpu.pc)
		if expectedPC != actualPC {
			t.Fatalf("pc = %v, want %v", actualPC, expectedPC)
		}
		expectedOpcode := text[6:8]
		actualOpcode := fmt.Sprintf("%02X", cpu.readByte(cpu.pc))
		if expectedOpcode != actualOpcode {
			t.Fatalf("opcode = %v, want %v", expectedOpcode, actualOpcode)
		}

		expectedA := text[50:52]
		actualA := fmt.Sprintf("%02X", cpu.a)
		if expectedA != actualA {
			t.Fatalf("a = %v, want %v", actualA, expectedA)
		}

		expectedX := text[55:57]
		actualX := fmt.Sprintf("%02X", cpu.x)
		if expectedX != actualX {
			t.Fatalf("x = %v, want %v", actualX, expectedX)
		}

		expectedY := text[60:62]
		actualY := fmt.Sprintf("%02X", cpu.y)
		if expectedY != actualY {
			t.Fatalf("y = %v, want %v", actualY, expectedY)
		}

		expectedStatus := text[65:67]
		actualStatus := fmt.Sprintf("%02X", cpu.status)
		if expectedStatus != actualStatus {
			t.Fatalf("status = %v, want %v", actualStatus, expectedStatus)
		}

		expectedSP := text[71:73]
		actualSP := fmt.Sprintf("%02X", cpu.sp)
		if expectedSP != actualSP {
			t.Fatalf("sp = %v, want %v", actualSP, expectedSP)
		}

		expectedScanLine := strings.TrimSpace(text[78:81])
		actualScanLine := fmt.Sprintf("%d", ppu.scanline)
		if expectedScanLine != actualScanLine {
			t.Fatalf("PPU scanLine = %v, want %v", actualScanLine, expectedScanLine)
		}

		expectedPPUCycles := strings.TrimSpace(text[82:85])
		actualPPUCycles := fmt.Sprintf("%d", ppu.cycle)
		if expectedPPUCycles != actualPPUCycles {
			t.Fatalf("PPU cycle = %v, want %v", actualPPUCycles, expectedPPUCycles)
		}

		expectedCycles := text[90:]
		actualCycles := fmt.Sprintf("%d", totalCycles)
		if expectedCycles != actualCycles {
			t.Fatalf("cycles = %v, want %v", actualCycles, expectedCycles)
		}

		cycles := cpu.Step()
		totalCycles += cycles
		cycles *= 3
		for i := 0; i < cycles; i++ {
			ppu.Step()
		}
	}
}
