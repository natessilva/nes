package main

import (
	"fmt"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8000", http.FileServer(http.Dir("web")))
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
