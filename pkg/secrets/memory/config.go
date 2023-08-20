package memory

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

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
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil)
}
