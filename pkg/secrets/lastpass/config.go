package lastpass

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the name of the config type for the lastpass secret keeper.
const ConfigType = "lastpass"

// Config is the configuration for the lastpass secret keeper.
type Config struct {
	// Username is the username to use to log into LastPass.
	Username string `mapstructure:"username" yaml:"username"`
	// Password is the password to use to log into LastPass.
	Password string `mapstructure:"password" yaml:"password"`
}

// Builder builds a new lastpass secret keeper.
func Builder(ctx context.Context, c any) (secrets.Keeper, error) {
	cfg, isLastpass := c.(*Config)
	if !isLastpass {
		return nil, plugin.ErrConfig
	}

	kpr, err := NewLastPass(ctx, cfg.Username, cfg.Password)
	if err != nil {
		return nil, err
	}

	return kpr, nil
}

// Print prints the configuration for the lastpass secret keeper.
func Print(c any, w io.Writer) error {
	cfg, isLastpass := c.(*Config)
	if !isLastpass {
		return plugin.ErrConfig
	}

	fmt.Fprintln(w, "username:", cfg.Username)
	passwordVal := "<not set>"
	if cfg.Password != "" {
		passwordVal = "<hidden>"
	}
	fmt.Fprintln(w, "password:", passwordVal)
	return nil
}

func init() {
	cmd := plugin.CmdConfig{
		Short: "Configure a LastPass secret keeper",
		Fields: map[string]string{
			"username": "The username to use to log into LastPass",
			"password": "The password to use to log into LastPass",
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			kc := config.KeeperConfig{
				"type": ConfigType,
			}

			if username, ok := fields["username"]; ok {
				kc["username"] = username
			}

			if password, ok := fields["password"]; ok {
				kc["password"] = password
			}

			return kc, nil
		},
	}
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, Print, cmd)
}
