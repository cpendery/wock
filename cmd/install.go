package cmd

import (
	"github.com/cpendery/wock/cert"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install wock's local certificate authority",
	Args:  cobra.ExactArgs(0),
	RunE:  runInstallCmd,
}

func runInstallCmd(_ *cobra.Command, _ []string) error {
	return cert.Install()
}
