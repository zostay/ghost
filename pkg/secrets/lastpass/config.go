package lastpass

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "lastpass"

type Config struct {
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
}

func Builder(ctx context.Context, c any) (secrets.Keeper, error) {
	cfg, isLastpass := c.(*Config)
	if !isLastpass {
		return nil, plugin.ErrConfig
	}

	kpr, err := NewLastPass(ctx, cfg.Username, cfg.Password)
	if err != nil {
		return nil, err
	}

	return kpr, nil
}

func init() {
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil)
}
