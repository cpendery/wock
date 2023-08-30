//go:build windows

package pipe

import (
	"context"
	"fmt"
	"net"
	"time"

	winio "github.com/Microsoft/go-winio"
)

const (
	wockServerPipe = `\\.\pipe\wock`
)

func DialServer() (net.Conn, error) {
	ctx, cancelCtx := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancelCtx()
	if err := waitForFile(wockServerPipe, ctx); err != nil {
		return nil, err
	}
	return winio.DialPipeContext(ctx, wockServerPipe)
}

func DialClient(clientId string) (net.Conn, error) {
	ctx, cancelCtx := context.WithDeadline(context.Background(), time.Now().Add(timeout))
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

func IsServerPipeOpen() bool {
	l, err := winio.ListenPipe(wockServerPipe, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;AU)"})
	if err != nil {
		return true
	}
	err = l.Close()
	return err != nil
}
