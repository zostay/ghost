package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/cmd/config"
)

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage the ghost configuration",
	}
)

func init() {
	configCmd.AddCommand(config.DeleteCmd)
	configCmd.AddCommand(config.GetCmd)
	configCmd.AddCommand(config.ListCmd)
	configCmd.AddCommand(config.SetCmd)
}
