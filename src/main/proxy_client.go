package main

import (
	"os"
	"proxy"
)

func main() {

	client := proxy.NewClient()
	client.Connect(os.Args[1])

}
