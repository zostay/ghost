package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/cmd/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the ghost configuration",
}

func init() {
	configCmd.AddCommand(config.KeepassCmd)
	configCmd.AddCommand(config.LastPassCmd)
	configCmd.AddCommand(config.LowCmd)
	configCmd.AddCommand(config.RouterCmd)
	configCmd.AddCommand(config.SeqCmd)
}
