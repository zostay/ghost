package cmd

import (
	_ "embed"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
)

//go:embed version.txt
var Version string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Args:  cobra.NoArgs,
	Run:   RunVersion,
}

func RunVersion(*cobra.Command, []string) {
	s.Printer.Println("ghost v" + Version)
}
