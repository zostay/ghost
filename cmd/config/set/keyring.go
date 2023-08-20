package set

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/keyring"
)

var (
	KeyringCmd = &cobra.Command{
		Use:    "keyring <keeper-name> [flags]",
		Short:  "Configure a secret keeper that uses the system keyring",
		Args:   cobra.ExactArgs(1),
		PreRun: PreRunSetKeyringKeeperConfig,
		Run:    RunSetKeeperConfig,
	}

	service string
)

func init() {
	KeyringCmd.Flags().StringVar(&service, "service", "ghost", "The name of the service to use in the keyring")
}

func PreRunSetKeyringKeeperConfig(cmd *cobra.Command, args []string) {
	keeperName := args[0]
	c := config.Instance()
	kc := c.Keepers[keeperName]
	if kc == nil {
		kc = map[string]any{
			"type": keyring.ConfigType,
		}
	}

	if service != "" {
		kc["service_name"] = service
	}

	Replacement = kc
}
