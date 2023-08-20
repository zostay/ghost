package set

import (
	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/http"
)

var GRPCCmd = &cobra.Command{
	Use:    "grpc <keeper-name>",
	Short:  "Configure a gRPC secret keeper",
	Args:   cobra.ExactArgs(1),
	PreRun: PreRunSetGRPCKeeperConfig,
	Run:    RunSetKeeperConfig,
}

func PreRunSetGRPCKeeperConfig(cmd *cobra.Command, args []string) {
	Replacement = config.KeeperConfig{
		"type": http.ConfigType,
	}
}
