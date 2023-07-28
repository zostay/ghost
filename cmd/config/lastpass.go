package config

import "github.com/spf13/cobra"

var LastPassCmd = &cobra.Command{
	Use:   "lastpass",
	Short: "Configure a LastPass secret keeper",
}
