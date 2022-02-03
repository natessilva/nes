package main

import (
	"fmt"
	"log"
	"os"

	"github.com/natessilva/nes"
)

func main() {
	file := os.Args[1]
	cart, err := nes.LoadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	cpu := nes.NewCPU(cart)

	fmt.Printf("%x\n", cpu.ReadByte(cpu.PC))

	for i := 0; i < 10; i++ {
		cpu.Step()
	}
}
