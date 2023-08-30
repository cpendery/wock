//go:build !windows

package pipe

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"time"
)

const (
	wockSocket             = "/tmp/wock.sock"
	wockClientSocketPrefix = "/tmp/wock"
)

func DialServer() (net.Conn, error) {
	ctx, cancelCtx := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancelCtx()
	if err := waitForFile(wockSocket, ctx); err != nil {
		return nil, err
	}
	return net.Dial("unix", wockSocket)
}

func DialClient(clientId string) (net.Conn, error) {
	ctx, cancelCtx := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancelCtx()
	clientSocket := fmt.Sprintf("%s-%s.sock", wockClientSocketPrefix, clientId)
	if err := waitForFile(clientSocket, ctx); err != nil {
		return nil, err
	}
	return net.Dial("unix", clientSocket)
}

func ServerListen() (net.Listener, error) {
	oldUmask := syscall.Umask(0)
	listener, err := net.Listen("unix", wockSocket)
	syscall.Umask(oldUmask)
	return listener, err
}

func ClientListen(clientId string) (net.Listener, error) {
	oldUmask := syscall.Umask(0)
	listener, err := net.Listen("unix", fmt.Sprintf("%s-%s.sock", wockClientSocketPrefix, clientId))
	syscall.Umask(oldUmask)
	return listener, err
}

func IsServerPipeOpen() bool {
	l, err := net.Listen("unix", wockSocket)
	if err != nil {
		return true
	}
	err = l.Close()
	return err != nil
}
