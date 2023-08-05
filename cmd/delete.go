package cmd

import (
	"context"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret",
	Args:  cobra.NoArgs,
	Run:   RunDelete,
}

func init() {
	getCmd.Flags().StringVar(&keeperName, "keeper", "", "The name of the secret keeper to use")
	getCmd.Flags().StringVar(&id, "id", "", "The ID of the secret to get")
	getCmd.Flags().StringVar(&name, "name", "", "The name of the secret to get")
}

func RunDelete(cmd *cobra.Command, args []string) {
	if name != "" && id != "" {
		s.Logger.Panic("Cannot specify both --id and --name.")
	}

	if name == "" && id == "" {
		s.Logger.Panic("Must specify either --id or --name.")
	}

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

	ctx := context.Background()
	kpr, err := keeper.Build(ctx, keeperName, c)
	if err != nil {
		s.Logger.Panic(err)
	}

	if name != "" {
		secs, err := kpr.GetSecretsByName(ctx, name)
		if err != nil {
			s.Logger.Panic(err)
		}

		switch len(secs) {
		case 0:
			s.Logger.Panicf("No secret named %q.", name)
		case 1:
			id = secs[0].ID()
		default:
			s.Logger.Panicf("Multiple secrets named %q. Please delete by --id", name)
		}
	}

	err = kpr.DeleteSecret(ctx, id)
	if err != nil {
		s.Logger.Panic(err)
	}
}
