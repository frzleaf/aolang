package test

import (
	"fmt"
	"net"
	"os"
	"proxy"
	"strconv"
	"strings"
	"testing"
	"time"
)

var serverAddr = "localhost:9999"
var server = proxy.NewServer()
var SendStr = "123456"
var LOG *proxy.Logger

var configWithUdp = &proxy.GameConfig{
	UdpPort: 10100,
	TcpPort: 10101,
	LocalIp: "127.0.0.1",
}

var hostConfig = &proxy.GameConfig{
	TcpPort: 10200,
	LocalIp: "0.0.0.0",
}

// Set up before testing
func init() {
	LOG = proxy.NewLoggerWithLevel(os.Stdout, proxy.DebugLevel)
	go func() {
		server.Start(serverAddr)
	}()
}

func TestIntegrate(t *testing.T) {
	t.Run(
		"testIntegrate",
		func(t *testing.T) {
			host := proxy.NewHost(serverAddr, hostConfig)

			go func() {
				if err := host.ConnectServer(); err != nil {
					t.Error(err)
					t.FailNow()
				} else {
					defer host.Close()
				}
			}()
			hostApp := NewMockHostApp(hostConfig)
			if err := hostApp.listen(); err != nil {
				t.Error(err)
				t.FailNow()
			} else {
				hostApp.OnData(func(data []byte, i int) {
					LOG.Debugf("Receive message from #%v: %v\n", i, string(data))
					receivedStr := string(data)
					if receivedStr != SendStr {
						t.Error("message not as expected", receivedStr, SendStr)
					}
					hostApp.sendData([]byte(strconv.Itoa(len(data))), i)
				})
				defer hostApp.close()
			}
			time.Sleep(time.Second)

			// Open clients connect to host concurrently
			for i := 0; i < 100; i++ {

				guestConfig := &proxy.GameConfig{
					TcpPort: 10100 + i,
					LocalIp: "0.0.0.0",
				}
				guest := proxy.NewGuest(serverAddr, guestConfig)
				go func() {
					if err := guest.ConnectServer(); err != nil {
						t.Error(err)
						t.FailNow()
					} else {
						defer guest.Close()
					}
				}()
				guest.SelectHost(host.ConnectionId())

				go func() {
					guestApp := NewMockGuestApp(guestConfig)
					if err := guestApp.connectToHost("localhost"); err != nil {
						t.Error(err)
						t.FailNow()
					} else {
						guestApp.OnHostReply(func(data []byte) {
							LOG.Debugf("Receive message from host: %v\n", string(data))
							receivedStr := string(data)
							expectedResult := strconv.Itoa(len(SendStr))
							if receivedStr != expectedResult {
								t.Error("message not as expected", receivedStr, expectedResult)
							}
						})
						defer guestApp.close()
					}
					for j := 0; j < 5; j++ {
						if err := guestApp.sendDataToHost([]byte(SendStr)); err != nil {
							t.Error(err)
						}
						time.Sleep(time.Millisecond * 300)
					}
				}()
			}
			time.Sleep(time.Second * 5)
		},
	)

}

func TestServer(t *testing.T) {
	t.Run(
		"testAssignIdOnConnect",
		func(t *testing.T) {
			dial, err := net.Dial("tcp", serverAddr)
			if err != nil {
				t.Error(err)
			}
			buf := make([]byte, 1000)
			read, err := dial.Read(buf)
			if err != nil {
				t.Error(err)
			}
			packages := proxy.PacketFromBytes(buf[0:read])
			if len(packages) == 0 {
				t.Error("no package found")
			}
			assignIdPkg := packages[0]
			serverCmd := string(assignIdPkg.Data())
			split := strings.Split(serverCmd, " ")
			if len(split) < 1 {
				t.Error("not assign id")
			}
			_, err = strconv.Atoi(split[1])
			if err != nil {
				t.Error("Invalid connection ID: ", split[1])
			}
		},
	)
}

func TestHost(t *testing.T) {
	t.Run(
		"testHostAndGuestConnection",
		func(t *testing.T) {
			host := proxy.NewHost(serverAddr, configWithUdp)
			guest := proxy.NewGuest(serverAddr, configWithUdp)
			go func() {
				host.ConnectServer()
			}()
			go func() {
				guest.ConnectServer()
			}()
		},
	)
}

func TestMockApp(t *testing.T) {
	t.Run(
		"Mock app data interacting",
		func(t *testing.T) {
			host := NewMockHostApp(configWithUdp)
			guest := NewMockGuestApp(configWithUdp)

			if err := host.listen(); err != nil {
				t.Error(err)
				t.FailNow()
			} else {
				defer host.close()
			}
			if err := guest.connectToHost("localhost"); err != nil {
				t.Error(err)
				t.FailNow()
			} else {
				defer guest.close()
			}

			host.OnData(func(bytes []byte, guestId int) {
				receivedStr := string(bytes)
				if receivedStr != SendStr {
					t.Error("received message not as expected", receivedStr, SendStr)
				}
				host.sendData([]byte(strconv.Itoa(len(bytes))), guestId)
			})

			guest.OnHostReply(func(bytes []byte) {
				sendStrLen := strconv.Itoa(len(SendStr))
				if string(bytes) != sendStrLen {
					t.Error("received message not as expected", string(bytes), sendStrLen)
				}
			})

			if err := guest.sendDataToHost([]byte(SendStr)); err != nil {
				t.Error(err)
			} else {
				time.Sleep(time.Millisecond * 100)
				err = guest.sendDataToHost([]byte(SendStr))
				time.Sleep(time.Millisecond * 100)
				err = guest.sendDataToHost([]byte(SendStr))
				time.Sleep(time.Millisecond * 100)
				err = guest.sendDataToHost([]byte(SendStr))
				time.Sleep(time.Millisecond * 100)
				err = guest.sendDataToHost([]byte(SendStr))
				if err != nil {
					t.Error(err)
				}
			}
			time.Sleep(time.Second * 3)
		},
	)
}

func TestShowLanIP(t *testing.T) {
	cloudFlareAddr := &net.TCPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Port: 80,
	}
	tcp, _ := net.DialTCP("tcp", nil, cloudFlareAddr)
	fmt.Printf(tcp.LocalAddr().String())

	//ifaces, _ := net.Interfaces()
	//// handle err
	//for _, i := range ifaces {
	//	addrs, _ := i.Addrs()
	//	// handle err
	//	for _, addr := range addrs {
	//		var ip net.IP
	//		switch v := addr.(type) {
	//		case *net.IPNet:
	//			ip = v.IP
	//		case *net.IPAddr:
	//			ip = v.IP
	//		}
	//		fmt.Println()
	//	}
	//}
}
func TestMockAppBroadcast(t *testing.T) {
	t.Run(
		"Mock app broadcast",
		func(t *testing.T) {
			host := NewMockHostApp(configWithUdp)
			guest := NewMockGuestApp(configWithUdp)

			if err := host.listenUdp(); err != nil {
				t.Error(err)
				t.FailNow()
			} else {
				defer host.close()
			}

			host.onUdpData(func(bytes []byte, conn *net.UDPConn) {
				receivedStr := string(bytes)
				fmt.Printf("received broadcast msg (%v): %v\n", conn.RemoteAddr(), receivedStr)
				if receivedStr != SendStr {
					t.Error("message not as expected", receivedStr, SendStr)
				}
			})

			guest.broadCast([]byte(SendStr))

			time.Sleep(time.Second * 3)
		},
	)
}
