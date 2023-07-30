package config

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret keeper configuration",
	Args:  cobra.ExactArgs(1),
	Run:   RunDeleteConfig,
}

func RunDeleteConfig(cmd *cobra.Command, args []string) {
	keeperName := args[0]
	c := config.Instance()

	delete(c.Keepers, keeperName)

	err := c.Save(s.ConfigFile)
	if err != nil {
		s.Logger.Panicf("Configuration failed. Error saving configuration: %v", err)
	}
}
