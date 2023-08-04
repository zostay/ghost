package cmd

import (
	"context"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var (
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a secret",
		Run:   RunGet,
	}

	id           string
	name         string
	flds         []string
	showPassword bool
)

func init() {
	getCmd.Flags().StringVar(&keeperName, "keeper", "", "The name of the secret keeper to use")
	getCmd.Flags().StringVar(&id, "id", "", "The ID of the secret to get")
	getCmd.Flags().StringVar(&name, "name", "", "The name of the secret to get")
	getCmd.Flags().StringSliceVar(&flds, "fields", []string{}, "The fields to display")
	getCmd.Flags().BoolVar(&showPassword, "show-password", false, "Show the password in the output")
}

func RunGet(cmd *cobra.Command, args []string) {
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

	sec, err := kpr.GetSecret(ctx, args[0])
	if err != nil {
		s.Logger.Panic(err)
	}

	s.PrintSecret(sec, showPassword, flds...)
}
