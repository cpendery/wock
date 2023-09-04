//go:build !windows

package admin

import (
	"log"
	"log/slog"
	"os"
	"os/user"
)

const (
	unixRoot = "root"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

func RunAsElevated() {
	logger.Println("Please run `sudo wock start` to start the wock daemon as editing hosts files requires elevated permissions")
}

func IsAdmin() bool {
	user, err := user.Current()
	if err != nil {
		slog.Error("unable to check os user", slog.String("error", err.Error()))
		return false
	}
	return user.Username == unixRoot
}
