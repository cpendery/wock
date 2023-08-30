package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cpendery/wock/admin"
	"github.com/cpendery/wock/client"
	"github.com/cpendery/wock/config"
	"github.com/cpendery/wock/daemon"
	"github.com/cpendery/wock/hosts"
	"github.com/cpendery/wock/pipe"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: `wock [domain] [directory] [flags]
  wock [alias]`,
		Short: "mock web hosts",
		Long: `wock - mock the web 

wock is a tool for mocking a host/domain and serving all traffic
that host locally via http/https.

complete documentation is available at https://github.com/cpendery/wock`,
		Args: func(_ *cobra.Command, args []string) error {
			switch len(args) {
			case 0:
				return errors.New("requires at least one arg")
			case 1:
				alias := strings.ToLower(args[0])
				if !config.IsValidAlias(alias) {
					return fmt.Errorf("unknown alias %s", alias)
				}
				return nil
			case 2:
				host := strings.ToLower(args[0])
				dir := strings.ToLower(args[1])
				if !hosts.IsValidHostname(host) {
					return fmt.Errorf("provided host '%s' is an invalid hostname", host)
				}
				if _, err := config.IsValidDirectory(dir); err != nil {
					return err
				}
			default:
				return errors.New("invalid args")
			}
			return nil
		},
		SilenceUsage: true,
		RunE:         rootExec,
	}
)

func startDaemon() {
	daemonRunning := pipe.IsOpen()
	if !daemonRunning && !admin.IsAdmin() {
		admin.RunAsElevated()
	}
	if !daemonRunning && admin.IsAdmin() {
		daemon.NewDaemon().Start()
	}
}

func rootExec(cmd *cobra.Command, args []string) error {
	startDaemon()
	time.Sleep(1 * time.Second)
	daemonRunning := pipe.IsOpen()
	if !admin.IsAdmin() && daemonRunning {
		c, err := client.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		defer c.Close()
		var host, dir string
		if len(args) == 1 {
			host, dir = config.GetAlias(args[0])
		} else {
			host = args[0]
			dir = args[1]
		}
		absDir, _ := config.IsValidDirectory(dir)
		err = c.Mock(host, *absDir)
		if err != nil {
			return fmt.Errorf("failed to mock host %s: %w", host, err)
		}
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1) // skipcq: RVV-A0003
	}
}
