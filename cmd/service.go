package cmd

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/cmd/service"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage the ghost service",
}

func init() {
	serviceCmd.AddCommand(service.StartCmd)
	serviceCmd.AddCommand(service.StopCmd)
	serviceCmd.AddCommand(service.StatusCmd)
}
