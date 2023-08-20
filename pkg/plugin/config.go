package plugin

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/zostay/ghost/pkg/secrets"
)

var ErrConfig = errors.New("incorrect configuration")

type BuilderFunc func(context.Context, any) (secrets.Keeper, error)
type ValidatorFunc func(context.Context, any) error

type RegisteredConfig struct {
	Config    reflect.Type
	Builder   BuilderFunc
	Validator ValidatorFunc
}

var configs = map[string]RegisteredConfig{}

func Register(
	name string,
	config reflect.Type,
	builder BuilderFunc,
	validator ValidatorFunc,
) {
	if _, ok := configs[name]; ok {
		panic(fmt.Errorf("config %q already registered", name))
	}

	if builder == nil {
		panic(fmt.Errorf("config %q has no builder", name))
	}

	if config == nil {
		panic(fmt.Errorf("config %q has no configuration type", name))
	}

	configs[name] = RegisteredConfig{
		Config:    config,
		Builder:   builder,
		Validator: validator,
	}
}

func Get(name string) (RegisteredConfig, bool) {
	cfg, hasCfg := configs[name]
	return cfg, hasCfg
}

func List() []string {
	keys := make([]string, 0, len(configs))
	for k := range configs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
