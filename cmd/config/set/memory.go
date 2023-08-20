package set

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/secrets/memory"
)

var MemoryCmd = &cobra.Command{
	Use:    "memory <keeper-name>",
	Short:  "Configure a memory secret keeper",
	Args:   cobra.ExactArgs(1),
	PreRun: PreRunSetMemoryKeeperConfig,
	Run:    RunSetKeeperConfig,
}

func PreRunSetMemoryKeeperConfig(cmd *cobra.Command, args []string) {
	Replacement = map[string]any{
		"type": memory.ConfigType,
	}
}
