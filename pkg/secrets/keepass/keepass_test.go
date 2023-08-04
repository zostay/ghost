package keepass_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zostay/fssafe"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/keepass"
	"github.com/zostay/ghost/pkg/secrets/keepertest"
)

func TestKeepass(t *testing.T) {
	t.Parallel()

	lss := make([]*fssafe.TestingLoaderSaver, 0)

	factory := func() (secrets.Keeper, error) {
		k, err := keepass.NewKeepassNoVerify("", "testing123")
		if !assert.NoError(t, err, "no error getting keepass") {
			return nil, err
		}

		ls := fssafe.NewTestingLoaderSaver()
		lss = append(lss, ls)
		k.LoaderSaver = ls

		return k, nil
	}

	ts := keepertest.New(factory)
	ts.Run(t)

	for _, ls := range lss {
		for i, r := range ls.Readers {
			assert.Truef(t, r.Closed, "reader %d was closed", i)
		}
		for i, w := range ls.Writers {
			assert.True(t, w.Closed, "writer %d was closed", i)
		}
	}
}
