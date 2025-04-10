package service

import (
	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/keeper"
)

var (
	StopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the ghost service",
		Run:   RunStopService,
	}

	quit bool
	kill bool
)

func init() {
	StopCmd.Flags().BoolVar(&quit, "quit", false, "send SIGQUIT instead of SIGHUP")
	StopCmd.Flags().BoolVar(&kill, "kill", false, "send SIGKILL instead of SIGHUP")
}

func RunStopService(_ *cobra.Command, _ []string) {
	quickness := keeper.StopGraceful
	switch {
	case kill:
		quickness = keeper.StopNow
	case quit:
		quickness = keeper.StopQuick
	}

	err := keeper.StopServer(quickness)
	if err != nil {
		s.Logger.Panic(err)
	}
}
