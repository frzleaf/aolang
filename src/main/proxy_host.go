package main

import (
	"log"
	"os"
	"proxy"
)

func main() {
	host := proxy.NewHost(os.Args[1], proxy.Warcraft3Config)
	if err := host.ConnectServer(); err != nil {
		log.Fatalln(err)
	}
}
