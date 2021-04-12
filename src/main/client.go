package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	ln, err := net.Dial("udp", ":6112")
	if err != nil {
		log.Fatal(err)
	}
	chunkData := make([]byte, 50)
	for ;; {
		if read, err := ln.Read(chunkData); err != nil {
			log.Fatal(err)
		} else if read > 0 {
			fmt.Print(chunkData)
		}
	}
}