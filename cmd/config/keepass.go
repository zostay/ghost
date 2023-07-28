package config

import "github.com/spf13/cobra"

var KeepassCmd = &cobra.Command{
	Use:   "keepass",
	Short: "Configure a keepass secret keeper",
}
