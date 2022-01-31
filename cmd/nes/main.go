package main

import (
	"log"
	"os"

	"github.com/natessilva/nes"
)

func main() {
	file := os.Args[1]
	_, err := nes.LoadFile(file)
	if err != nil {
		log.Fatal(err)
	}
}
