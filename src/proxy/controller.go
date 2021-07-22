package proxy

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	CmdPrefixTo        = "/to"
	CmdPrefixFind      = "/find"
	CmdPrefixFindShort = "/f"
	CmdPrefixExit      = "/exit"
	CmdPrefixPing      = "/ping"
	CmdPrefixPingShort = "/p"
	CmdPrefixHelp      = "/"
)

const (
	ControllerStatusIdle = iota
	ControllerStatusInteracting
	ControllerStatusStop
)

// Controller handle user interacting
type Controller struct {
	pinging bool
	status  int
}

func NewController() *Controller {
	return &Controller{
		status:  ControllerStatusIdle,
		pinging: false,
	}
}

func (c *Controller) BroadCast(client Client) (err error) {
	if c.pinging {
		return
	}
	c.pinging = true
	addr := &net.UDPAddr{
		Port: client.GameConfig().UdpPort,
		IP:   net.ParseIP(client.ServerConnector().LocalAddr()),
	}
	if udp, err := net.DialUDP("udp", nil, addr); err != nil {
		return err
	} else {
		go func() {
			defer func() {
				c.pinging = false
			}()
			for c.pinging {
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

func (c *Controller) InteractOnClient(client Client) {
	if c.status != ControllerStatusIdle {
		LOG.Error("can not interact")
		return
	}
	c.printListCmd()
	c.ChangeStatus(ControllerStatusInteracting)
	for c.status == ControllerStatusInteracting {
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
			case CmdPrefixTo:
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
			case CmdPrefixFind, CmdPrefixFindShort:
				client.ServerConnector().sendData(PackageTypeClientStatus, ServerConnectorID, nil)
			case CmdPrefixExit:
				c.StopInteract()
				c.ChangeStatus(ControllerStatusStop)
				client.Close()
			case CmdPrefixPing, CmdPrefixPingShort:
				if c.pinging {
					c.pinging = false
				} else {
					if err := c.BroadCast(client); err != nil {
						LOG.Error("ping error", err)
					}
				}
			case CmdPrefixHelp:
				c.printListCmd()
			default:
				if _, err = client.ServerConnector().sendData(PackageTypeConverse, client.TargetId(), line); err != nil {
					LOG.Error(err)
				}
			}
		}
	}
}

func (c *Controller) StopInteract() {
	c.ChangeStatus(ControllerStatusIdle)
	c.pinging = false
}

func (c *Controller) ChangeStatus(newStatus int) {
	if c.status == ControllerStatusStop {
		return
	}
	c.status = newStatus
}

func (c *Controller) IsStop() bool {
	return c.status == ControllerStatusStop
}

func (c *Controller) printListCmd() {
	fmt.Printf(`
 Danh sách các lệnh:
   - %-12s: kết nối tới host (tìm <hostId> bằng lệnh %v)
   - %-12s: tìm kiếm các client cùng server
   - %-12s: thoát game
   - %-12s: bật/tắt tự động tìm kiếm nếu không thấy host
   - %-12s: tra cứu lệnh
`,
		CmdPrefixTo+" <hostId>", CmdPrefixFind,
		CmdPrefixFind+"("+CmdPrefixFindShort+")",
		CmdPrefixExit,
		CmdPrefixPing+"("+CmdPrefixPingShort+")",
		CmdPrefixHelp,
	)
}
