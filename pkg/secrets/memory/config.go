package memory

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the name of the config type for the memory secret keeper.
const ConfigType = "memory"

// Config is the configuration required for the secret store.
type Config struct{}

// Builder constructs a new internal secret store.
func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	_, isInternal := c.(*Config)
	if !isInternal {
		return nil, plugin.ErrConfig
	}

	return New()
}

func init() {
	cmd := plugin.CmdConfig{
		Short: "Configure an in-memory, temporary secret keeper",
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			return config.KeeperConfig{"type": ConfigType}, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, cmd)
}
