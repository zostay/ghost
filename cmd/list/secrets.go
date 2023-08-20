package list

import (
	"context"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var (
	SecretsCmd = &cobra.Command{
		Use:   "secrets",
		Short: "List secrets",
		Args:  cobra.NoArgs,
		Run:   RunListSecrets,
	}

	location     string
	flds         []string
	showPassword bool
)

func init() {
	SecretsCmd.Flags().StringVar(&keeperName, "keeper", "", "The name of the secret keeper to use")
	SecretsCmd.Flags().StringVar(&location, "location", "", "The location to list secrets from")
	SecretsCmd.Flags().StringSliceVar(&flds, "fields", []string{}, "The fields to display")
	SecretsCmd.Flags().BoolVar(&showPassword, "show-password", false, "Show the password in the output")
}

func RunListSecrets(cmd *cobra.Command, args []string) {
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

	ctx := keeper.WithBuilder(context.Background(), c)
	kpr, err := keeper.Build(ctx, keeperName)
	if err != nil {
		s.Logger.Panic(err)
	}

	secs, err := kpr.ListSecrets(ctx, location)
	if err != nil {
		s.Logger.Panic(err)
	}

	for _, id := range secs {
		sec, err := kpr.GetSecret(ctx, id)
		if err != nil {
			s.Logger.Panic(err)
		}
		s.PrintSecret(sec, showPassword, flds...)
	}
}
