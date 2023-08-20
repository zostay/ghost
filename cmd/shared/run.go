package shared

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

func RunRoot(cmd *cobra.Command, args []string) {
	Logger = log.New(cmd.OutOrStderr(), "", 0)

	var err error
	err = config.Instance().Load(ConfigFile)
	if err != nil {
		Logger.Panicf("Failure to load configuration: %v", err)
	}

	cfg := config.Instance()
	ctx := keeper.WithBuilder(context.Background(), cfg)
	err = keeper.CheckConfig(ctx, cfg)
	if err != nil {
		Logger.Panicf("Configuration errors: %v", err)
	}
}
