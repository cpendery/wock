//go:build windows

package pipe

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	winio "github.com/Microsoft/go-winio"
)

const (
	wockServerPipe = `\\.\pipe\wock`
)

func DialServer(timeout *time.Duration) (net.Conn, error) {
	dialTimeout := defaultTimeout
	if timeout != nil {
		dialTimeout = *timeout
	}
	ctx, cancelCtx := context.WithDeadline(context.Background(), time.Now().Add(dialTimeout))
	defer cancelCtx()
	if err := waitForFile(wockServerPipe, ctx); err != nil {
		return nil, err
	}
	return winio.DialPipeContext(ctx, wockServerPipe)
}

func DialClient(clientId string) (net.Conn, error) {
	ctx, cancelCtx := context.WithDeadline(context.Background(), time.Now().Add(defaultTimeout))
	defer cancelCtx()

	clientPipe := fmt.Sprintf("%s-%s", wockServerPipe, clientId)
	if err := waitForFile(wockServerPipe, ctx); err != nil {
		return nil, err
	}
	return winio.DialPipeContext(ctx, clientPipe)
}

func ServerListen() (net.Listener, error) {
	return winio.ListenPipe(wockServerPipe, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;AU)"})
}

func ClientListen(clientId string) (net.Listener, error) {
	return winio.ListenPipe(fmt.Sprintf("%s-%s", wockServerPipe, clientId), &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;AU)"})
}

func Teardown() error {
	return os.Remove(wockServerPipe)
}

func IsServerPipeOpen() bool {
	l, err := winio.ListenPipe(wockServerPipe, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;AU)"})
	if err != nil {
		return true
	}
	err = l.Close()
	return err != nil
}
