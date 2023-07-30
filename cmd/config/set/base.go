package set

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
)

var Replacement config.KeeperConfig

func RunSetKeeperConfig(cmd *cobra.Command, args []string) {
	keeperName := args[0]
	c := config.Instance()

	keeper, hasKeeper := c.Keepers[keeperName]
	var was config.KeeperType
	if hasKeeper {
		was = keeper.Type()
	}

	c.Keepers[keeperName] = &Replacement

	if hasKeeper && was != keeper.Type() {
		s.Logger.Panicf("Configuration failed. New keeper type %q does not match old type %q.", keeper.Type(), was)
	}

	err := c.Check()
	if err != nil {
		s.Logger.Panicf("Configuration failed. Configuration errors: %v", err)
	}

	err = c.Save(s.ConfigFile)
	if err != nil {
		s.Logger.Panicf("Configuration failed. Error saving configuration: %v", err)
	}
}
