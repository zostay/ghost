package set

import (
	"github.com/spf13/cobra"
)

var (
	LowSecurityCmd = &cobra.Command{
		Use:    "low <keeper-name> [flags]",
		Short:  "Configure a low-security secret keeper",
		Args:   cobra.MinimumNArgs(1),
		PreRun: PreRunSetLowKeeperConfig,
		Run:    RunSetKeeperConfig,
	}

	lowPath string
)

func init() {
	LowSecurityCmd.Flags().StringVar(&lowPath, "path", "", "Path to the low-level configuration file")
}

func PreRunSetLowKeeperConfig(cmd *cobra.Command, args []string) {
	Replacement.Low.Path = lowPath
}
