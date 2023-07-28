package low_test

import (
	"testing"

	"github.com/zostay/fssafe"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/keepertest"
	"github.com/zostay/ghost/pkg/secrets/low"
)

func TestLowSecurity(t *testing.T) {
	factory := func() (secrets.Keeper, error) {
		return low.NewLowSecurityCustom(fssafe.NewTestingLoaderSaver()), nil
	}

	ts := keepertest.New(factory)
	ts.Run(t)
}
