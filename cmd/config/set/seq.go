package set

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/seq"
)

var (
	SeqCmd = &cobra.Command{
		Use:     "seq <keeper-name> [flags] [<keepers>...]",
		Short:   "Configure a sequential secret keeper",
		Args:    cobra.MinimumNArgs(2),
		PreRunE: PreRunSetSeqKeeperConfig,
		Run:     RunSetKeeperConfig,
	}

	appendKeepers bool
	deleteKeepers bool
)

func init() {
	SeqCmd.Flags().BoolVarP(&appendKeepers, "append", "a", false, "Append the named keepers to the end of the sequence")
	SeqCmd.Flags().BoolVarP(&deleteKeepers, "delete", "d", false, "Remove the named keepers from the sequence")
}

func PreRunSetSeqKeeperConfig(cmd *cobra.Command, args []string) error {
	keeperName := args[0]
	modKeepers := args[1:]

	c := config.Instance()
	kc := c.Keepers[keeperName]
	if kc == nil {
		kc = map[string]any{
			"type":    seq.ConfigType,
			"keepers": []string{},
		}
	}

	if appendKeepers && deleteKeepers {
		return errors.New("cannot append and delete at the same time")
	}

	if appendKeepers {
		keepers := kc["keepers"].([]string)
		keepers = append(keepers, modKeepers...)
		kc["keepers"] = keepers
		return nil
	}

	if deleteKeepers {
		keepers := kc["keepers"].([]string)
		for _, k := range modKeepers {
			keepers = slices.DeleteValue(keepers, k)
		}
		kc["keepers"] = keepers
		return nil
	}

	Replacement = kc
	return nil
}
