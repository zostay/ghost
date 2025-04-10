package list

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var LocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List the locations that may contain secrets",
	Args:  cobra.NoArgs,
	Run:   RunListLocations,
}

func init() {
	LocationsCmd.Flags().StringVar(&keeperName, "keeper", "", "The name of the secret keeper to use")
}

func RunListLocations(cmd *cobra.Command, _ []string) {
	c := config.Instance()
	if keeperName == "" {
		keeperName = c.MasterKeeper
	}

	if keeperName == "" {
		s.Logger.Panic("No keeper specified.")
	}

	if _, hasConfig := c.Keepers[keeperName]; !hasConfig {
		s.Logger.Panicf("No keeper named %q.", keeperName)
	}

	ctx := keeper.WithBuilder(cmd.Context(), c)
	kpr, err := keeper.Build(ctx, keeperName)
	if err != nil {
		s.Logger.Panic(err)
	}

	locs, err := kpr.ListLocations(ctx)
	if err != nil {
		s.Logger.Panic(err)
	}

	for _, loc := range locs {
		s.Printer.Print(loc)
	}
}
