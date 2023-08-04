package list

import (
	"context"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var LocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List the locations that may contain secrets",
	Args:  cobra.ExactArgs(1),
	Run:   RunListLocations,
}

func RunListLocations(cmd *cobra.Command, args []string) {
	keeperName := args[0]
	c := config.Instance()

	if _, hasConfig := c.Keepers[keeperName]; !hasConfig {
		s.Logger.Panicf("No keeper named %q.", keeperName)
	}

	ctx := context.Background()
	kpr, err := keeper.Build(ctx, keeperName, c)
	if err != nil {
		s.Logger.Panic(err)
	}

	locs, err := kpr.ListLocations(ctx)
	if err != nil {
		s.Logger.Panic(err)
	}

	for _, loc := range locs {
		s.Logger.Print(loc)
	}
}
