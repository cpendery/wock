//go:build !windows

package admin

import (
	"log"
	"log/slog"
	"os/user"
)

const (
	unixRoot = "root"
)

func RunAsElevated() {
	log.Println("Please run `sudo wock start` to start the wock daemon as editing hosts files requires elevated permissions")
}

func IsAdmin() bool {
	user, err := user.Current()
	if err != nil {
		slog.Error("unable to check os user", slog.String("error", err.Error()))
		return false
	}
	return user.Username == unixRoot
}
