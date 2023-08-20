package memory_test

import (
	"testing"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/keepertest"
)

func TestInternal(t *testing.T) {
	t.Parallel()

	factory := func() (secrets.Keeper, error) {
		return memory.NewInternal()
	}

	ts := keepertest.New(factory)
	ts.Run(t)
}
