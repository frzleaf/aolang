package proxy

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func FindPortOpen(host string, ports []int) (result int) {
	lock := sync.Mutex{}
	lock.Lock()
	result = -1
	finished := 0
	for i := range ports {
		port := ports[i]
		go func() {
			defer func() {
				finished++
				if finished == len(ports) && result == -1 {
					lock.Unlock()
				}
			}()
			server, error := net.DialTimeout("tcp", host+":"+strconv.Itoa(port), time.Second)
			if error != nil {
				return
			}
			server.Close()
			if result == -1 {
				lock.Unlock()
				result = port
			}
			return
		}()
	}
	lock.Lock()
	return
}

func CommandToString(command string, args ...interface{}) string {
	format := make([]string, 0)
	for i := 0; i < len(args); i++ {
		format = append(format, "%v")
	}
	return command + CharSplitCommand + fmt.Sprintf(strings.Join(format, CharSplitCommand), args...)
}
