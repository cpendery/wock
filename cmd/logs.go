package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cpendery/wock/daemon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logsCmd)
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "prints daemon's logs to stdout",
	Args:  cobra.ExactArgs(0),
	RunE:  runLogsCmd,
}

func runLogsCmd(_ *cobra.Command, _ []string) error {
	f, err := os.Open(daemon.WockDaemonLogFile)
	if err != nil {
		return fmt.Errorf("unable to read daemon logs: %w", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	if _, err := r.WriteTo(os.Stdout); err != nil {
		return fmt.Errorf("unable to write daemon logs: %w", err)
	}
	return nil
}
