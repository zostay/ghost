package config

import (
	"context"
	"sort"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/maps"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secret keeper configurations",
	Args:  cobra.NoArgs,
	Run:   RunListConfig,
}

func RunListConfig(cmd *cobra.Command, args []string) {
	c := config.Instance()
	ctx := keeper.WithBuilder(context.Background(), c)

	keys := maps.Keys(c.Keepers)
	sort.Strings(keys)

	for _, keeperName := range keys {
		kpr := c.Keepers[keeperName]
		s.Printer.Print(keeperName)
		PrintKeeper(ctx, keeperName, kpr, 1)
	}
}
