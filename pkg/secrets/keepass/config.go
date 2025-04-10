package keepass

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/mitchellh/go-homedir"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the name of the config type for the Keepass secret keeper.
const ConfigType = "keepass"

// Config is the configuration for the Keepass secret keeper.
type Config struct {
	// Path is the path to the Keepass database.
	Path string `mapstructure:"path" yaml:"path"`
	// Master is the master password to use to unlock the Keepass database.
	Master string `mapstructure:"master_password" yaml:"master_password"`
}

// Builder builds a new Keepass secret keeper.
func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	cfg, isKeepass := c.(*Config)
	if !isKeepass {
		return nil, plugin.ErrConfig
	}

	path, err := homedir.Expand(os.ExpandEnv(cfg.Path))
	if err != nil {
		return nil, err
	}

	return NewKeepass(path, cfg.Master)
}

// Print prints the configuration for the Keepass secret keeper.
func Print(c any, w io.Writer) error {
	cfg, isKeepass := c.(*Config)
	if !isKeepass {
		return plugin.ErrConfig
	}

	fmt.Fprintln(w, "path:", cfg.Path)
	masterVal := "<not set>"
	if cfg.Master != "" {
		masterVal = "<hidden>"
	}
	fmt.Fprintln(w, "master:", masterVal)
	return nil
}

func init() {
	cmd := plugin.CmdConfig{
		Short: "Configure a Keepass secret keeper",
		Fields: map[string]string{
			"path":            "Path to the Keepass database",
			"master-password": "The master password to use to unlock the Keepass database",
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			kc := config.KeeperConfig{
				"type": ConfigType,
			}

			if path, ok := fields["path"]; ok {
				kc["path"] = path
			}

			if master, ok := fields["master-password"]; ok {
				kc["master_password"] = master
			}

			return kc, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, Print, cmd)
}
