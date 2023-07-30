package set

import (
	"github.com/spf13/cobra"
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
	Replacement.LastPass.Username = lastPassUsername
}
