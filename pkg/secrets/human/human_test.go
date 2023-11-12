package human_test

import (
	"testing"

	"github.com/zostay/ghost/pkg/secrets"
	"github.com/zostay/ghost/pkg/secrets/human"
	"github.com/zostay/ghost/pkg/secrets/keepertest"
)

func TestHuman(t *testing.T) {
	t.Skip("normally don't test this because it requires feedback from the user")

	factory := func() (secrets.Keeper, error) {
		h := human.New()
		h.AddQuestion("test question", []string{"password"}, map[string]string{
			"username": "bob",
		})
		return h, nil
	}

	ts := keepertest.New(factory)
	ts.AddGetPreset(secrets.NewSecret("test question", "bob", "OK",
		secrets.WithID("test question")))
	ts.RunWithPresets(t)
}
