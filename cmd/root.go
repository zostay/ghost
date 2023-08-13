package cmd

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
)

var (
	rootCmd = &cobra.Command{
		Use:              "ghost",
		Short:            "ghost is a tool for managing personal secrets",
		PersistentPreRun: s.RunRoot,
	}

	keeperName string
)

func init() {
	rootCmd.AddCommand(
		configCmd,
		deleteCmd,
		enforcePolicyCmd,
		getCmd,
		listCmd,
		serviceCmd,
		setCmd,
		syncCmd,
	)

	rootCmd.PersistentFlags().StringVarP(&s.ConfigFile, "config", "c", "", "path to the ghost configuration file")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
