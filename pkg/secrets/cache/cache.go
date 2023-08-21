package cache

import (
	"context"
	"errors"
	"time"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/memory"
)

type Cache struct {
	secrets.Keeper
	*memory.Internal

	touchOnRead bool
}

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

func (c *Cache) ListLocations(ctx context.Context) ([]string, error) {
	return c.Keeper.ListLocations(ctx)
}

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

func (c *Cache) SetSecret(context.Context, secrets.Secret) (secrets.Secret, error) {
	return nil, errors.New("caching secret keeper does not allow direct writes")
}

func (c *Cache) CopySecret(context.Context, string, string) (secrets.Secret, error) {
	return nil, errors.New("caching secret keeper does not allow direct writes")
}

func (c *Cache) MoveSecret(context.Context, string, string) (secrets.Secret, error) {
	return nil, errors.New("caching secret keeper does not allow direct writes")
}

func (c *Cache) DeleteSecret(ctx context.Context, id string) error {
	return c.Internal.DeleteSecret(ctx, id)
}
