package low

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the name of the config type for the low-security secret keeper.
const ConfigType = "low"

// Config is the configuration for the low-security secret keeper.
type Config struct {
	// Path is the path to the low-level configuration file.
	Path string `mapstructure:"path" yaml:"path"`
}

// Builder builds a new low-security secret keeper.
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
