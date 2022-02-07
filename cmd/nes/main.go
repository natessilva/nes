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
	cycles := 0
	for i := 0; i < 10; i++ {
		c := cpu.Step()
		cycles += c
		fmt.Println("cycles", c)
	}
	fmt.Println("total cycles", cycles)
}
