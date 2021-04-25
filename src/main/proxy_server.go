package main

import (
	"proxy"
)

func main() {

	server := proxy.NewServer()
	server.Start(":9999")

}
