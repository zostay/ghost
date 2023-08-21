package config

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets/http"
	"github.com/zostay/ghost/pkg/secrets/human"
	"github.com/zostay/ghost/pkg/secrets/keepass"
	"github.com/zostay/ghost/pkg/secrets/keyring"
	"github.com/zostay/ghost/pkg/secrets/lastpass"
	"github.com/zostay/ghost/pkg/secrets/low"
	"github.com/zostay/ghost/pkg/secrets/memory"
	"github.com/zostay/ghost/pkg/secrets/policy"
	"github.com/zostay/ghost/pkg/secrets/router"
	"github.com/zostay/ghost/pkg/secrets/seq"
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
	ctx := keeper.WithBuilder(context.Background(), c)
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
	// TODO this should be rolled into plugin configuration too
	indent := makeIndent(i)
	dc, err := keeper.Decode(ctx, keeperName)
	if err != nil {
		s.Logger.Panicf("failed to decode configuration for keeper %q: %v", keeperName, err)
	}
	s.Logger.Printf("%stype: %s", indent, plugin.Type(kc))
	switch plugin.Type(kc) {
	case keepass.ConfigType:
		kpc := dc.(*keepass.Config)
		s.Logger.Printf("%spath: %s", indent, kpc.Path)
		s.Logger.Printf("%smaster: <hidden>", indent)
	case lastpass.ConfigType:
		lpc := dc.(*lastpass.Config)
		s.Logger.Printf("%susername: %s", indent, lpc.Username)
		s.Logger.Printf("%spassword: <hidden>", indent)
	case low.ConfigType:
		lc := dc.(*low.Config)
		s.Logger.Printf("%spath: %s", indent, lc.Path)
	case http.ConfigType:
	case keyring.ConfigType:
		krc := dc.(*keyring.Config)
		s.Logger.Printf("%skeyring: %s", indent, krc.ServiceName)
	case memory.ConfigType:
	case human.ConfigType:
		hc := dc.(*human.Config)
		s.Logger.Printf("%squestions:", indent)
		for _, q := range hc.Questions {
			indentP1 := makeIndent(i + 1)
			indentP2 := makeIndent(i + 2)
			s.Logger.Printf("%sid: %s", indentP1, q.ID)
			s.Logger.Printf("%spresets:", indentP1)
			for k, v := range q.Presets {
				s.Logger.Printf("%s%s = %s", indentP2, k, v)
			}
			s.Logger.Printf("%sasking for:", indentP1)
			for _, k := range q.AskFor {
				s.Logger.Printf("%s%s", indentP2, k)
			}
		}
	case policy.ConfigType:
		pc := dc.(*policy.Config)
		s.Logger.Printf("%skeeper: %s", indent, pc.Keeper)
		s.Logger.Printf("%sdefault rule:", indent)
		s.Logger.Printf("%s  acceptance: %s", indent, pc.DefaultRule.Acceptance)
		if pc.DefaultRule.Lifetime >= 0 {
			s.Logger.Printf("%s  lifetime: %v", indent, pc.DefaultRule.Lifetime)
		}
		s.Logger.Printf("%srules:", indent)
		for _, r := range pc.Rules {
			if policy.ValidAcceptance(r.Acceptance, false) {
				s.Logger.Printf("%s - acceptance: %s", indent, r.Acceptance)
			}
			if r.Lifetime > 0 {
				s.Logger.Printf("%s - lifetime: %v", indent, r.Lifetime)
			}
		}
	case router.ConfigType:
		rc := dc.(*router.Config)
		s.Logger.Printf("%sdefault route: %s", indent, rc.DefaultRoute)
		s.Logger.Printf("%sroutes:", indent)
		for _, route := range rc.Routes {
			s.Logger.Printf("%s - locations: %s", indent, strings.Join(route.Locations, ", "))
			s.Logger.Printf("%s   keeper: %s", indent, route.Keeper)
		}
	case seq.ConfigType:
		sc := dc.(*seq.Config)
		s.Logger.Printf("%skeepers:", indent)
		for _, name := range sc.Keepers {
			s.Logger.Printf("%s  - %s", indent, name)
		}
	default:
		s.Logger.Printf("%sERROR: unknown keeper type", indent)
	}
}
