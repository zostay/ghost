package plugin

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/spf13/pflag"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets"
)

// ErrConfig is returned when a keeper configuration does not match the expected
// type.
var ErrConfig = errors.New("incorrect configuration")

// BuilderFunc is a factory function that should returned a constructed secret
// keeper object from the given configuration. The BuilderFunc should expect the
// configuration to be provided as the type defined in the Config field of
// RegisteredConfig. If it is not that type, it should return ErrConfig.
type BuilderFunc func(context.Context, any) (secrets.Keeper, error)

// ValidatorFunc is a function that should validate the given configuration and
// return an error if it is invalid. The ValidatorFunc should expect the
// configuration to be provided as the type defined in the Config field of
// RegisteredConfig. If it is not that type, it should return ErrConfig. It is
// recommended that the errors returned by the ValidatorFunc be build using the
// ValidationError type.
type ValidatorFunc func(context.Context, any) error

// CmdFunc is a function that gets run via the command-line interface. It is
// used to configure the secret keeper. The CmdFunc will be editing the
// configuration for the keeper with the given name. The name is provided in
// case your configuration command needs to build up the configuration
// incrementally and so it can lookup the configuration and modify it. The
// fields argument will contain those fields that have been configured via the
// Fields field of CmdConfig. These will either be a literal string value or a
// secret reference, which has the form of a map. The CmdFunc should return the
// configuration for the keeper, which will be a config.KeeperConfig, the
// unstructured configuration map.
type CmdFunc func(keeperName string, fields map[string]any) (config.KeeperConfig, error)

// FlagsFunc is a function that gets run via the command-line interface. It is
// used to add flags to the command-line interface for the configuration
// command. The flags argument is the flag set that will be used to parse the
// command-line arguments. It may return an error if there's a problem setting
// up the flags.
type FlagsFunc func(flags *pflag.FlagSet) error

// CmdConfig is the configuration for the command-line interface for the
// configuration command. The Short field is a short description of the
// configuration command. The Run field is the function that will be run when
// the command is executed. The FlagInit field is the function that will be run
// to add flags to the command-line interface. The Fields field is a map of
// additional flags that are configured to be either provided as a string value
// or as a secret reference, simplifying the configuration of these special
// fields a bit.
type CmdConfig struct {
	Short    string
	Run      CmdFunc
	FlagInit FlagsFunc
	Fields   map[string]string
}

// RegisteredConfig is the configuration for a secret keeper plugin. The Config
// field is the type of the configuration object that will be passed to the
// BuilderFunc and ValidatorFunc. The BuilderFunc is the factory function that
// will be used to construct the secret keeper. The ValidatorFunc is the
// function that will be used to validate the configuration. The CmdConfig is
// the configuration for the command-line interface for the configuration
// command.
type RegisteredConfig struct {
	Config    reflect.Type
	Builder   BuilderFunc
	Validator ValidatorFunc
	CmdConfig CmdConfig
}

var configs = map[string]RegisteredConfig{}

// Register registers a secret keeper plugin. The name is the name of the
// secret keeper. The config is the type of the configuration object that will
// be passed to the BuilderFunc and ValidatorFunc. The builder is the factory
// function that will be used to construct the secret keeper. The validator is
// the function that will be used to validate the configuration. The cmdConfig
// is the configuration for the command-line interface for the configuration
// command.
//
// This should be run in an init function in your plugin. Packages containing a
// secret keeper should be added to main.go for import side-effects. At this
// time, new plugins require a pull request or a fork of the software. Patches welcome.
//
// If there's demand for more custom plugins, I could look into a plugin
// interface like Hashcorp's or using a tool like Yaegi. ~~ zostay.
func Register(
	name string,
	config reflect.Type,
	builder BuilderFunc,
	validator ValidatorFunc,
	cmdConfig CmdConfig,
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
		CmdConfig: cmdConfig,
	}
}

// Get returns the registered configuration for the given keeper driver name. If
// the name is not registered, the second return value will be false.
func Get(name string) (RegisteredConfig, bool) {
	cfg, hasCfg := configs[name]
	return cfg, hasCfg
}

// List returns a list of all the registered keeper driver names. The list is
// sorted.
func List() []string {
	keys := make([]string, 0, len(configs))
	for k := range configs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Type returns the type of the keeper driver for the given configuration. If
// the configuration has no type or the type is not registered, the empty string
// will be returned.
func Type(c config.KeeperConfig) string {
	if typ, hasTyp := c["type"].(string); hasTyp {
		if _, isRegistered := Get(typ); isRegistered {
			return typ
		}
	}
	return ""
}
