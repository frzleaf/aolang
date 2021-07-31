package proxy

import (
	"errors"
	"net"
)

type LocalConnector struct {
	localAddr  string
	gConn      net.Conn
	gListener  net.Listener
	udpConn    *net.UDPConn
	onDataFunc func([]byte)
	isStopped  bool
}

func NewLocalConnector(localAddr string) *LocalConnector {
	return &LocalConnector{
		localAddr: localAddr,
	}
}

func (l *LocalConnector) close() (err error) {
	if l.gConn != nil {
		l.gConn.Close()
		l.gConn = nil
	}
	if l.gListener != nil {
		l.gListener.Close()
		l.gListener = nil
	}
	l.isStopped = true
	return
}

func (l *LocalConnector) sentBytes(data []byte) (int, error) {
	if l.gConn != nil {
		return l.gConn.Write(data)
	} else {
		return 0, errors.New("LocalConnector have not initialized")
	}
}

func (l *LocalConnector) sentUdpBytesTo(data []byte, targetAddr string, port int) (int, error) {
	addr := &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(targetAddr),
	}
	if dial, err := net.DialUDP("udp", nil, addr); err != nil {
		return 0, err
	} else {
		defer dial.Close()
		return dial.Write(data)
	}
}

func (l *LocalConnector) onData(onDataFunc func([]byte)) {
	l.onDataFunc = onDataFunc
}

func (l *LocalConnector) startListen() (err error) {
	l.gListener, err = net.Listen("tcp", l.localAddr)
	if err != nil {
		return err
	}

	LOG.Debug("Proxy listen on: ", l.localAddr)
	defer l.gListener.Close()

	defer l.gListener.Close()

	// Only connection a time
	for !l.isStopped {
		if l.gConn, err = l.gListener.Accept(); err != nil {
			return err
		} else {
			buffered := CreateBuffer()
			for l.gConn != nil {
				read, err := l.gConn.Read(buffered)
				if err != nil {
					break
				}
				l.onDataFunc(buffered[0:read])
			}
			l.Release()
		}
	}
	return err
}

func (l *LocalConnector) startListenUdp(localAddr string, port int, onUdpFunc func(data []byte)) (err error) {
	udpAddr := &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(localAddr),
	}
	if udpConn, err := net.ListenUDP("udp", udpAddr); err != nil {
		return err
	} else {
		LOG.Debugf("Open udp listener at: %v", udpConn.LocalAddr())
		go func() {
			defer udpConn.Close()
			buffered := CreateBuffer()
			for {
				read, err := udpConn.Read(buffered)
				if err != nil {
					break
				}
				LOG.Debugf("Receive udp msg: %v\n", buffered[0:read])
				onUdpFunc(buffered[0:read])
			}
		}()
	}
	return
}

func (l *LocalConnector) Release() error {
	if l.gConn != nil {
		if err := l.gConn.Close(); err != nil {
			return errors.New("error on renewing: " + err.Error())
		}
		l.gConn = nil
	}
	return nil
}

func (l *LocalConnector) isOnline() bool {
	return l.gConn != nil
}
