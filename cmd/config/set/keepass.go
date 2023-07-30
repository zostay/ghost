package set

import (
	"github.com/spf13/cobra"
)

var (
	KeepassCmd = &cobra.Command{
		Use:    "keepass <keeper-name> [flags]",
		Short:  "Configure a Keepass secret keeper",
		Args:   cobra.MinimumNArgs(1),
		PreRun: PreRunSetKeepassKeeperConfig,
		Run:    RunSetKeeperConfig,
	}

	keepassPath string
)

func init() {
	KeepassCmd.Flags().StringVar(&keepassPath, "path", "", "Path to the KeePass database")
}

func PreRunSetKeepassKeeperConfig(cmd *cobra.Command, args []string) {
	Replacement.Keepass.Path = keepassPath
}
