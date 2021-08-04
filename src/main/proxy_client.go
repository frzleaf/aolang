package main

import (
	"bufio"
	"fmt"
	"os"
	"proxy"
	"strings"
)

const AolangVersion = "0.3.1"

func main() {
	fmt.Println(getBanner())
	controller := proxy.NewController()

	client := clientByPrompt(true)

	for !controller.IsStop() {
		client.OnConnectSuccess(func(c proxy.Client) {
			controller.InteractOnClient(c)
		})
		if err := client.ConnectServer(); err != nil {
			fmt.Println("Lỗi kết nối server, vui lòng thử lại...")
			proxy.LOG.Debug(err)
			client = clientByPrompt(false)
		}
		client.Close()
	}
}

func clientByPrompt(fromArgs bool) proxy.Client {

	var serverAddr, mode string
	if fromArgs {
		if len(os.Args) > 1 {
			serverAddr = os.Args[1]
		}
		if len(os.Args) > 2 {
			mode = strings.ToLower(strings.TrimSpace(os.Args[2]))
		}
	}

	if serverAddr == "" {
		serverAddr = readLineFromStdin("Vui lòng nhập server (vd: aolang.net:9999) - ")
	}
	if mode == "" {
		mode = readLineFromStdin("Mode (vd: guest, host) - ")
	}
	mode = filterMode(mode)

	var client proxy.Client
	switch mode {
	case proxy.ClientModeHost:
		client = proxy.NewHost(serverAddr, proxy.Warcraft3Config)
	default:
		client = proxy.NewGuest(serverAddr, proxy.Warcraft3Config)
	}

	fmt.Printf(`
__________________ Thông tin cài đặt __________________

                     Server: %v
                       Mode: %v
_______________________________________________________

`,
		serverAddr, mode)
	return client
}

func readFromStdin(promptInput string, delim byte) (result string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(promptInput)
	result, _ = reader.ReadString(delim)
	result = strings.TrimSpace(result)
	return
}

func readLineFromStdin(promptInput string) (result string) {
	return readFromStdin(promptInput, '\n')
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
	return fmt.Sprintf(`
 █████╗  ██████╗     ██╗      █████╗ ███╗   ██╗ ██████╗ 
██╔══██╗██╔═══██╗    ██║     ██╔══██╗████╗  ██║██╔════╝ 
███████║██║   ██║    ██║     ███████║██╔██╗ ██║██║  ███╗
██╔══██║██║   ██║    ██║     ██╔══██║██║╚██╗██║██║   ██║
██║  ██║╚██████╔╝    ███████╗██║  ██║██║ ╚████║╚██████╔╝
╚═╝  ╚═╝ ╚═════╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ 
                     version %v
          (https://github.com/frzleaf/aolang)
`, AolangVersion)
}
