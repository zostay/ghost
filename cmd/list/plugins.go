package list

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/plugin"
)

var PluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "List secret keeper plugin types",
	Args:  cobra.NoArgs,
	Run:   RunListPlugins,
}

func RunListPlugins(_ *cobra.Command, _ []string) {
	for _, kt := range plugin.List() {
		s.Printer.Print(kt)
	}
}
