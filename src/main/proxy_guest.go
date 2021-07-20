package main

import (
	"log"
	"os"
	"proxy"
)

func main() {
	guest := proxy.NewGuest(os.Args[1], proxy.Warcraft3Config)
	if err := guest.ConnectServer(); err != nil {
		log.Fatalln(err)
	}
}
