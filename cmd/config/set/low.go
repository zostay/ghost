package set

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/low"
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
	keeperName := args[0]
	c := config.Instance()
	kc := c.Keepers[keeperName]
	if kc == nil {
		kc = map[string]any{
			"type": low.ConfigType,
		}
	}

	if lowPath != "" {
		kc["path"] = lowPath
	}

	Replacement = kc
}
