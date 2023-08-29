package cache_test

import (
	"testing"

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
