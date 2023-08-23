package cache

import (
	"context"
	"errors"
	"time"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/memory"
)

// Cache is a secret keeper that wraps another secret keeper and caches
// secrets in memory. Writing to it directly is not permitted.
type Cache struct {
	secrets.Keeper   // the secret keeper to cache
	*memory.Internal // the memory keeper used to store cached secrets

	touchOnRead bool // update last modified on GetSecret* calls
}

// New creates a new caching secret keeper. The keeper will cache
// secrets in memory and will wrap the given secret keeper. The
// touchOnRead flag will cause the last modified date of secrets to
// be updated on GetSecret* calls.
func New(k secrets.Keeper, touchOnRead bool) (*Cache, error) {
	mem, err := memory.New()
	if err != nil {
		return nil, err
	}

	return &Cache{
		Internal: mem,
		Keeper:   k,

		touchOnRead: touchOnRead,
	}, nil
}

var _ secrets.Keeper = &Cache{}

// ListLocations returns the list of locations in the wrapped secret keeper.
func (c *Cache) ListLocations(ctx context.Context) ([]string, error) {
	return c.Keeper.ListLocations(ctx)
}

// ListSecrets returns the list of secrets in the wrapped secret keeper.
func (c *Cache) ListSecrets(ctx context.Context, loc string) ([]string, error) {
	return c.Keeper.ListSecrets(ctx, loc)
}

func (c *Cache) touchSecret(ctx context.Context, sec secrets.Secret) (secrets.Secret, error) {
	updSec := secrets.NewSingleFromSecret(sec,
		secrets.WithLastModified(time.Now()))
	cacheSec, err := c.Internal.SetSecret(ctx, updSec)
	if err != nil {
		return sec, nil
	}
	return cacheSec, nil
}

// GetSecret returns the secret with the given ID from the wrapped secret keeper
// on first call. Subsequent calls will return the cached secret. If the
// touchOnRead flag is set, the last modified date of the secret will be updated
// on each call.
func (c *Cache) GetSecret(ctx context.Context, id string) (secrets.Secret, error) {
	sec, _ := c.Internal.GetSecret(ctx, id)
	if sec != nil {
		if c.touchOnRead {
			return c.touchSecret(ctx, sec)
		}

		return sec, nil
	}

	sec, err := c.Keeper.GetSecret(ctx, id)
	if err != nil {
		return nil, err
	}

	return c.touchSecret(ctx, sec)
}

func (c *Cache) touchSecrets(ctx context.Context, secs []secrets.Secret) ([]secrets.Secret, error) {
	var (
		err     error
		newSecs = make([]secrets.Secret, len(secs))
	)

	for i, sec := range secs {
		newSecs[i], err = c.touchSecret(ctx, sec)
		if err != nil {
			return secs, nil
		}
	}

	return newSecs, nil
}

// GetSecretsByName returns the list of secrets with the given name from
// the wrapped secret keeper on first call. Subsequent calls will return the
// cached list of secrets. If the touchOnRead flag is set, the last modified
// date of the secrets will be updated on each call.
func (c *Cache) GetSecretsByName(ctx context.Context, name string) ([]secrets.Secret, error) {
	secs, _ := c.Internal.GetSecretsByName(ctx, name)
	if len(secs) > 0 {
		if c.touchOnRead {
			return c.touchSecrets(ctx, secs)
		}
	}

	secs, err := c.Keeper.GetSecretsByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return c.touchSecrets(ctx, secs)
}

// SetSecret cannot be used and always fails with an error.
func (c *Cache) SetSecret(context.Context, secrets.Secret) (secrets.Secret, error) {
	return nil, errors.New("caching secret keeper does not allow direct writes")
}

// CopySecret cannot be used and always fails with an error.
func (c *Cache) CopySecret(context.Context, string, string) (secrets.Secret, error) {
	return nil, errors.New("caching secret keeper does not allow direct writes")
}

// MoveSecret cannot be used and always fails with an error.
func (c *Cache) MoveSecret(context.Context, string, string) (secrets.Secret, error) {
	return nil, errors.New("caching secret keeper does not allow direct writes")
}

// DeleteSecret deletes the secret with the given ID from the cache only. This
// does not delete the secret from the wrapped secret keeper.
func (c *Cache) DeleteSecret(ctx context.Context, id string) error {
	return c.Internal.DeleteSecret(ctx, id)
}
