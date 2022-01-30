package main

import (
	"log"
	"os"
)

func main() {
	file := os.Args[1]
	err := loadFile(file)
	if err != nil {
		log.Fatal(err)
	}
}
