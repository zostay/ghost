package onepassword

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the name of the config type for the 1password secret keeper.
const ConfigType = "1password"

// Config is the configuration for the 1password secret keeper.
type Config struct {
	// ConnectHost is the host to connect to the 1password service.
	ConnectHost string `mapstructure:"connect_host" yaml:"connect_host"`

	// ConnectToken is the token to use to connect to the 1password service.
	ConnectToken string `mapstructure:"connect_token" yaml:"connect_token"`
}

// Builder builds a new 1password secret keeper.
func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	cfg, is1password := c.(*Config)
	if !is1password {
		return nil, plugin.ErrConfig
	}

	kpr := NewOnePassword(cfg.ConnectHost, cfg.ConnectToken)

	return kpr, nil
}

// Print prints the configuration for the 1password secret keeper.
func Print(c any, w io.Writer) error {
	cfg, is1password := c.(*Config)
	if !is1password {
		return plugin.ErrConfig
	}

	fmt.Fprintln(w, "connect host:", cfg.ConnectHost)
	tokenVal := "<not set>"
	if cfg.ConnectToken != "" {
		tokenVal = "<hidden>"
	}
	fmt.Fprintln(w, "connect token:", tokenVal)
	return nil
}

func init() {
	cmd := plugin.CmdConfig{
		Short: "Configure a 1Password secret keeper (requires a Connect Server)",
		Fields: map[string]string{
			"connect_host":  "The host to connect to the 1password connect service",
			"connect_token": "The token to use to connect to the 1password connect service",
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			kc := config.KeeperConfig{
				"type": ConfigType,
			}

			if connectHost, ok := fields["connect_host"]; ok {
				kc["connect_host"] = connectHost
			}

			if connectToken, ok := fields["connect_token"]; ok {
				kc["connect_token"] = connectToken
			}

			return kc, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, Print, cmd)
}
