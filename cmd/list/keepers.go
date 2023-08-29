package list

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/plugin"
)

var KeepersCmd = &cobra.Command{
	Use:   "keepers",
	Short: "List secret keeper configurations",
	Args:  cobra.NoArgs,
	Run:   RunListKeepers,
}

func RunListKeepers(cmd *cobra.Command, args []string) {
	for _, kt := range plugin.List() {
		s.Printer.Print(kt)
	}
}
