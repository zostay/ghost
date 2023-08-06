package set

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
)

var (
	HumanCmd = &cobra.Command{
		Use:    "human <keeper-name>",
		Short:  "Configure a human secret keeper",
		Args:   cobra.ExactArgs(1),
		PreRun: PreRunSetHumanKeeperConfig,
		Run:    RunSetKeeperConfig,
	}

	set, remove string
	askFor      []string
	presets     map[string]string
)

func init() {
	HumanCmd.Flags().StringSliceVar(&askFor, "ask-for", []string{}, "Ask for a secret value")
	HumanCmd.Flags().StringToStringVar(&presets, "preset", map[string]string{}, "Set a preset value")
	HumanCmd.Flags().StringVar(&set, "set", "", "Add or update a secret value with the given ID")
	HumanCmd.Flags().StringVar(&remove, "remove", "", "Remove a secret value with the given ID")
}

func PreRunSetHumanKeeperConfig(cmd *cobra.Command, args []string) {
	if set != "" && remove != "" {
		s.Logger.Panic("cannot set and remove a secret value in the same step")
	}

	if set != "" || remove != "" {
		s.Logger.Panic("you must set or remove a secret value with this command")
	}

	if remove != "" && (len(askFor) > 0 || len(presets) > 0) {
		s.Logger.Panic("--remove is incompatible with --ask-for or --preset")
	}

	if set != "" && len(askFor) == 0 && len(presets) == 0 {
		s.Logger.Panic("--set requires --ask-for or --preset")
	}

	if remove != "" {
		Replacement.Human.Remove(remove)
		return
	}

	Replacement.Human.Set(set, presets, askFor)
}
