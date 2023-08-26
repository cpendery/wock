package cmd

import (
	"fmt"
	"log/slog"

	"github.com/cpendery/wock/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
}

var daemonCmd = &cobra.Command{
	Use:   "status",
	Short: "check the current status of the wock daemon",
	Args:  cobra.ExactArgs(0),
	Run:   runStatusCmd,
}

func printDaemonStatus(hosts *[]string) {
	if hosts == nil {
		fmt.Print("\n")
		fmt.Printf("wock daemon [%s]\n", color.RedString("offline"))
	} else {
		fmt.Print("\n")
		fmt.Printf("wock daemon [%s]\n", color.GreenString("online"))
		fmt.Print("\n")
		fmt.Println("mocked hosts:")
		for _, host := range *hosts {
			fmt.Printf("> %s\n", host)
		}
	}
}

func runStatusCmd(_ *cobra.Command, _ []string) {
	defer func() {
		if r := recover(); r != nil {
			slog.Debug("recovered from panic in status command")
		}
	}()
	c, err := client.NewClient()
	if err != nil {
		slog.Debug("failed to create client", slog.String("error", err.Error()))
		printDaemonStatus(nil)
		return
	}
	mockedHosts, err := c.CheckStatus()
	if err != nil {
		slog.Debug("failed to check daemon status", slog.String("error", err.Error()))
		printDaemonStatus(nil)
		return
	}
	printDaemonStatus(mockedHosts)
}
