package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func main() {
	tcpServer, err := net.Listen("tcp", "0.0.0.0:7128")
	if err != nil {
		log.Fatal(err)
	}
	defer tcpServer.Close()
	connectionMap := make(map[string]net.Conn)
	connectionCounter := 0
	for {
		client, err := tcpServer.Accept()
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()
		connectionMap[strconv.Itoa(connectionCounter)] = client
		connectionCounter = connectionCounter + 1
		fmt.Println("New connection from: ", client.RemoteAddr())
		if _, err = client.Write([]byte("Your ID: " + strconv.Itoa(connectionCounter))); err != nil {
			fmt.Println(err)
		}
		go func() {
			go func() {
				newChar := make([]byte, 1)
				buf := make([]byte, 1024)
				bufRead := 0
				name := client.RemoteAddr()
				fmt.Println("Connection from: ", name)
				var targetConnection net.Conn
				for {
					if read, err := client.Read(newChar); err != nil {
						fmt.Println(err)
						break
					} else {
						if newChar[0] == byte('\n') {
							message := string(buf[0:bufRead])
							if strings.HasPrefix(message, "to") {
								target := strings.Replace(message, "to", "", 1)
								if strings.Trim(target, " \r\n\t") != "" {
									targetConnection = connectionMap[strings.TrimSpace(target)]
									if targetConnection != nil {
										client.Write([]byte("You are connecting to: " + targetConnection.RemoteAddr().String()))
									}
								}
							} else if targetConnection != nil {
								targetConnection.Write(buf[0:bufRead])
							} else {
								fmt.Println(message)
							}
							buf = make([]byte, 1024)
							bufRead = 0
						} else {
							copy(buf[bufRead:bufRead+read], newChar)
							bufRead = bufRead + read
						}
					}
				}
			}()
		}()
	}

}
