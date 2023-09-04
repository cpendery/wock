package cmd

import (
	"errors"
	"fmt"

	"github.com/cpendery/wock/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stops the wock daemon",
	Args:  cobra.ExactArgs(0),
	RunE:  runStopCommand,
}

func runStopCommand(_ *cobra.Command, _ []string) error {
	c, err := client.NewClient()
	if err != nil && errors.Is(err, client.ErrUnableToDialDaemon) {
		logger.Println("Daemon is already offline")
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()
	if err := c.Stop(); err != nil {
		return err
	}
	logger.Println("Successfully stopped daemon")
	return nil
}
