package keepass

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "keepass"

type Config struct {
	Path   string `mapstructure:"path" yaml:"path"`
	Master string `mapstructure:"master_password" yaml:"master_password"`
}

func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	cfg, isKeepass := c.(*Config)
	if !isKeepass {
		return nil, plugin.ErrConfig
	}

	return NewKeepass(cfg.Path, cfg.Master)
}

func init() {
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil)
}
