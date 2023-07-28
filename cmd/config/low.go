package config

import "github.com/spf13/cobra"

var LowCmd = &cobra.Command{
	Use:   "low",
	Short: "Configure a low security secret keeper",
}
