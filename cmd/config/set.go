package config

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/cmd/config/set"
	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
)

var SetCmd = &cobra.Command{
	Use:              "set",
	Short:            "Add or update a secret keeper configuration",
	PersistentPreRun: PreRunSet,
}

func init() {
	SetCmd.AddCommand(set.KeepassCmd)
	SetCmd.AddCommand(set.LastPassCmd)
	SetCmd.AddCommand(set.LowSecurityCmd)
	SetCmd.AddCommand(set.RouterCmd)
	SetCmd.AddCommand(set.SeqCmd)
}

func PreRunSet(cmd *cobra.Command, args []string) {
	s.RunRoot(cmd, args)

	keeperName := args[0]
	c := config.Instance()

	if keeper, hasKeeper := c.Keepers[keeperName]; hasKeeper {
		set.Replacement = *keeper
	}
}
