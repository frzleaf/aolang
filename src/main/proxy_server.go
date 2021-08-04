package main

import (
	"log"
	"os"
	"proxy"
)

func main() {

	server := proxy.NewServer()
	if len(os.Args) > 1 {
		server.Start(os.Args[1])
	} else {
		log.Fatalln("Please run with args, example: aolang_server 0.0.0.0:9999")
	}

}
