package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/cpendery/wock/admin"
	"github.com/cpendery/wock/client"
	"github.com/cpendery/wock/daemon"
	"github.com/cpendery/wock/pipe"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   `wock [domain] [directory]`,
		Short: "mock web hosts",
		Long: `wock - mock the web 

wock is a tool for mocking a host/domain and serving all traffic
that host locally via http/https.

complete documentation is available at https://github.com/cpendery/wock`,
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE:         rootExec,
	}
)

func rootExec(cmd *cobra.Command, args []string) error {
	daemonRunning := pipe.IsOpen()
	if !daemonRunning && !admin.IsAdmin() {
		admin.RunAsElevated()
	}
	if !daemonRunning && admin.IsAdmin() {
		daemon.NewDaemon().Start()
	}
	if !admin.IsAdmin() {
		time.Sleep(1 * time.Second)
		c, err := client.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		defer c.Close()
		host := args[0]
		err = c.Mock(host)
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
