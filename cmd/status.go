package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cpendery/wock/client"
	"github.com/cpendery/wock/model"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "check the current status of the wock daemon",
	Args:  cobra.ExactArgs(0),
	Run:   runStatusCmd,
}

func printDaemonStatus(hosts *[]model.MockedHost) {
	if hosts == nil {
		fmt.Print("\n")
		fmt.Printf("wock daemon [%s]\n", color.RedString("offline"))
	} else {
		fmt.Print("\n")
		fmt.Printf("wock daemon [%s]\n", color.GreenString("online"))
		fmt.Print("\n")
		data := [][]string{}
		for _, host := range *hosts {
			data = append(data, []string{host.Host, host.Directory})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Mocked Host", "Directory Served"})
		for _, v := range data {
			table.Append(v)
		}
		table.SetBorder(false)
		table.Render()
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
