package main

import "proxy"

func main() {
	client := proxy.NewClient()
	client.Run("10.0.1.105:9999")
}
