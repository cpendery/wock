//go:build !windows

package admin

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"time"

	"github.com/cpendery/wock/pipe"
)

const (
	unixRoot = "root"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

func RunAsElevated() {
	cmd := exec.Command("sudo", os.Args...)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		return
	}
	timeout := 30 * time.Second
	conn, err := pipe.DialServer(&timeout)
	if err != nil {
		fmt.Println(err)
		if cmd.Cancel != nil {
			defer cmd.Cancel()
		}
	} else {
		defer conn.Close()
	}
}

func IsAdmin() bool {
	user, err := user.Current()
	if err != nil {
		slog.Error("unable to check os user", slog.String("error", err.Error()))
		return false
	}
	return user.Username == unixRoot
}
