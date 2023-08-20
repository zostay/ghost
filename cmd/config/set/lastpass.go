package set

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/lastpass"
)

var (
	LastPassCmd = &cobra.Command{
		Use:    "lastpass <keeer-name> [flags]",
		Short:  "Configure a LastPass secret keeper",
		Args:   cobra.MinimumNArgs(1),
		PreRun: PreRunSetLastPassKeeperConfig,
		Run:    RunSetKeeperConfig,
	}

	lastPassUsername string
)

func init() {
	LastPassCmd.Flags().StringVar(&lastPassUsername, "username", "", "LastPass username")
}

func PreRunSetLastPassKeeperConfig(cmd *cobra.Command, args []string) {
	keeperName := args[0]
	c := config.Instance()
	kc := c.Keepers[keeperName]
	if kc == nil {
		kc = map[string]any{
			"type": lastpass.ConfigType,
		}
	}

	if lastPassUsername != "" {
		kc["username"] = lastPassUsername
	}

	Replacement = kc
}
