package policy

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/zostay/ghost/pkg/secrets"
)

type Policy struct {
	secrets.Keeper
	defaultRule *Rule
	matchRule   []*MatchRule
}

var _ secrets.Keeper = &Policy{}

func New(kpr secrets.Keeper) *Policy {
	return &Policy{
		Keeper: kpr,
		defaultRule: &Rule{
			acceptance: Allow,
			lifetime:   0,
		},
		matchRule: []*MatchRule{},
	}
}

func (p *Policy) AddRule(r *MatchRule) {
	p.matchRule = append(p.matchRule, r)
}

// EnforceGlobally iterates through all the secrets in the nested keeper and
// applies the lifetime policy against those secrets.
func (p *Policy) EnforceGlobally(ctx context.Context) error {
	out := make(chan secrets.Secret)
	defer close(out)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for sec := range out {
			select {
			case <-ctx.Done():
				return
			default:
				wg.Add(1)
				go func(secret secrets.Secret) {
					defer wg.Done()
					_ = p.EnforceOne(ctx, secret)
				}(sec)
			}
		}
	}()

	err := secrets.ForEach(ctx, p.Keeper, func(sec secrets.Secret) error {
		out <- sec
		return ctx.Err()
	})

	wg.Wait()
	return err
}

// EnforceOne enforces the lifetime policy against a single secret.
func (p *Policy) EnforceOne(ctx context.Context, sec secrets.Secret) error {
	lifetime := p.lifetimeForSecret(sec)
	if lifetime == 0 {
		return nil
	}

	mtime := sec.LastModified()
	if time.Since(mtime) > lifetime {
		return p.DeleteSecret(ctx, sec.ID())
	}

	return nil
}

func (p *Policy) SetDefaultAcceptance(a Acceptance) {
	if a == InheritAcceptance {
		panic("default acceptance may not be set to inherit")
	}
	p.defaultRule.acceptance = a
}

func (p *Policy) SetDefaultLifetime(l time.Duration) {
	p.defaultRule.lifetime = l
}

// TODO Is the visibility/acceptance dichotomy sensible?

func (p *Policy) ListLocations(ctx context.Context) ([]string, error) {
	locs, err := p.Keeper.ListLocations(ctx)
	if err != nil {
		return nil, err
	}

	retLocs := make([]string, 0, len(locs))
Loc:
	for _, loc := range locs {
		for _, r := range p.matchRule {
			m := r.matchLocationAndAcceptable(loc)
			if m == matchYes {
				retLocs = append(retLocs, loc)
			} else if m == matchNo {
				continue Loc
			}
		}

		if p.defaultRule.acceptance == Allow {
			retLocs = append(retLocs, loc)
		}
	}

	return retLocs, nil
}

func (p *Policy) accessibleSecret(sec secrets.Secret) bool {
	for _, r := range p.matchRule {
		m := r.matchSecretAndAccessible(p.defaultRule, sec)
		if m == matchYes {
			return true
		} else if m == matchNo {
			return false
		}
	}

	return p.defaultRule.acceptance == Allow
}

func (p *Policy) lifetimeForSecret(sec secrets.Secret) time.Duration {
	for _, r := range p.matchRule {
		m, lt := r.matchSecretAndLifetime(sec)
		if m == matchYes {
			return lt
		}
	}

	return p.defaultRule.lifetime
}

func (p *Policy) ListSecrets(ctx context.Context, location string) ([]string, error) {
	ids, err := p.Keeper.ListSecrets(ctx, location)
	if err != nil {
		return nil, err
	}

	retSecs := make([]string, 0, len(ids))
	for _, id := range ids {
		sec, err := p.Keeper.GetSecret(ctx, id)
		if err != nil {
			return nil, err
		}

		if p.accessibleSecret(sec) {
			retSecs = append(retSecs, id)
		}
	}

	return retSecs, nil
}

func (p *Policy) GetSecretsByName(ctx context.Context, name string) ([]secrets.Secret, error) {
	secs, err := p.Keeper.GetSecretsByName(ctx, name)
	if err != nil {
		return nil, err
	}

	retSecs := make([]secrets.Secret, 0, len(secs))
	for _, sec := range secs {
		if p.accessibleSecret(sec) {
			retSecs = append(retSecs, sec)
		}
	}

	return retSecs, nil
}

func (p *Policy) GetSecret(ctx context.Context, id string) (secrets.Secret, error) {
	sec, err := p.Keeper.GetSecret(ctx, id)
	if err != nil {
		return nil, err
	}

	if p.accessibleSecret(sec) {
		return sec, nil
	}

	return nil, secrets.ErrNotFound
}

func (p *Policy) SetSecret(ctx context.Context, secret secrets.Secret) (secrets.Secret, error) {
	if p.accessibleSecret(secret) {
		return p.Keeper.SetSecret(ctx, secret)
	}
	return nil, errors.New("secret is not writable")
}

func (p *Policy) CopySecret(ctx context.Context, id string, location string) (secrets.Secret, error) {
	sec, err := p.Keeper.GetSecret(ctx, id)
	if err != nil {
		return nil, err
	}

	if !p.accessibleSecret(sec) {
		return nil, secrets.ErrNotFound
	}

	potentialSec := secrets.NewSingleFromSecret(sec,
		secrets.WithLocation(location))

	if !p.accessibleSecret(potentialSec) {
		return nil, errors.New("secret is not writable")
	}

	return p.Keeper.CopySecret(ctx, id, location)
}

func (p *Policy) MoveSecret(ctx context.Context, id string, location string) (secrets.Secret, error) {
	sec, err := p.Keeper.GetSecret(ctx, id)
	if err != nil {
		return nil, err
	}

	if !p.accessibleSecret(sec) {
		return nil, secrets.ErrNotFound
	}

	potentialSec := secrets.NewSingleFromSecret(sec,
		secrets.WithLocation(location))

	if !p.accessibleSecret(potentialSec) {
		return nil, errors.New("secret is not writable")
	}

	return p.Keeper.MoveSecret(ctx, id, location)
}

func (p *Policy) DeleteSecret(ctx context.Context, id string) error {
	sec, err := p.Keeper.GetSecret(ctx, id)
	if err != nil {
		return err
	}

	if !p.accessibleSecret(sec) {
		return nil
	}

	return p.Keeper.DeleteSecret(ctx, id)
}
