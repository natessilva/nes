package main

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/pkg/errors"
)

// Every .nes file starts with the 16 byte header
type iNESHeader struct {
	MagicNumber int32
	NumPRG      byte
	NumCHR      byte
	Control1    byte
	Control2    byte
	NumRAM      byte
	_           [7]byte
}

// Every .nes file starts with ASCII NES followed by $1A
const magicNumber = 0x1a53454e

// Load a file, read the header and PRG-ROM and CHR-ROM
func loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "couldn't open")
	}
	defer file.Close()

	header := iNESHeader{}
	err = binary.Read(file, binary.LittleEndian, &header)
	if err != nil {
		return errors.Wrap(err, "read")
	}

	if header.MagicNumber != magicNumber {
		return errors.New("Invalid nes file")
	}

	// TODO not doing anything with these bytes just yet
	prg := make([]byte, int(header.NumPRG)*0x4000)
	_, err = io.ReadFull(file, prg)
	if err != nil {
		return errors.Wrap(err, "PRG")
	}

	chr := make([]byte, int(header.NumCHR)*0x2000)
	_, err = io.ReadFull(file, chr)
	if err != nil {
		return errors.Wrap(err, "CHR")
	}

	return nil
}
