//go:build !windows

package pipe

import "net"

const wockSocket = "/tmp/wock.sock"

func Dial() (net.Conn, error) {
	return net.Dial("unix", wockSocket)
}
