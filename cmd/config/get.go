package config

import (
	"bytes"
	"context"
	"strings"

	"github.com/pborman/indent"
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a secret keeper configuration",
	Args:  cobra.ExactArgs(1),
	Run:   RunGet,
}

func RunGet(cmd *cobra.Command, args []string) {
	keeperName := args[0]
	c := config.Instance()
	ctx := keeper.WithBuilder(cmd.Context(), c)
	kpr, hasKeeper := c.Keepers[keeperName]
	if !hasKeeper {
		s.Logger.Panicf("Keeper %q is not configured.", keeperName)
	}

	PrintKeeper(ctx, keeperName, kpr, 0)
}

func makeIndent(i int) string {
	return strings.Repeat(" ", i*2)
}

func PrintKeeper(
	ctx context.Context,
	keeperName string,
	kc config.KeeperConfig,
	i int,
) {
	sp := makeIndent(i)
	dc, err := keeper.DecodePartial(ctx, keeperName)
	if err != nil {
		s.Printer.Printf("%sERROR: failed to decode configuration for keeper %q: %v", sp, keeperName, err)
	}
	typ := plugin.Type(kc)
	s.Printer.Printf("%stype: %s", sp, plugin.Type(kc))
	r, hasKeeper := plugin.Get(typ)
	if !hasKeeper {
		s.Printer.Printf("%sERROR: unknown keeper type", sp)
	}

	buf := &bytes.Buffer{}
	w := indent.New(buf, sp)
	err = r.Printer(dc, w)
	if err != nil {
		s.Printer.Printf("%sERROR: failed to print configuration for keeper %q: %v", sp, keeperName, err)
	}

	s.Printer.Print(buf.String())
}
