package policy

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/zostay/ghost/pkg/secrets"
)

// Policy is a secret keeper that wraps another secret keeper and applies
// policy rules to the secrets in the nested keeper.
type Policy struct {
	secrets.Keeper
	defaultRule *Rule
	matchRule   []*MatchRule
}

var _ secrets.Keeper = &Policy{}

// New creates a new policy secret keeper.
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

// AddRule adds a rule to the policy.
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

// SetDefaultAcceptance sets the default acceptance policy for the policy.
func (p *Policy) SetDefaultAcceptance(a Acceptance) {
	if a == InheritAcceptance {
		panic("default acceptance may not be set to inherit")
	}
	p.defaultRule.acceptance = a
}

// SetDefaultLifetime sets the default lifetime for the policy.
func (p *Policy) SetDefaultLifetime(l time.Duration) {
	p.defaultRule.lifetime = l
}

// ListLocations lists the locations in the nested keeper that are accessible
// to the policy.
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
		switch m {
		case matchYes:
			return true
		case matchNo:
			return false
		case matchMiss:
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

// ListSecrets lists the secrets in the nested keeper that are accessible to
// the policy.
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

// GetSecretsByName retrieves all secrets with the given name that are
// accessible by the policy.
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

// GetSecret retrieves the identified secret from the nested keeper if it is
// accessible by the policy.
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

// SetSecret saves the named secret to the given value in the nested keeper if
// it is accessible by the policy.
func (p *Policy) SetSecret(ctx context.Context, secret secrets.Secret) (secrets.Secret, error) {
	if p.accessibleSecret(secret) {
		return p.Keeper.SetSecret(ctx, secret)
	}
	return nil, errors.New("secret is not writable")
}

// CopySecret copies the identified secret to the given location in the nested
// keeper if it is accessible by the policy.
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

// MoveSecret moves the identified secret to the given location in the nested
// keeper if it is accessible by the policy.
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

// DeleteSecret deletes the identified secret from the nested keeper if it is
// accessible by the policy.
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
