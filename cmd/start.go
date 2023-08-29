package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the wock daemon",
	Args:  cobra.ExactArgs(0),
	Run:   runStartCommand,
}

func runStartCommand(_ *cobra.Command, _ []string) {
	startDaemon()
}
