package main

import (
	"os"
	"proxy"
)

func main() {

	server := proxy.NewServer()
	server.Start(os.Args[1])

}
