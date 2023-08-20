package config

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/cmd/config/set"
)

var SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Add or update a secret keeper configuration",
}

func init() {
	SetCmd.AddCommand(
		set.KeepassCmd,
		set.LastPassCmd,
		set.LowSecurityCmd,
		set.GRPCCmd,
		set.KeyringCmd,
		set.MemoryCmd,
		set.HumanCmd,

		set.PolicyCmd,
		set.RouterCmd,
		set.SeqCmd,
	)
}
