package keeper

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/http"
	"github.com/zostay/ghost/pkg/secrets/human"
	"github.com/zostay/ghost/pkg/secrets/keepass"
	"github.com/zostay/ghost/pkg/secrets/keyring"
	"github.com/zostay/ghost/pkg/secrets/lastpass"
	"github.com/zostay/ghost/pkg/secrets/low"
	"github.com/zostay/ghost/pkg/secrets/policy"
	"github.com/zostay/ghost/pkg/secrets/router"
	"github.com/zostay/ghost/pkg/secrets/seq"
)

func GetMasterPassword(name string) string {
	panic("not implemented")
}

func Build(
	ctx context.Context,
	name string,
	c *config.Config,
) (secrets.Keeper, error) {
	kc := c.Keepers[name]
	if kc == nil {
		return nil, fmt.Errorf("no configuration for keeper named %q", name)
	}

	switch kc.Type() {
	case config.KTKeepass:
		masterPassword := GetMasterPassword(name)
		kp, err := keepass.NewKeepass(kc.Keepass.Path, masterPassword)
		if err != nil {
			return nil, fmt.Errorf("unable to configure Keepass client %q: %w", name, err)
		}
		return kp, nil
	case config.KTLastPass:
		masterPassword := GetMasterPassword(name)
		lp, err := lastpass.NewLastPass(ctx, kc.LastPass.Username, masterPassword)
		if err != nil {
			return nil, fmt.Errorf("unable to configure LastPass client %q: %w", name, err)
		}
		return lp, nil
	case config.KTLowSecurity:
		return low.NewLowSecurity(kc.Low.Path), nil
	case config.KTGRPC:
		return http.NewClient(), nil
	case config.KTKeyring:
		return keyring.New(kc.Keyring.ServiceName), nil
	case config.KTMemory:
		return secrets.NewInternal()
	case config.KTHuman:
		kpr := human.New()
		for _, q := range kc.Human.Questions {
			kpr.AddQuestion(
				q.ID,
				q.AskFor,
				q.Presets,
			)
		}
		return kpr, nil
	case config.KTPolicy:
		nested, err := Build(ctx, kc.Policy.Keeper, c)
		if err != nil {
			return nil, fmt.Errorf("unable to build the secret keeper named %q for the policy keeper named %q: %w", kc.Policy.Keeper, name, err)
		}

		return policy.New(nested), nil

	case config.KTRouter:
		defaultKeeper, err := Build(ctx, kc.Router.DefaultRoute, c)
		if err != nil {
			return nil, fmt.Errorf("unable to build the secret keeper named %q for the default route of router named %q: %w", kc.Router.DefaultRoute, name, err)
		}

		r := router.NewRouter(defaultKeeper)
		for _, rt := range kc.Router.Routes {
			keeper, err := Build(ctx, rt.Keeper, c)
			if err != nil {
				locs := strings.Join(rt.Locations, ",")
				return nil, fmt.Errorf("unable to build the secret keeper named %q for the route to %q of router named %q: %w", rt.Keeper, locs, name, err)
			}

			err = r.AddKeeper(keeper, rt.Locations...)
			if err != nil {
				locs := strings.Join(rt.Locations, ",")
				return nil, fmt.Errorf("unable to add a route for the secret keeper named %q for the route to %q of router named %q: %w", rt.Keeper, locs, name, err)
			}
		}
		return r, nil
	case config.KTSeq:
		keepers := make([]secrets.Keeper, len(kc.Seq.Keepers))
		for i, k := range kc.Seq.Keepers {
			var err error
			keepers[i], err = Build(ctx, k, c)
			if err != nil {
				return nil, fmt.Errorf("unable to build the secret keeper named %q for the seq keeper named %q: %w", k, name, err)
			}
		}
		return seq.NewSeq(keepers...)
	case config.KTNone:
		return nil, errors.New("illegal use of none")
	case config.KTConflict:
		return nil, errors.New("illegal use of conflict")
	}
	return nil, fmt.Errorf("unknown secret keeper type for keeper named %q", name)
}
