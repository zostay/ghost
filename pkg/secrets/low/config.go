package low

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "low"

type Config struct {
	Path string `mapstructure:"path" yaml:"path"`
}

func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	cfg, isLow := c.(*Config)
	if !isLow {
		return nil, plugin.ErrConfig
	}

	return NewLowSecurity(cfg.Path), nil
}

func init() {
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil)
}
