package test

import (
	"fmt"
	"net"
	"proxy"
	"strconv"
	"strings"
	"testing"
	"time"
)

var serverAddr = "localhost:9999"
var server = proxy.NewServer()

var gameConfig = &proxy.GameConfig{
	UdpPort: 10100,
	TcpPort: 10101,
	LocalIp: "127.0.0.1",
}

// Set up before testing
func init() {
	go func() {
		server.Start(serverAddr)
	}()
}

func createMokApp() {

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
			host := proxy.NewHost(serverAddr)
			guest := proxy.NewGuest(serverAddr, gameConfig)
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
			SEND_STR := "123456"
			host := NewMockHostApp(gameConfig)
			guest := NewMockGuestApp(gameConfig)

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
				if receivedStr != SEND_STR {
					t.Error("received message not as expected", receivedStr, SEND_STR)
				}
				host.sendData([]byte(strconv.Itoa(len(bytes))), guestId)
			})

			guest.OnHostReply(func(bytes []byte) {
				sendStrLen := strconv.Itoa(len(SEND_STR))
				if string(bytes) != sendStrLen {
					t.Error("received message not as expected", string(bytes), sendStrLen)
				}
			})

			if err := guest.sendDataToHost([]byte(SEND_STR)); err != nil {
				t.Error(err)
			} else {
				time.Sleep(time.Millisecond * 100)
				err = guest.sendDataToHost([]byte(SEND_STR))
				time.Sleep(time.Millisecond * 100)
				err = guest.sendDataToHost([]byte(SEND_STR))
				time.Sleep(time.Millisecond * 100)
				err = guest.sendDataToHost([]byte(SEND_STR))
				time.Sleep(time.Millisecond * 100)
				err = guest.sendDataToHost([]byte(SEND_STR))
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
			SEND_STR := "123456"
			host := NewMockHostApp(gameConfig)
			guest := NewMockGuestApp(gameConfig)

			if err := host.listenUdp(); err != nil {
				t.Error(err)
				t.FailNow()
			} else {
				defer host.close()
			}

			host.onUdpData(func(bytes []byte, conn *net.UDPConn) {
				receivedStr := string(bytes)
				fmt.Printf("received broadcast msg (%v): %v\n", conn.RemoteAddr(), receivedStr)
				if receivedStr != SEND_STR {
					t.Error("message not as expected", receivedStr, SEND_STR)
				}
			})

			guest.broadCast([]byte(SEND_STR))

			time.Sleep(time.Second * 3)
		},
	)
}
