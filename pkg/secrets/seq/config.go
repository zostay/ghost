package seq

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/spf13/pflag"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the name of the config type for the seq secret keeper.
const ConfigType = "seq"

// Config is the configuration for the seq secret keeper.
type Config struct {
	// Keepers is the list of keepers to use for the seq keeper.
	Keepers []string `mapstructure:"keepers" yaml:"keepers"`
}

// Builder constructs a new seq secret keeper.
func Builder(ctx context.Context, c any) (secrets.Keeper, error) {
	cfg, isSeq := c.(*Config)
	if !isSeq {
		return nil, plugin.ErrConfig
	}

	keepers := make([]secrets.Keeper, len(cfg.Keepers))
	for i, k := range cfg.Keepers {
		var err error
		keepers[i], err = keeper.Build(ctx, k)
		if err != nil {
			return nil, fmt.Errorf("unable to build the secret keeper named %q for the seq keeper: %w", k, err)
		}
	}
	return NewSeq(keepers...)
}

// Validator validates the seq keeper configuration.
func Validator(ctx context.Context, c any) error {
	cfg, isSeq := c.(*Config)
	if !isSeq {
		return plugin.ErrConfig
	}

	errs := plugin.NewValidationError()

	for _, k := range cfg.Keepers {
		if !keeper.Exists(ctx, k) {
			errs.Append(fmt.Errorf("seq keeper %q does not exist", k))
			continue
		}

		kpr, err := keeper.Build(ctx, k)
		if err != nil {
			return fmt.Errorf("unexpected error during validation: %w", err)
		}

		if _, isSeq := kpr.(*Seq); isSeq {
			errs.Append(fmt.Errorf("seq keeper %q is also a seq, seq keepers inside of seq keepers are not permitted", k))
		}
	}

	return errs.Return()
}

func init() {
	var (
		modKeepers                   []string
		appendKeepers, deleteKeepers bool
	)
	cmd := plugin.CmdConfig{
		Short: "Configure a sequential secret keeper",
		FlagInit: func(flags *pflag.FlagSet) error {
			flags.StringSliceVarP(&modKeepers, "keepers", "k", []string{}, "The list of keepers to make changes with")
			flags.BoolVarP(&appendKeepers, "append", "a", false, "Append the named keepers to the end of the sequence (default is to replace)")
			flags.BoolVarP(&deleteKeepers, "delete", "d", false, "Remove the named keepers from the sequence")
			return nil
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			c := config.Instance()
			kc := c.Keepers[keeperName]
			if kc == nil {
				kc = map[string]any{
					"type":    ConfigType,
					"keepers": []string{},
				}
			}

			if appendKeepers && deleteKeepers {
				return nil, errors.New("cannot append and delete at the same time")
			}

			if appendKeepers {
				keepers := kc["keepers"].([]any)
				for _, k := range modKeepers {
					keepers = append(keepers, k)
				}
				kc["keepers"] = keepers
				return kc, nil
			}

			if deleteKeepers {
				keepers := kc["keepers"].([]any)
				for _, k := range modKeepers {
					for {
						ix := slices.FirstIndex(keepers, func(v any) bool {
							return v.(string) == k
						})
						if ix < 0 {
							break
						}
						keepers = slices.Delete(keepers, ix)
					}
				}
				kc["keepers"] = keepers
			}

			kc["keepers"] = modKeepers

			return kc, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, Validator, cmd)
}
