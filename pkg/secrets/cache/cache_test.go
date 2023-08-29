package cache_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/cache"
	"github.com/zostay/ghost/pkg/secrets/keepertest"
	"github.com/zostay/ghost/pkg/secrets/memory"
)

func TestCache(t *testing.T) { //nolint:tparallel // it is parallel, you dolt
	t.Parallel()

	factory := func() (secrets.Keeper, error) {
		m, err := memory.New()
		if err != nil {
			return nil, err
		}
		return cache.New(m, false)
	}

	ts := keepertest.New(factory)
	t.Run("SecretKeeperGetMissingTest", ts.SecretKeeperGetMissingTest)
}

func TestCache_GetSecret(t *testing.T) {
	t.Parallel()

	m, err := memory.New()
	assert.NoError(t, err)
	require.NotNil(t, m)

	c, err := cache.New(m, false)
	assert.NoError(t, err)
	require.NotNil(t, c)

	ctx := context.Background()

	s1, err := m.SetSecret(ctx, secrets.NewSecret("test", "test", "test"))
	assert.NoError(t, err)
	assert.NotEmpty(t, s1.ID())

	s2, err := c.GetSecret(ctx, s1.ID())
	assert.NoError(t, err)
	require.NotNil(t, s2)
	assert.Equal(t, s1.ID(), s2.ID())
	assert.Equal(t, s1.Name(), s2.Name())
	assert.Equal(t, s1.Password(), s2.Password())

	err = m.DeleteSecret(ctx, s1.ID())
	assert.NoError(t, err)

	s3, err := c.GetSecret(ctx, s1.ID())
	assert.NoError(t, err)
	require.NotNil(t, s3)
	assert.Equal(t, s1.ID(), s3.ID())
	assert.Equal(t, s1.Name(), s3.Name())
	assert.Equal(t, s1.Password(), s3.Password())
}

func TestCache_GetSecretsByName(t *testing.T) {
	t.Parallel()

	m, err := memory.New()
	assert.NoError(t, err)
	require.NotNil(t, m)

	c, err := cache.New(m, false)
	assert.NoError(t, err)
	require.NotNil(t, c)

	ctx := context.Background()

	s1, err := m.SetSecret(ctx, secrets.NewSecret("test", "test", "test"))
	assert.NoError(t, err)
	assert.NotEmpty(t, s1.ID())

	s2s, err := c.GetSecretsByName(ctx, "test")
	assert.NoError(t, err)
	require.Len(t, s2s, 1)
	assert.Equal(t, s1.ID(), s2s[0].ID())
	assert.Equal(t, s1.Name(), s2s[0].Name())
	assert.Equal(t, s1.Password(), s2s[0].Password())

	err = m.DeleteSecret(ctx, s1.ID())
	assert.NoError(t, err)

	s3s, err := c.GetSecretsByName(ctx, "test")
	assert.NoError(t, err)
	require.Len(t, s3s, 1)
	assert.Equal(t, s1.ID(), s3s[0].ID())
	assert.Equal(t, s1.Name(), s3s[0].Name())
	assert.Equal(t, s1.Password(), s3s[0].Password())
}

func TestCache_DeleteSecret(t *testing.T) {
	t.Parallel()

	m, err := memory.New()
	assert.NoError(t, err)
	require.NotNil(t, m)

	c, err := cache.New(m, false)
	assert.NoError(t, err)
	require.NotNil(t, c)

	ctx := context.Background()

	s1, err := m.SetSecret(ctx, secrets.NewSecret("test", "test", "test"))
	assert.NoError(t, err)
	assert.NotEmpty(t, s1.ID())

	s2, err := c.GetSecret(ctx, s1.ID())
	assert.NoError(t, err)
	require.NotNil(t, s2)
	assert.Equal(t, s1.ID(), s2.ID())
	assert.Equal(t, s1.Name(), s2.Name())
	assert.Equal(t, s1.Password(), s2.Password())

	err = c.DeleteSecret(ctx, s1.ID())
	assert.NoError(t, err)

	s3, err := m.GetSecret(ctx, s1.ID())
	assert.NoError(t, err)
	require.NotNil(t, s3)
	assert.Equal(t, s1.ID(), s3.ID())
	assert.Equal(t, s1.Name(), s3.Name())
	assert.Equal(t, s1.Password(), s3.Password())
}
