package service

import "github.com/spf13/cobra"

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the ghost service",
}
