package cmd

import (
	"fmt"
	"time"

	"github.com/cpendery/wock/admin"
	"github.com/cpendery/wock/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clearCmd)
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clears all wocked hosts",
	Args:  cobra.ExactArgs(0),
	RunE:  runClearCommand,
}

func runClearCommand(_ *cobra.Command, _ []string) error {
	startDaemon()
	if !admin.IsAdmin() {
		time.Sleep(1 * time.Second)
		c, err := client.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		defer c.Close()
		return c.Clear()
	}
	return nil
}
