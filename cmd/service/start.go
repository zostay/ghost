package service

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets/policy"
)

var (
	StartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the ghost service",
		Run:   RunStartService,
	}

	enforceAllPolicies bool
	enforcePolicies    []string
	enforcementPeriod  time.Duration
	keeperService      string
)

func init() {
	StartCmd.Flags().BoolVar(&enforceAllPolicies, "enforce-all-policies", false, "enforce all policies")
	StartCmd.Flags().StringSliceVar(&enforcePolicies, "enforce-policy", []string{}, "enforce the named policies")
	StartCmd.Flags().DurationVar(&enforcementPeriod, "enforcement-period", 1*time.Minute, "enforce policies every period")
	StartCmd.Flags().StringVar(&keeperService, "keeper", "", "the name of the keeper service to use (master used by default)")
}

func RunStartService(cmd *cobra.Command, args []string) {
	if enforceAllPolicies && len(enforcePolicies) > 0 {
		s.Logger.Panic("cannot use --enforce-all-policies and --enforce-policy together")
		return
	}

	if enforcementPeriod < 2*time.Second {
		s.Logger.Panic("enforcement period is too short")
		return
	}

	c := config.Instance()
	if keeperService == "" {
		keeperService = c.MasterKeeper
	}

	if keeperService == "" {
		s.Logger.Panic("Please specify --keeper or set a master in the configuration file")
		return
	}

	err := keeper.RecoverService()
	if err != nil {
		s.Logger.Panicf("Service is in a bad state, cannot start: %v", err)
		return
	}

	ctx := keeper.WithBuilder(context.Background(), c)
	kpr, err := keeper.Build(ctx, keeperService)
	if err != nil {
		s.Logger.Panicf("Failed to configure master keeper %q: %v", keeperService, err)
		return
	}

	if enforceAllPolicies {
		for name, cfg := range c.Keepers {
			if plugin.Type(cfg) == policy.ConfigType {
				enforcePolicies = append(enforcePolicies, name)
			}
		}
	}

	startPolicyEnforcement(ctx, c)

	err = keeper.StartServer(
		s.Logger,
		kpr,
		keeperService,
		enforcementPeriod,
		enforcePolicies)
	if err != nil {
		s.Logger.Panic(err)
	}
}

func startPolicyEnforcement(ctx context.Context, c *config.Config) {
	for _, name := range enforcePolicies {
		if plugin.Type(c.Keepers[name]) != policy.ConfigType {
			s.Logger.Panicf("keeper %q is not a policy keeper", name)
		}

		go enforcePolicy(ctx, name)
	}
}

func enforcePolicy(
	ctx context.Context,
	name string,
) {
	kpr, err := keeper.Build(ctx, name)
	if err != nil {
		s.Logger.Panicf("failed to configure policy keeper %q: %v", name, err)
	}

	p := kpr.(*policy.Policy)
	for {
		enforcePolicyThenWait(ctx, name, p)
	}
}

func enforcePolicyThenWait(
	ctx context.Context,
	name string,
	p *policy.Policy,
) {
	ctx, cancel := context.WithTimeout(ctx, enforcementPeriod-1*time.Second)
	defer cancel()
	go func() {
		err := p.EnforceGlobally(ctx)
		if err != nil {
			s.Logger.Printf("failed to enforce policy %q: %v", name, err)
		}
	}()
	<-time.After(enforcementPeriod)
}
