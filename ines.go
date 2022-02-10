package nes

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
	Flags6      byte
	Flags7      byte
	NumRAM      byte
	Unused      [7]byte
}

// Every .nes file starts with ASCII NES followed by $1A
const magicNumber = 0x1a53454e

// Load a file, read the header and PRG-ROM and CHR-ROM
func LoadFile(path string) (*Cart, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't open")
	}
	defer file.Close()
	return ReadFile(file)
}

func ReadFile(r io.Reader) (*Cart, error) {
	header := iNESHeader{}
	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return nil, errors.Wrap(err, "read")
	}

	if header.MagicNumber != magicNumber {
		return nil, errors.New("Invalid nes file")
	}

	prg := make([]byte, int(header.NumPRG)*0x4000)
	_, err = io.ReadFull(r, prg)
	if err != nil {
		return nil, errors.Wrap(err, "PRG")
	}

	chr := make([]byte, int(header.NumCHR)*0x2000)
	_, err = io.ReadFull(r, chr)
	if err != nil {
		return nil, errors.Wrap(err, "CHR")
	}

	mirror := header.Flags6 & 1

	return &Cart{
		Mirror: mirror,
		PRG:    prg,
		CHR:    chr,
	}, nil
}
