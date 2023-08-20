package http

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

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
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil)
}
