package config

import "github.com/spf13/cobra"

var RouterCmd = &cobra.Command{
	Use:   "router",
	Short: "Configure a router secret keeper",
}
