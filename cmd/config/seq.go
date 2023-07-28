package config

import "github.com/spf13/cobra"

var SeqCmd = &cobra.Command{
	Use:   "seq",
	Short: "Configure a seq secret keeper",
}
