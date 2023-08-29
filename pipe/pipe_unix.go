//go:build !windows

package pipe

import (
	"fmt"
	"net"
)

const (
	wockSocket             = "/tmp/wock.sock"
	wockClientSocketPrefix = "/tmp/wock"
)

func DialServer() (net.Conn, error) {
	return net.Dial("unix", wockSocket)
}

func DialClient(clientId string) (net.Conn, error) {
	return net.Dial("unix", fmt.Sprintf("%s-%s.sock", wockClientSocketPrefix, clientId))
}

func ServerListen() (net.Listener, error) {
	return net.Listen("unix", wockSocket)
}

func ClientListen(clientId string) (net.Listener, error) {
	return net.Listen("unix", fmt.Sprintf("%s-%s.sock", wockClientSocketPrefix, clientId))
}

func IsOpen() bool {
	l, err := net.Listen("unix", wockSocket)
	if err != nil {
		return true
	}
	err = l.Close()
	if err != nil {

	}
	return false
}
