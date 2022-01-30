package main

import (
	"log"
	"os"
)

func main() {
	file := os.Args[1]
	_, err := loadFile(file)
	if err != nil {
		log.Fatal(err)
	}
}
