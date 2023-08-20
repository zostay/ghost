package router

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "router"

type Config struct {
	Routes       []RouteConfig `mapstructure:"routes" yaml:"routes"`
	DefaultRoute string        `mapstructure:"default" yaml:"default,omitempty"`
}

type RouteConfig struct {
	Locations []string `mapstructure:"locations" yaml:"locations"`
	Keeper    string   `mapstructure:"keeper" yaml:"keeper"`
}

func Builder(ctx context.Context, c any) (secrets.Keeper, error) {
	cfg, isRouter := c.(*Config)
	if !isRouter {
		return nil, plugin.ErrConfig
	}

	defaultKeeper, err := keeper.Build(ctx, cfg.DefaultRoute)
	if err != nil {
		return nil, fmt.Errorf("unable to build the secret keeper named %q for the default route: %w", cfg.DefaultRoute, err)
	}

	r := NewRouter(defaultKeeper)
	for _, rt := range cfg.Routes {
		keeper, err := keeper.Build(ctx, rt.Keeper)
		if err != nil {
			locs := strings.Join(rt.Locations, ",")
			return nil, fmt.Errorf("unable to build the secret keeper named %q for the route to %q: %w", rt.Keeper, locs, err)
		}

		err = r.AddKeeper(keeper, rt.Locations...)
		if err != nil {
			locs := strings.Join(rt.Locations, ",")
			return nil, fmt.Errorf("unable to add a route for the secret keeper named %q for the route to %q: %w", rt.Keeper, locs, err)
		}
	}
	return r, nil
}

func Validator(ctx context.Context, c any) error {
	cfg, isRouter := c.(*Config)
	if !isRouter {
		return plugin.ErrConfig
	}

	errs := plugin.NewValidationError()

	if cfg.DefaultRoute != "" {
		if keeper.Exists(ctx, cfg.DefaultRoute) {
			errs.Append(fmt.Errorf("default route keeper %q does not exist", cfg.DefaultRoute))
		}
	}

	for _, r := range cfg.Routes {
		if keeper.Exists(ctx, r.Keeper) {
			errs.Append(fmt.Errorf("route keeper %q does not exist", r.Keeper))
		}

		if len(r.Locations) == 0 {
			errs.Append(fmt.Errorf("route keeper %q has no locations", r.Keeper))
		}
	}

	return errs.Return()
}

func init() {
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil)
}
