package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync <from> <to>",
		Short: "Synchronize secrets between secret keepers",
		Long: ` The synchronize routine will copy all secrets from one secret keeper to 
another. This is done by using the name, username, and location as a unique key 
for each secret. If more than one secret exists with the same key in the 
original, this operation will fail to synchronize unless 
the --ignore-duplicates option is given. When given, the most recent of the 
duplicates will be transferred.

 Normally, this operation only adds secrets to the destination. Using
the --delete option, however, will cause any secret found in the destination 
not matching one in the source to be deleted.

 Please note that sync, especially with a large LastPass database can take 
several minutes or even hours due to API rate limits.`,
		Args: cobra.ExactArgs(2),
		Run:  RunSync,
	}

	alsoDelete      bool
	ignoreDuplicate bool
)

func init() {
	syncCmd.Flags().BoolVar(&alsoDelete, "delete", false, "Delete secrets from the destination keeper")
	syncCmd.Flags().BoolVar(&ignoreDuplicate, "ignore-duplicates", false, "When synchronizing, ignore duplicates (keep latest by last-modified date)")
}

func RunSync(cmd *cobra.Command, args []string) {
	fromKeeper := args[0]
	toKeeper := args[1]

	c := config.Instance()
	ctx := keeper.WithBuilder(context.Background(), c)
	fromKpr, err := keeper.Build(ctx, fromKeeper)
	if err != nil {
		s.Logger.Panic(err)
		return
	}

	toKpr, err := keeper.Build(ctx, toKeeper)
	if err != nil {
		s.Logger.Panic(err)
		return
	}

	syncer, err := keeper.NewSync()
	if err != nil {
		s.Logger.Panic(err)
		return
	}

	err = syncer.AddSecretKeeper(ctx, fromKpr, ignoreDuplicate)
	if err != nil {
		if errors.Is(err, keeper.ErrDuplicate) {
			s.Logger.Panic("The source secret keeper contains secrets with duplicate name, username, and location. Either de-duplicate or use --ignore-duplicates.")
			return
		}
		s.Logger.Panic(err)
		return
	}

	err = syncer.CopyTo(ctx, toKpr)
	if err != nil {
		s.Logger.Panic(err)
		return
	}

	if alsoDelete {
		err = syncer.DeleteAbsent(ctx, toKpr)
		if err != nil {
			s.Logger.Panic(err)
			return
		}
	}
}
