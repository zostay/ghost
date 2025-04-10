package service

import (
	"github.com/spf13/cobra"
	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/keeper"
)

var (
	StatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show whether the ghost service is running other stats",
		Run:   RunServiceStatus,
	}
)

func RunServiceStatus(_ *cobra.Command, _ []string) {
	info, err := keeper.CheckServer()
	if err != nil {
		s.Logger.Panic(err)
	}

	if len(info.EnforcedPolicies) > 0 {
		s.Logger.Printf(
			"Ghost is running: PID=%d Keeper=%q Enforcement={Period=%v Policies=%v}",
			info.Pid,
			info.Keeper,
			info.EnforcementPeriod,
			info.EnforcedPolicies)
	} else {
		s.Logger.Printf(
			"Ghost is running: PID=%d Keeper=%q Enforcement=none",
			info.Pid,
			info.Keeper)
	}
}
