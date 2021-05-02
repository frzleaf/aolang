package main

import (
	"proxy"
)

func main() {

	server := proxy.NewServer()
	server.Start("10.0.1.105:9999")

}
