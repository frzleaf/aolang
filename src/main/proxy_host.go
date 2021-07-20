package main

import (
	"log"
	"os"
	"proxy"
)

func main() {
	host := proxy.NewHost(os.Args[1])
	if err := host.ConnectServer(); err != nil {
		log.Fatalln(err)
	}
}
