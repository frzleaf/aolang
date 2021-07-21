package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"proxy"
	"strings"
)

const LANCraftVersion = "0.3"

func main() {
	fmt.Println(getBanner())
	// Default setting
	serverAddr := ""
	mode := proxy.ClientModeGuest

	controller := proxy.NewController()

	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}
	if len(os.Args) > 2 {
		modeArg := strings.ToLower(strings.TrimSpace(os.Args[2]))
		if modeArg == proxy.ClientModeHost {
			mode = modeArg
		}
	}
	if len(os.Args) == 1 {
		serverAddr = readFromStdin("Vui lòng nhập server (vd: lancraft.net:9999) - ")
		promptMode := readFromStdin("Mode (guest/host) - ")
		mode = filterMode(promptMode)
	}

	var client proxy.Client
	switch mode {
	case proxy.ClientModeHost:
		client = proxy.NewHost(serverAddr, proxy.Warcraft3Config)
	default:
		client = proxy.NewGuest(serverAddr, proxy.Warcraft3Config)
	}

	go controller.InteractOnClient(client)
	fmt.Printf(`
------------------ Thông tin cài đặt ------------------
   			Server 		 	 : %v
   			Mode   		 	 : %v
			Lancraft version : %v
-------------------------------------------------------
`,
		serverAddr, mode, LANCraftVersion)
	if err := client.ConnectServer(); err != nil {
		log.Fatalln(err)
	}
}

func readFromStdin(promptInput string) (result string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(promptInput)
	result, _ = reader.ReadString('\n')
	result = strings.TrimSpace(result)
	return
}

func filterMode(inputMode string) string {
	modeArg := strings.ToLower(strings.TrimSpace(inputMode))
	switch modeArg {
	case proxy.ClientModeHost:
		return proxy.ClientModeHost
	default:
		return proxy.ClientModeGuest
	}
}

func getBanner() string {
	return `
██╗      █████╗ ███╗   ██╗ ██████╗██████╗  █████╗ ███████╗████████╗
██║     ██╔══██╗████╗  ██║██╔════╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝
██║     ███████║██╔██╗ ██║██║     ██████╔╝███████║█████╗     ██║   
██║     ██╔══██║██║╚██╗██║██║     ██╔══██╗██╔══██║██╔══╝     ██║   
███████╗██║  ██║██║ ╚████║╚██████╗██║  ██║██║  ██║██║        ██║   
╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝
`
}
