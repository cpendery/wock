package cmd

import (
	"errors"
	"fmt"

	"github.com/cpendery/wock/cert"
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
	if !cert.IsInstalled() {
		return errors.New("local CA is not installed, run `wock install` to install the CA")
	}

	startDaemon()
	c, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer c.Close()
	if err := c.Clear(); err != nil {
		return err
	}
	logger.Println("Successfully cleared all hosts")
	return nil
}
