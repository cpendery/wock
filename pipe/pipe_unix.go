//go:build !windows

package pipe

import (
	"context"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
)

const (
	wockSocket             = "/tmp/wock"
	wockClientSocketPrefix = "/tmp/wock"
)

func DialServer(timeout *time.Duration) (net.Conn, error) {
	dialTimeout := defaultTimeout
	if timeout != nil {
		dialTimeout = *timeout
	}
	ctx, cancelCtx := context.WithDeadline(context.Background(), time.Now().Add(dialTimeout))
	defer cancelCtx()
	if err := waitForFile(wockSocket, ctx); err != nil {
		return nil, err
	}
	return net.Dial("unix", wockSocket)
}

func DialClient(clientId string) (net.Conn, error) {
	ctx, cancelCtx := context.WithDeadline(context.Background(), time.Now().Add(defaultTimeout))
	defer cancelCtx()
	clientSocket := fmt.Sprintf("%s-%s", wockClientSocketPrefix, clientId)
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
	listener, err := net.Listen("unix", fmt.Sprintf("%s-%s", wockClientSocketPrefix, clientId))
	syscall.Umask(oldUmask)
	return listener, err
}

func Teardown() error {
	return os.Remove(wockSocket)
}

func IsServerPipeOpen() bool {
	conn, err := DialServer(nil)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
