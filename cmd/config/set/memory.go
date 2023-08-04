package set

import (
	"github.com/spf13/cobra"
)

var MemoryCmd = &cobra.Command{
	Use:    "memory <keeper-name>",
	Short:  "Configure a memory secret keeper",
	Args:   cobra.ExactArgs(1),
	PreRun: PreRunSetMemoryKeeperConfig,
	Run:    RunSetKeeperConfig,
}

func PreRunSetMemoryKeeperConfig(cmd *cobra.Command, args []string) {
	Replacement.Memory.Enable = true
}
