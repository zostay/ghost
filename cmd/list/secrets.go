package list

import "github.com/spf13/cobra"

var SecretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "List secrets",
}
