package set

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/slices"
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
	if appendKeepers && deleteKeepers {
		return errors.New("cannot append and delete at the same time")
	}

	if appendKeepers {
		Replacement.Seq.Keepers = append(Replacement.Seq.Keepers, args[1:]...)
		return nil
	}

	if deleteKeepers {
		for _, k := range args[1:] {
			Replacement.Seq.Keepers = slices.DeleteValue(
				Replacement.Seq.Keepers,
				k)
		}
		return nil
	}

	Replacement.Seq.Keepers = args[1:]
	return nil
}
