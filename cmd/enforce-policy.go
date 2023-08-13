package cmd

import (
	"context"

	"github.com/spf13/cobra"

	s "github.com/zostay/ghost/cmd/shared"
	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/secrets/policy"
)

var enforcePolicyCmd = &cobra.Command{
	Use:   "enforce-policy <keeper-name>",
	Short: "Enforce the named policy on the secret keeper",
	Args:  cobra.ExactArgs(1),
	Run:   RunEnforcePolicy,
}

func RunEnforcePolicy(cmd *cobra.Command, args []string) {
	keeperName := args[0]
	c := config.Instance()
	_, hasKeeper := c.Keepers[keeperName]
	if !hasKeeper {
		s.Logger.Panicf("Keeper %q is not configured.", keeperName)
	}

	ctx := context.Background()
	kpr, err := keeper.Build(ctx, keeperName, c)
	if err != nil {
		s.Logger.Panicf("Failed to load keeper %q: %s", keeperName, err)
	}

	if p, isPolicy := kpr.(*policy.Policy); isPolicy {
		err := p.EnforceGlobally(ctx)
		if err != nil {
			s.Logger.Panicf("Failed to enforce policy: %s", err)
		}
	} else {
		s.Logger.Panicf("Keeper %q is not a policy keeper.", keeperName)
	}
}
