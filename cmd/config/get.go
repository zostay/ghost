package config

import (
	"strings"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
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
	keeper, hasKeeper := c.Keepers[keeperName]
	if !hasKeeper {
		s.Logger.Panicf("Keeper %q is not configured.", keeperName)
	}

	PrintKeeper(keeper, 0)
}

func PrintKeeper(keeper *config.KeeperConfig, i int) {
	indent := strings.Repeat(" ", i*2)
	switch keeper.Type() {
	case config.KTKeepass:
		s.Logger.Printf("%stype: keepass", indent)
		s.Logger.Printf("%spath: %s", indent, keeper.Keepass.Path)
	case config.KTLastPass:
		s.Logger.Printf("%stype: lastpass", indent)
		s.Logger.Printf("%susername: %s", indent, keeper.LastPass.Username)
	case config.KTLowSecurity:
		s.Logger.Printf("%stype: low", indent)
		s.Logger.Printf("%spath: %s", indent, keeper.Low.Path)
	case config.KTGRPC:
		s.Logger.Printf("%stype: grpc", indent)
		s.Logger.Printf("%slistener: %s", indent, keeper.GRPC.Listener)
	case config.KTKeyring:
		s.Logger.Printf("%stype: keyring", indent)
		s.Logger.Printf("%skeyring: %s", indent, keeper.Keyring.ServiceName)
	case config.KTMemory:
		s.Logger.Printf("%stype: memory", indent)
	case config.KTRouter:
		s.Logger.Printf("%stype: router", indent)
		s.Logger.Printf("%sdefault route: %s", indent, keeper.Router.DefaultRoute)
		s.Logger.Printf("%sroutes:", indent)
		for _, route := range keeper.Router.Routes {
			s.Logger.Printf("%s - locations: %s", indent, strings.Join(route.Locations, ", "))
			s.Logger.Printf("%s   keeper: %s", indent, route.Keeper)
		}
	case config.KTSeq:
		s.Logger.Printf("%stype: seq", indent)
		s.Logger.Printf("%skeepers:", indent)
		for _, keeper := range keeper.Seq.Keepers {
			s.Logger.Printf("%s  - %s", indent, keeper)
		}
	}
}
