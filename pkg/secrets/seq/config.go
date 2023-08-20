package seq

import (
	"context"
	"fmt"
	"reflect"

	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "seq"

type Config struct {
	Keepers []string `mapstructure:"keepers" yaml:"keepers"`
}

func Builder(ctx context.Context, c any) (secrets.Keeper, error) {
	cfg, isSeq := c.(*Config)
	if !isSeq {
		return nil, plugin.ErrConfig
	}

	keepers := make([]secrets.Keeper, len(cfg.Keepers))
	for i, k := range cfg.Keepers {
		var err error
		keepers[i], err = keeper.Build(ctx, k)
		if err != nil {
			return nil, fmt.Errorf("unable to build the secret keeper named %q for the seq keeper: %w", k, err)
		}
	}
	return NewSeq(keepers...)
}

func Validator(ctx context.Context, c any) error {
	cfg, isSeq := c.(*Config)
	if !isSeq {
		return plugin.ErrConfig
	}

	errs := plugin.NewValidationError()

	for _, k := range cfg.Keepers {
		if !keeper.Exists(ctx, k) {
			errs.Append(fmt.Errorf("seq keeper %q does not exist", k))
			continue
		}

		kpr, err := keeper.Build(ctx, k)
		if err != nil {
			return fmt.Errorf("unexpected error during validation: %w", err)
		}

		if _, isSeq := kpr.(*Seq); isSeq {
			errs.Append(fmt.Errorf("seq keeper %q is also a seq, seq keepers inside of seq keepers are not permitted", k))
		}
	}

	return errs.Return()
}

func init() {
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, Validator)
}
