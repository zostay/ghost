package config

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secret keeper configurations",
	Args:  cobra.NoArgs,
	Run:   RunListConfig,
}

func RunListConfig(cmd *cobra.Command, args []string) {
	c := config.Instance()

	for name, kc := range c.Keepers {
		s.Logger.Print(name)
		PrintKeeper(kc, 1)
	}
}
