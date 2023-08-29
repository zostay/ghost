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
	secrets.Keeper // the secret keeper to cache
	*memory.Memory // the memory keeper used to store cached secrets

	origToCacheId map[string]string
	cacheToOrigId map[string]string

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
		Memory: mem,
		Keeper: k,

		origToCacheId: map[string]string{},
		cacheToOrigId: map[string]string{},

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

func (c *Cache) touchSecret(
	ctx context.Context,
	sec secrets.Secret,
	cacheId,
	id string,
) (secrets.Secret, error) {
	updSec := secrets.NewSingleFromSecret(sec,
		secrets.WithID(cacheId),
		secrets.WithLastModified(time.Now()))
	cacheSec, err := c.Memory.SetSecret(ctx, updSec)
	if err != nil {
		return sec, nil
	}
	if cacheId == "" {
		cacheId = cacheSec.ID()
	}
	c.origToCacheId[id] = cacheId
	c.cacheToOrigId[cacheId] = id
	return secrets.NewSingleFromSecret(cacheSec, secrets.WithID(id)), nil
}

// GetSecret returns the secret with the given ID from the wrapped secret keeper
// on first call. Subsequent calls will return the cached secret. If the
// touchOnRead flag is set, the last modified date of the secret will be updated
// on each call.
func (c *Cache) GetSecret(ctx context.Context, id string) (secrets.Secret, error) {
	if cacheId, isCached := c.origToCacheId[id]; isCached {
		sec, _ := c.Memory.GetSecret(ctx, cacheId)
		if sec != nil {
			if c.touchOnRead {
				return c.touchSecret(ctx, sec, cacheId, id)
			}

			return secrets.NewSingleFromSecret(sec, secrets.WithID(id)), nil
		}
	}

	sec, err := c.Keeper.GetSecret(ctx, id)
	if err != nil {
		return nil, err
	}

	return c.touchSecret(ctx, sec, "", id)
}

func (c *Cache) touchSecretsFromCache(ctx context.Context, cachedSecs []secrets.Secret) ([]secrets.Secret, error) {
	var (
		err     error
		newSecs = make([]secrets.Secret, len(cachedSecs))
	)

	for i, sec := range cachedSecs {
		id := c.cacheToOrigId[sec.ID()]
		fixedSec := secrets.NewSingleFromSecret(sec, secrets.WithID(id))
		newSecs[i], err = c.touchSecret(ctx, fixedSec, sec.ID(), id)
		if err != nil {
			return nil, err
		}
	}

	return newSecs, nil
}

func (c *Cache) rewriteCachedIds(cachedSecs []secrets.Secret) []secrets.Secret {
	origSecs := make([]secrets.Secret, len(cachedSecs))
	for i, sec := range cachedSecs {
		id := c.cacheToOrigId[sec.ID()]
		origSecs[i] = secrets.NewSingleFromSecret(sec, secrets.WithID(id))
	}
	return origSecs
}

func (c *Cache) touchSecretsFromOrig(ctx context.Context, secs []secrets.Secret) ([]secrets.Secret, error) {
	for _, sec := range secs {
		_, err := c.touchSecret(ctx, sec, "", sec.ID())
		if err != nil {
			return nil, err
		}
	}

	return secs, nil
}

// GetSecretsByName returns the list of secrets with the given name from
// the wrapped secret keeper on first call. Subsequent calls will return the
// cached list of secrets. If the touchOnRead flag is set, the last modified
// date of the secrets will be updated on each call.
func (c *Cache) GetSecretsByName(ctx context.Context, name string) ([]secrets.Secret, error) {
	secs, _ := c.Memory.GetSecretsByName(ctx, name)
	if len(secs) > 0 {
		if c.touchOnRead {
			return c.touchSecretsFromCache(ctx, secs)
		}

		return c.rewriteCachedIds(secs), nil
	}

	secs, err := c.Keeper.GetSecretsByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return c.touchSecretsFromOrig(ctx, secs)
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
	if cacheId, isCached := c.origToCacheId[id]; isCached {
		err := c.Memory.DeleteSecret(ctx, cacheId)
		if err != nil {
			return err
		}

		delete(c.origToCacheId, id)
		delete(c.cacheToOrigId, cacheId)
	}
	return nil
}
