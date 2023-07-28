package service

import "github.com/spf13/cobra"

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the ghost service",
}
