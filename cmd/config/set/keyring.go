package set

import "github.com/spf13/cobra"

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
	Replacement.Keyring.ServiceName = service
}
