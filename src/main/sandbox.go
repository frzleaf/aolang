package main

import (
	"fmt"
	"net"
	"proxy"
)

func main() {
	//ip, _ := findLocalIp("ist02.com:9999")
	//fmt.Println(ip)
	fmt.Print(proxy.FindPortOpen("localhost", []int{6111, 6112, 6110}))
}

func findLocalIp(targetAddr string) (string, error) {
	if dial, err := net.Dial("tcp", targetAddr); err != nil {
		return "", err
	} else {
		return dial.LocalAddr().String(), nil
	}
}
