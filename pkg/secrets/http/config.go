package http

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the type name for the HTTP secrets keeper.
const ConfigType = "http"

// Config is the configuration of the HTTP secrets keeper.
type Config struct{}

// Builder is the builder function for the HTTP secrets keeper.
func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	_, isGrpc := c.(*Config)
	if !isGrpc {
		return nil, plugin.ErrConfig
	}

	return NewClient(), nil
}

func init() {
	cmd := plugin.CmdConfig{
		Short: "Configure an HTTP secret keeper",
		Run: func(keeperName string, _ map[string]any) (config.KeeperConfig, error) {
			return config.KeeperConfig{
				"type": ConfigType,
			}, nil
		},
	}

	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, cmd)
}
