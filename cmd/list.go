package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/cmd/list"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
}

func init() {
	listCmd.AddCommand(
		list.PluginsCmd,
		list.LocationsCmd,
		list.SecretsCmd,
	)
}
