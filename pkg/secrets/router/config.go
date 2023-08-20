package router

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/pflag"
	"github.com/zostay/go-std/set"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/config"
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
	var (
		removeLocations []string
		addLocations    []string
		addKeeper       string
		defaultKeeper   string
	)

	cmd := plugin.CmdConfig{
		Short: "Configure a router secret keeper",
		FlagInit: func(flags *pflag.FlagSet) error {
			flags.StringSliceVar(&removeLocations, "remove", []string{}, "Remove one or more locations from the router")
			flags.StringSliceVar(&addLocations, "add", []string{}, "Add one or more locations to the router")
			flags.StringVar(&addKeeper, "keeper", "", "Keeper to use with to the added locations")
			flags.StringVar(&defaultKeeper, "default", "", "Default keeper to use with the router")
			return nil
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			if len(removeLocations) > 0 && len(addLocations) > 0 {
				return nil, errors.New("cannot remove and add locations in the same step")
			}

			if len(removeLocations) > 0 && defaultKeeper != "" {
				return nil, errors.New("cannot remove locations and set the default keeper in the same step")
			}

			if len(removeLocations) > 0 && addKeeper != "" {
				return nil, errors.New("cannot specify keeper while removing locaitons")
			}

			if len(addLocations) > 0 && addKeeper == "" {
				return nil, errors.New("must specify a keeper to use with the added locations")
			}

			c := config.Instance()
			kc := c.Keepers[keeperName]
			if kc == nil {
				kc = map[string]any{
					"type":   ConfigType,
					"routes": []map[string]any{},
				}
			}

			if defaultKeeper != "" {
				kc["default"] = defaultKeeper
			}

			if len(removeLocations) > 0 {
				RemoveLocationsAndRoutes(kc, removeLocations...)
				return kc, nil
			}

			AddRoute(kc, addKeeper, addLocations...)

			return kc, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, cmd)
}

func AddRoute(rc config.KeeperConfig, keeper string, locations ...string) {
	routes := rc["routes"].([]map[string]any)
	if routes == nil {
		routes = make([]map[string]any, 0, 1)
	}

	var foundRoute map[string]any
	for _, r := range routes {
		if r["keeper"] == keeper {
			foundRoute = r
			break
		}
	}

	if foundRoute != nil {
		for _, loc := range locations {
			ix := slices.FirstIndex(foundRoute["locations"].([]any), func(v any) bool {
				return v.(string) == loc
			})
			if ix < 0 {
				foundRoute["locations"] = append(foundRoute["locations"].([]any), loc)
			}
		}

		return
	}

	routes = append(routes, map[string]any{
		"locations": locations,
		"keeper":    keeper,
	})

	rc["routes"] = routes
}

func RemoveLocationsAndRoutes(rc config.KeeperConfig, removeLocations ...string) {
	routes := rc["routes"].([]map[string]any)
	removeSet := set.New(removeLocations...)

	deleteRoutes := []int{}
	for i, r := range routes {
		locations := r["locations"].([]any)
		for _, loc := range locations {
			if removeSet.Contains(loc.(string)) {
				for {
					ix := slices.FirstIndex(locations, func(v any) bool {
						return v.(string) == loc.(string)
					})
					if ix < 0 {
						break
					}
					locations = slices.Delete(locations, ix)
					if len(locations) == 0 {
						deleteRoutes = append(deleteRoutes, i)
						break
					}
				}
				routes[i]["locations"] = locations
			}
		}
	}

	sort.Sort(sort.Reverse(sort.IntSlice(deleteRoutes)))

	for _, i := range deleteRoutes {
		routes = slices.Delete(routes, i)
	}

	rc["routes"] = routes
}
