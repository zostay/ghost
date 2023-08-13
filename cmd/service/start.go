package service

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
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
)

func init() {
	StartCmd.Flags().BoolVar(&enforceAllPolicies, "enforce-all-policies", false, "enforce all policies")
	StartCmd.Flags().StringSliceVar(&enforcePolicies, "enforce-policy", []string{}, "enforce the named policies")
	StartCmd.Flags().DurationVar(&enforcementPeriod, "enforcement-period", 1*time.Minute, "enforce policies every period")
}

func RunStartService(cmd *cobra.Command, args []string) {
	if enforceAllPolicies && len(enforcePolicies) > 0 {
		s.Logger.Panic("cannot use --enforce-all-policies and --enforce-policy together")
	}

	if enforcementPeriod < 2*time.Second {
		s.Logger.Panic("enforcement period is too short")
	}

	c := config.Instance()
	if c.MasterKeeper == "" {
		s.Logger.Panic("no master keeper set")
	}

	ctx := context.Background()

	kpr, err := keeper.Build(ctx, c.MasterKeeper, c)
	if err != nil {
		s.Logger.Panicf("failed to configure master keeper %q: %v", c.MasterKeeper, err)
	}

	if enforceAllPolicies {
		for name, cfg := range c.Keepers {
			if cfg.Type() == config.KTPolicy {
				enforcePolicies = append(enforcePolicies, name)
			}
		}
	}

	startPolicyEnforcement(ctx, c)

	err = keeper.StartServer(kpr)
	if err != nil {
		s.Logger.Panic(err)
	}
}

func startPolicyEnforcement(ctx context.Context, c *config.Config) {
	for _, name := range enforcePolicies {
		if c.Keepers[name].Type() != config.KTPolicy {
			s.Logger.Panicf("keeper %q is not a policy keeper", name)
		}

		go func() {
			kpr, err := keeper.Build(ctx, name, c)
			if err != nil {
				s.Logger.Panicf("failed to configure policy keeper %q: %v", name, err)
			}

			p := kpr.(*policy.Policy)
			for {
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
		}()
	}
}
