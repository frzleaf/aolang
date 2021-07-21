package proxy

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Controller handle user interacting
type Controller struct {
	autoPing bool
}

func NewController() *Controller {
	return &Controller{}
}

func (s *Controller) BroadCast(client Client) (err error) {
	if s.autoPing {
		return
	}
	s.autoPing = true
	addr := &net.UDPAddr{
		Port: client.GameConfig().UdpPort,
		IP:   net.ParseIP(client.ServerConnector().LocalAddr()),
	}
	if udp, err := net.DialUDP("udp", nil, addr); err != nil {
		return err
	} else {
		go func() {
			defer func() {
				s.autoPing = false
			}()
			for {
				for i := 0; i < 2 && !client.OnMatch(); i++ {
					if _, err2 := udp.Write(
						[]byte{0xf7, 0x2f, 0x10, 0x00, 0x50, 0x58, 0x33, 0x57, 0x18, 0x00, 0x00, 0x00, byte(i), 0x00, 0x00, 0x00},
					); err2 != nil {
						return
					}
				}
				time.Sleep(time.Second)
			}
		}()
	}
	return nil
}

func (s *Controller) InteractOnClient(client Client) {
	for {
		reader := bufio.NewReader(os.Stdin)
		line, _, err := reader.ReadLine()
		if err != nil {
			LOG.Info("Command error: ", err)
			continue
		}
		lineStr := string(line)
		split := strings.Split(lineStr, " ")
		if len(split) > 0 {
			switch split[0] {
			case "/to":
				input2nd, err := strconv.Atoi(split[1])
				if err != nil {
					LOG.Info("Invalid args: ", err)
					continue
				}
				if input2nd != client.ServerConnector().connectionId {
					client.SelectTargetId(input2nd)
					LOG.Info("Selected: ", input2nd)
				} else {
					LOG.Info("Please chose another target")
				}
			case "/find":
			case "/exit":
			case "/ping":
				if err := s.BroadCast(client); err != nil {
					LOG.Error("ping error", err)
				}
			default:
				if _, err = client.ServerConnector().sendData(PackageTypeConverse, client.TargetId(), line); err != nil {
					LOG.Error(err)
				}
			}
		}
	}
}
