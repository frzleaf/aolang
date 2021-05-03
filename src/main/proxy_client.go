package main

import (
	"os"
	"proxy"
)

func main() {
	client := proxy.NewClient()
	client.Run(os.Args[1])
}
