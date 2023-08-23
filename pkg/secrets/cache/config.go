package cache

import (
	"context"
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the type name for the cache keeper.
const ConfigType = "cache"

// Config is the configuration for the cache keeper.
type Config struct {
	// Keeper is the name of the keeper to cache.
	Keeper string `mapstructure:"keeper"`

	// TouchOnRead will cause the last modified date of secrets to be updated
	// on GetSecret* calls.
	TouchOnRead bool `mapstructure:"touch_on_read"`
}

// Builder creates a new cache keeper from the given configuration.
func Builder(ctx context.Context, c any) (secrets.Keeper, error) {
	cfg, isCache := c.(*Config)
	if !isCache {
		return nil, plugin.ErrConfig
	}

	kpr, err := keeper.Build(ctx, cfg.Keeper)
	if err != nil {
		return nil, fmt.Errorf("unable to load keeper to cache %q: %w", cfg.Keeper, err)
	}

	return New(kpr, cfg.TouchOnRead)
}

// Validate checks that the configuration is correct for the cache keeper.
// It will check that the wrapped keeper to cache exists.
func Validate(ctx context.Context, c any) error {
	cfg, isCache := c.(*Config)
	if !isCache {
		return plugin.ErrConfig
	}

	errs := plugin.NewValidationError()

	if !keeper.Exists(ctx, cfg.Keeper) {
		errs.Append(fmt.Errorf("cache keeper %q does not exist", cfg.Keeper))
	}

	return nil
}

func init() {
	var (
		keeperName  string
		touchOnRead bool
	)

	cmd := plugin.CmdConfig{
		Short: "Configure a caching keeper that wraps another keeper",
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			return config.KeeperConfig{
				"type":          ConfigType,
				"keeper":        keeperName,
				"touch_on_read": touchOnRead,
			}, nil
		},
		FlagInit: func(flags *pflag.FlagSet) error {
			flags.StringVar(&keeperName, "keeper", "", "the name of the keeper to cache")
			flags.BoolVar(&touchOnRead, "touch-on-read", false, "update the last modified date of secrets on read")

			if err := cobra.MarkFlagRequired(flags, "keeper"); err != nil {
				return err
			}

			return nil
		},
	}

	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, Validate, cmd)
}
