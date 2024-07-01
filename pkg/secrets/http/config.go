package http

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the type name for the HTTP secrets keeper.
const (
	ConfigType  = "http"
	ServiceName = "ghost.keeper"
)

// Config is the configuration of the HTTP secrets keeper.
type Config struct{}

// BuildServiceClient builds the gRPC client for the HTTP secrets keeper.
func BuildServiceClient() (KeeperClient, error) {
	sockName := MakeHttpServerSocketName()
	clientConn, err := grpc.NewClient("unix:"+sockName,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return NewKeeperClient(clientConn), nil
}

// Builder is the builder function for the HTTP secrets keeper.
func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	_, isGrpc := c.(*Config)
	if !isGrpc {
		return nil, plugin.ErrConfig
	}

	client, err := BuildServiceClient()
	if err != nil {
		return nil, err
	}

	return NewClient(client), nil
}

func init() {
	cmd := plugin.CmdConfig{
		Short: "Configure an HTTP secret keeper",
		Run: func(keeperName string, _ map[string]any) (config.KeeperConfig, error) {
			return config.KeeperConfig{
				"type": ConfigType,
			}, nil
		},
	}

	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, nil, cmd)
}

func MakeHttpServerSocketName() string {
	tmp := os.TempDir()
	uid := os.Getuid()
	return filepath.Join(tmp, fmt.Sprintf("%s.%d", ServiceName, uid))
}
