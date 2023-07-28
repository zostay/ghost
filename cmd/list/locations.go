package list

import "github.com/spf13/cobra"

var LocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List the locations that may contain secrets",
}
