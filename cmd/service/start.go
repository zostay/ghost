package service

import (
	"context"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the ghost service",
	Run:   RunStartService,
}

func RunStartService(cmd *cobra.Command, args []string) {
	c := config.Instance()
	if c.MasterKeeper == "" {
		s.Logger.Panic("no master keeper set")
	}

	ctx := context.Background()

	kpr, err := keeper.Build(ctx, c.MasterKeeper, c)
	if err != nil {
		s.Logger.Panicf("failed to configure master keeper %q: %v", c.MasterKeeper, err)
	}

	err = keeper.StartServer(kpr)
	if err != nil {
		s.Logger.Panic(err)
	}
}
