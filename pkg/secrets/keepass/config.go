package keepass

import (
	"context"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "keepass"

type Config struct {
	Path   string `mapstructure:"path" yaml:"path"`
	Master string `mapstructure:"master_password" yaml:"master_password"`
}

func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	cfg, isKeepass := c.(*Config)
	if !isKeepass {
		return nil, plugin.ErrConfig
	}

	return NewKeepass(cfg.Path, cfg.Master)
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
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, cmd)
}
