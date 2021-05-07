package main

import (
	"bufio"
	"fmt"
	"os"
	"proxy"
	"strings"
)

func main() {
	client := proxy.NewClient()

	var hostAddr string
	if len(os.Args) < 2 {
		fmt.Println("Không tìm thấy host, vui lòng nhập địa chỉ host và port:")
		fmt.Println("Ví dụ: 192.168.0.1:9999 hoặc lancraft.com:9999")

		reader := bufio.NewReader(os.Stdin)
		hostAddr, _ = reader.ReadString('\n')
		hostAddr = strings.TrimSpace(hostAddr)
	} else {
		hostAddr = os.Args[1]
	}

	client.Run(hostAddr)
}
