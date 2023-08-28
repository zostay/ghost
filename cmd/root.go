package cmd

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
)

var (
	RootCmd = &cobra.Command{
		Use:              "ghost",
		Short:            "ghost is a tool for managing personal secrets",
		PersistentPreRun: s.RunRoot,
	}

	keeperName string
)

func init() {
	RootCmd.AddCommand(
		configCmd,
		deleteCmd,
		enforcePolicyCmd,
		getCmd,
		listCmd,
		randomCmd,
		serviceCmd,
		setCmd,
		syncCmd,
	)

	RootCmd.PersistentFlags().StringVarP(&s.ConfigFile, "config", "c", "", "path to the ghost configuration file")
}

func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}
