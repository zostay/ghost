package keyring

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "keyring"

type Config struct {
	ServiceName string `mapstructure:"service_name" yaml:"service_name"`
}

func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	cfg, isKeyring := c.(*Config)
	if !isKeyring {
		return nil, plugin.ErrConfig
	}

	return New(cfg.ServiceName), nil
}

func init() {
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil)
}
