package set

import "github.com/spf13/cobra"

var GRPCCmd = &cobra.Command{
	Use:    "grpc <keeper-name>",
	Short:  "Configure a gRPC secret keeper",
	Args:   cobra.ExactArgs(1),
	PreRun: PreRunSetGRPCKeeperConfig,
	Run:    RunSetKeeperConfig,
}

func PreRunSetGRPCKeeperConfig(cmd *cobra.Command, args []string) {
	Replacement.GRPC.Listener = "unix"
}
