package cmd

import (
	"log"
	"log/slog"

	"github.com/cpendery/wock/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rmCmd)
}

var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove a currently wocked host",
	Args:  cobra.ExactArgs(1),
	RunE:  runRmCmd,
}

func runRmCmd(_ *cobra.Command, args []string) error {
	c, err := client.NewClient()
	if err != nil {
		log.Println("Daemon is offline, no hosts to remove")
		return nil
	}
	host := args[0]

	if err := c.Remove(host); err != nil {
		slog.Debug("failed to remove host", slog.String("error", err.Error()), slog.String("host", host))
		return err
	}
	return nil
}
