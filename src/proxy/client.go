package proxy

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

var dstClient int

type Client struct {
	host *Host
}

func NewClient() *Client {
	return &Client{
		host: NewHost(),
	}
}

func (c *Client) Run(serverAddr string) {
	go func() {
		err := c.host.start(serverAddr)
		if err != nil {
			LOG.Error(err)
		}
	}()
	c.readCommand()
}

func (c *Client) readCommand() {
	for {
		reader := bufio.NewReader(os.Stdin)
		LOG.Info("Gửi lệnh lên server: ")
		line, _, err := reader.ReadLine()
		if err != nil {
			LOG.Info("Command error: ", err)
			continue
		}
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "/to ") {
			split := strings.Split(lineStr, " ")
			if len(split) == 1 {
				LOG.Info("Missing 2nd argument")
				continue
			} else {
				input2nd, err := strconv.Atoi(split[1])
				if err != nil {
					LOG.Info("Invalid args: ", err)
					continue
				}
				dstClient = input2nd
				LOG.Info("Selected: ", input2nd)
			}
		} else if strings.HasPrefix(lineStr, "/find") {

		} else if strings.HasPrefix(lineStr, "/exit") {
			c.host.Close()
			return
		} else {
			if err := c.host.SendDataToServer(NewInformPacket(c.host.id, dstClient, line)); err != nil {
				LOG.Error(err)
			}
		}
	}
}
