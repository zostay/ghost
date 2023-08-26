package memory_test

import (
	"testing"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/keepertest"
	"github.com/zostay/ghost/pkg/secrets/memory"
)

func TestInternal(t *testing.T) {
	t.Parallel()

	factory := func() (secrets.Keeper, error) {
		return memory.New()
	}

	ts := keepertest.New(factory)
	ts.Run(t)
}
