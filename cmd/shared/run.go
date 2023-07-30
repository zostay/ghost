package shared

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/config"
)

func RunRoot(cmd *cobra.Command, args []string) {
	Logger = log.New(cmd.OutOrStderr(), "", 0)

	var err error
	err = config.Instance().Load(ConfigFile)
	if err != nil {
		Logger.Panicf("Failure to load configuration: %v", err)
	}

	err = config.Instance().Check()
	if err != nil {
		Logger.Panicf("Configuration errors: %v", err)
	}
}
