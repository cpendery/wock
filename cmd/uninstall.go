package cmd

import (
	"github.com/cpendery/wock/cert"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall wock's local certificate authority",
	Args:  cobra.ExactArgs(0),
	RunE:  runUninstallCmd,
}

func runUninstallCmd(_ *cobra.Command, _ []string) error {
	return cert.Uninstall()
}
