package set

import (
	"context"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
)

var Replacement config.KeeperConfig

func RunSetKeeperConfig(cmd *cobra.Command, args []string) {
	keeperName := args[0]
	c := config.Instance()

	kc := c.Keepers[keeperName]
	var was string
	if kc != nil {
		was := plugin.Type(kc)
		if was == "" {
			s.Logger.Panicf("Configuration failed. Keeper %q has no type.", keeperName)
		}
	}

	c.Keepers[keeperName] = Replacement

	if kc != nil && was != plugin.Type(kc) {
		s.Logger.Panicf("Configuration failed. New kc type %q does not match old type %q.", plugin.Type(kc), was)
	}

	err := keeper.CheckConfig(context.Background(), c)
	if err != nil {
		s.Logger.Panicf("Configuration failed. Configuration errors: %v", err)
	}

	err = c.Save(s.ConfigFile)
	if err != nil {
		s.Logger.Panicf("Configuration failed. Error saving configuration: %v", err)
	}
}
