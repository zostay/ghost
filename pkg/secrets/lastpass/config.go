package lastpass

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
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
	cmd := plugin.CmdConfig{
		Short: "Configure a LastPass secret keeper",
		Fields: map[string]string{
			"username": "The username to use to log into LastPass",
			"password": "The password to use to log into LastPass",
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			kc := config.KeeperConfig{
				"type": ConfigType,
			}

			if username, ok := fields["username"]; ok {
				kc["username"] = username
			}

			if password, ok := fields["password"]; ok {
				kc["password"] = password
			}

			return kc, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, cmd)
}
