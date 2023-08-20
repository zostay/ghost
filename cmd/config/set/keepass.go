package set

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/secrets/keepass"
)

var (
	KeepassCmd = &cobra.Command{
		Use:    "keepass <keeper-name> [flags]",
		Short:  "Configure a Keepass secret keeper",
		Args:   cobra.ExactArgs(1),
		PreRun: PreRunSetKeepassKeeperConfig,
		Run:    RunSetKeeperConfig,
	}

	keepassPath string
)

func init() {
	KeepassCmd.Flags().StringVar(&keepassPath, "path", "", "Path to the KeePass database")
}

func PreRunSetKeepassKeeperConfig(cmd *cobra.Command, args []string) {
	Replacement = map[string]any{
		"type": keepass.ConfigType,
	}
}
