package keeper

import (
	"context"
	"fmt"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
)

// CheckConfig validates the configuration for all of the ghost app.
func CheckConfig(ctx context.Context, c *config.Config) error {
	errs := plugin.NewValidationError()

	ctx = WithBuilder(ctx, c)

	for k := range c.Keepers {
		err := Validate(ctx, k)
		if err != nil {
			errs.Append(err)
		}
	}

	if c.MasterKeeper != "" && !Exists(ctx, c.MasterKeeper) {
		errs.Append(fmt.Errorf("master keeper %q does not exist", c.MasterKeeper))
	}

	return errs.Return()
}
