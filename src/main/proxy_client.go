package main

import (
	"log"
	"os"
	"proxy"
)

func main() {
	controller := proxy.NewController()
	switch os.Args[2] {
	case "host":
		host := proxy.NewHost(os.Args[1], proxy.Warcraft3Config)
		go controller.InteractOnClient(host)
		if err := host.ConnectServer(); err != nil {
			log.Fatalln(err)
		}
	case "guest":
		guest := proxy.NewGuest(os.Args[1], proxy.Warcraft3Config)
		go controller.InteractOnClient(guest)
		if err := guest.ConnectServer(); err != nil {
			log.Fatalln(err)
		}
	}
}
