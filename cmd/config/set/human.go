package set

import (
	"github.com/spf13/cobra"
	"github.com/zostay/go-std/slices"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/human"
)

var (
	HumanCmd = &cobra.Command{
		Use:    "human <keeper-name>",
		Short:  "Configure a human secret keeper",
		Args:   cobra.ExactArgs(1),
		PreRun: PreRunSetHumanKeeperConfig,
		Run:    RunSetKeeperConfig,
	}

	setQuestion, removeQuestion string
	askFor                      []string
	presets                     map[string]string
)

func init() {
	HumanCmd.Flags().StringSliceVar(&askFor, "ask-for", []string{}, "Ask for a secret value")
	HumanCmd.Flags().StringToStringVar(&presets, "preset", map[string]string{}, "Set a preset value")
	HumanCmd.Flags().StringVar(&setQuestion, "set", "", "Add or update a secret value with the given ID")
	HumanCmd.Flags().StringVar(&removeQuestion, "remove", "", "Remove a secret value with the given ID")
}

func PreRunSetHumanKeeperConfig(cmd *cobra.Command, args []string) {
	keeperName := args[0]

	if setQuestion != "" && removeQuestion != "" {
		s.Logger.Panic("cannot set and remove a secret value in the same step")
	}

	if setQuestion != "" || removeQuestion != "" {
		s.Logger.Panic("you must set or remove a secret value with this command")
	}

	if removeQuestion != "" && (len(askFor) > 0 || len(presets) > 0) {
		s.Logger.Panic("--remove is incompatible with --ask-for or --preset")
	}

	if setQuestion != "" && len(askFor) == 0 && len(presets) == 0 {
		s.Logger.Panic("--set requires --ask-for or --preset")
	}

	c := config.Instance()
	kc := c.Keepers[keeperName]
	if kc == nil {
		kc = map[string]any{
			"type": human.ConfigType,
		}
	}

	hc := HumanConfig(kc)
	Replacement = kc

	if removeQuestion != "" {
		hc.Remove(removeQuestion)
		return
	}

	hc.Set(setQuestion, presets, askFor)
}

type HumanConfig config.KeeperConfig

func (hc HumanConfig) Set(
	id string,
	presets map[string]string,
	askFor []string,
) {
	qs := hc["questions"].([]map[string]any)
	if qs == nil {
		qs = []map[string]any{}
	}

	qs = append(qs, map[string]any{
		"id":      id,
		"presets": presets,
		"ask_for": askFor,
	})

	hc["questions"] = qs
}

func (hc HumanConfig) Remove(id string) {
	qs := hc["questions"].([]map[string]any)
	if qs == nil {
		return
	}

	i := slices.FirstIndex(qs, func(q map[string]any) bool {
		return q["id"] == id
	})
	if i >= 0 {
		qs = slices.Delete(qs, i)
	}

	hc["questions"] = qs
}
