package keepertest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/ghost/pkg/secrets"
)

type KeeperFactory func() (secrets.Keeper, error)

type Suite struct {
	factory KeeperFactory
}

func New(f KeeperFactory) *Suite {
	return &Suite{
		factory: f,
	}
}

func (s *Suite) Run(t *testing.T) {
	t.Run("SecretKeeperGetMissingTest", s.SecretKeeperGetMissingTest)
	t.Run("SecretKeeperSetAndGet", s.SecretKeeperSetAndGet)
}

func (s *Suite) SecretKeeperGetMissingTest(t *testing.T) {
	t.Parallel()

	k, err := s.factory()
	require.NoError(t, err, "factory returns keeper")

	ctx := context.Background()

	secs, err := k.GetSecretsByName(ctx, "missing")
	assert.NoError(t, err, "missing secret returns no error")
	assert.Empty(t, secs, "missing secret is nil")
}

func (s *Suite) SecretKeeperSetAndGet(t *testing.T) {
	t.Parallel()

	k, err := s.factory()
	require.NoError(t, err, "factory returns keeper")

	ctx := context.Background()

	// create
	var sec secrets.Secret = secrets.NewSecret("set1", "username1", "secret1")
	sec, err = k.SetSecret(ctx, sec)

	require.NoError(t, err, "setting doesn't error")
	assert.NotEmpty(t, sec.ID())

	got, err := k.GetSecret(ctx, sec.ID())
	require.NoError(t, err, "getting doesn't error")
	require.NotNil(t, got, "got something")

	assert.Equal(t, sec.ID(), got.ID())
	assert.Equal(t, "set1", got.Name(), "got secret name set1")
	assert.Equal(t, "secret1", got.Password(), "got secret value secret1")

	// update
	sec = secrets.SetPassword(sec, "secret2")
	sec, err = k.SetSecret(ctx, sec)

	require.NoError(t, err, "setting again doesn't error")
	assert.NotEmpty(t, sec.ID())

	got, err = k.GetSecret(ctx, sec.ID())
	require.NoError(t, err, "getting again doesn't error")

	require.NotNil(t, got, "got something again")
	assert.Equal(t, sec.ID(), got.ID())

	assert.Equal(t, "set1", got.Name(), "got secret name still set1")
	assert.Equal(t, "secret2", got.Password(), "but got secret value changed to secret2")
}
