package low

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
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
	cmd := plugin.CmdConfig{
		Short: "Configure a low-security secret keeper",
		Fields: map[string]string{
			"path": "Path to the low-level configuration file",
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			kc := config.KeeperConfig{
				"type": ConfigType,
			}

			if path, ok := fields["path"]; ok {
				kc["path"] = path
			}

			return kc, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, cmd)
}
