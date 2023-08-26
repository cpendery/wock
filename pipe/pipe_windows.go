//go:build windows

package pipe

import (
	"fmt"
	"net"

	winio "github.com/Microsoft/go-winio"
)

const (
	wockServerPipe = `\\.\pipe\wock`
)

func DialServer() (net.Conn, error) {
	return winio.DialPipe(wockServerPipe, nil)
}

func DialClient(clientId string) (net.Conn, error) {
	return winio.DialPipe(fmt.Sprintf("%s-%s", wockServerPipe, clientId), nil)
}

func ServerListen() (net.Listener, error) {
	return winio.ListenPipe(wockServerPipe, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;AU)"})
}

func ClientListen(clientId string) (net.Listener, error) {
	return winio.ListenPipe(fmt.Sprintf("%s-%s", wockServerPipe, clientId), &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;AU)"})
}

func IsOpen() bool {
	l, err := winio.ListenPipe(wockServerPipe, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;AU)"})
	if err != nil {
		return true
	}
	err = l.Close()
	if err != nil {

	}
	return false
}
