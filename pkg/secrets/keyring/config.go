package keyring

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the name of the config type for the keyring secret keeper.
const ConfigType = "keyring"

// Config is the configuration for the keyring secret keeper.
type Config struct {
	// ServiceName is the name of the service to use in the keyring.
	ServiceName string `mapstructure:"service_name" yaml:"service_name"`
}

// Builder builds a new keyring secret keeper.
func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	cfg, isKeyring := c.(*Config)
	if !isKeyring {
		return nil, plugin.ErrConfig
	}

	return New(cfg.ServiceName), nil
}

func init() {
	cmd := plugin.CmdConfig{
		Short: "Configure a secret keeper that uses the system keyring",
		Fields: map[string]string{
			"service-name": "The name of the service to use in the keyring",
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			kc := config.KeeperConfig{
				"type": ConfigType,
			}

			if service, ok := fields["service-name"]; ok {
				kc["service_name"] = service
			}

			return kc, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, cmd)
}
