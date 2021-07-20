package proxy

import (
	"errors"
	"io"
	"net"
	"time"
)

type LocalConnector struct {
	localAddr   string
	gConn       net.Conn
	gListener   net.Listener
	udpListener net.Listener
	onDataFunc  func([]byte)
}

func NewLocalConnector(localAddr string) *LocalConnector {
	return &LocalConnector{
		localAddr: localAddr,
	}
}

func (l *LocalConnector) close() (err error) {
	if l.gConn != nil {
		if err = l.gConn.Close(); err == nil {
			l.gConn = nil
		}
	}
	if l.gListener != nil {
		if err = l.gListener.Close(); err == nil {
			l.gListener = nil
		}
	}
	return
}

func (l *LocalConnector) sentBytes(data []byte) (int, error) {
	if l.gConn != nil {
		return l.gConn.Write(data)
	} else {
		return 0, errors.New("LocalConnector have not initialized")
	}
}

func (l *LocalConnector) sentBytesTo(data []byte, targetAddr string) (int, error) {
	if dial, err := net.Dial("tcp", targetAddr); err != nil {
		return 0, err
	} else {
		return dial.Write(data)
	}
}

func (l *LocalConnector) onData(onDataFunc func([]byte)) {
	l.onDataFunc = onDataFunc
}

func (l *LocalConnector) startListen() error {
	if l.gConn == nil {
		listener, err := net.Listen("tcp", l.localAddr)
		if err != nil {
			return err
		} else {
			if l.gConn, err = listener.Accept(); err != nil {
				return err
			} else {
				buffered := make([]byte, 1000)
				for {
					read, err := l.gConn.Read(buffered)
					if err != nil {
						if err == io.EOF {
							time.Sleep(time.Millisecond * 100)
							continue
						} else {
							break
						}
					}
					l.onDataFunc(buffered[0:read])
				}
			}
		}
	}
	return errors.New("local connection is not empty")
}

func (l *LocalConnector) startListenUdp(localUdpAddr string, onUdpFunc func(data []byte)) (err error) {
	if l.udpListener, err = net.Listen("udp", localUdpAddr); err != nil {
		return err
	} else {
		go func() {
			if accept, err := l.udpListener.Accept(); err != nil {
				LOG.Error("listenUdp error", err)
				return
			} else {
				buffered := make([]byte, 1000)
				for {
					read, err := accept.Read(buffered)
					if err != nil || read == 0 {
						accept.Close()
						break
					}
					onUdpFunc(buffered[0:read])
				}
			}
		}()
	}
	return
}
