package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/set"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync <from> <to>",
		Short: "Synchronize secrets between secret keepers",
		Args:  cobra.ExactArgs(2),
		Run:   RunSync,
	}

	alsoDelete bool
)

func init() {
	syncCmd.Flags().BoolVar(&alsoDelete, "delete", false, "Delete secrets from the destination keeper")
}

func RunSync(cmd *cobra.Command, args []string) {
	fromKeeper := args[0]
	toKeeper := args[1]

	ctx := context.Background()
	c := config.Instance()

	fromKpr, err := keeper.Build(ctx, fromKeeper, c)
	if err != nil {
		s.Logger.Panic(err)
	}

	toKpr, err := keeper.Build(ctx, toKeeper, c)
	if err != nil {
		s.Logger.Panic(err)
	}

	fromLocs, err := fromKpr.ListLocations(ctx)
	if err != nil {
		s.Logger.Panic(err)
	}

	toLocs, err := toKpr.ListLocations(ctx)
	if err != nil {
		s.Logger.Panic(err)
	}

	fromLocSet := set.New(fromLocs...)
	toLocSet := set.New(toLocs...)

	commonLocs, toAddLocs, toDelLocs := set.Diff(fromLocSet, toLocSet)
	upsertLocs := set.Union(commonLocs, toAddLocs)
	for _, loc := range upsertLocs.Keys() {
		fromIds, err := fromKpr.ListSecrets(ctx, loc)
		if err != nil {
			s.Logger.Panic(err)
		}

		toIds, err := toKpr.ListSecrets(ctx, loc)
		if err != nil {
			s.Logger.Panic(err)
		}

		fromIdSet := set.New(fromIds...)
		toIdSet := set.New(toIds...)

		commonIds, toAddIds, toDelIds := set.Diff(fromIdSet, toIdSet)
		upsertIds := set.Union(commonIds, toAddIds)
		for _, id := range upsertIds.Keys() {
			sec, err := fromKpr.GetSecret(ctx, id)
			if err != nil {
				s.Logger.Panic(err)
			}

			_, err = toKpr.SetSecret(ctx, sec)
			if err != nil {
				s.Logger.Panic(err)
			}
		}

		if alsoDelete {
			for _, id := range toDelIds.Keys() {
				err := toKpr.DeleteSecret(ctx, id)
				if err != nil {
					s.Logger.Panic(err)
				}
			}
		}
	}

	if alsoDelete {
		for _, loc := range toDelLocs.Keys() {
			toIds, err := toKpr.ListSecrets(ctx, loc)
			if err != nil {
				s.Logger.Panic(err)
			}

			for _, id := range toIds {
				err := toKpr.DeleteSecret(ctx, id)
				if err != nil {
					s.Logger.Panic(err)
				}
			}
		}
	}
}
