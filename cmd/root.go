package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "ghost",
	Short: "ghost is a tool for managing personal secrets",
}

func init() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(syncCmd)
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
