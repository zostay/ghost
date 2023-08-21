package keeper

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

type builderKey struct{}

type builderContext struct {
	ctx context.Context
	c   *config.Config
}

func WithBuilder(ctx context.Context, c *config.Config) context.Context {
	return &builderContext{
		ctx: ctx,
		c:   c,
	}
}

func (mb *builderContext) Deadline() (time.Time, bool) {
	return mb.ctx.Deadline()
}

func (mb *builderContext) Done() <-chan struct{} {
	return mb.ctx.Done()
}

func (mb *builderContext) Err() error {
	return mb.ctx.Err()
}

func (mb *builderContext) Value(key any) any {
	bk := builderKey{}
	if key == bk {
		return mb
	}
	return mb.ctx.Value(key)
}

func Build(ctx context.Context, name string) (secrets.Keeper, error) {
	builder, isBuilder := ctx.Value(builderKey{}).(*builderContext)
	if builder.c.Keepers[name] == nil {
		return nil, fmt.Errorf("secret keeper %q: unable to find the secret keeper factory in context", name)
	}

	if isBuilder {
		err := builder.Validate(name)
		if err != nil {
			return nil, fmt.Errorf("secret keeper %q: %w", name, err)
		}

		return builder.Build(name)
	}

	return nil, errors.New("unable to find the secret keeper factory in context")
}

func Validate(ctx context.Context, name string) error {
	builder, isBuilder := ctx.Value(builderKey{}).(*builderContext)
	if builder.c.Keepers[name] == nil {
		return fmt.Errorf("secret keeper %q: unable to find the secret keeper factory in context", name)
	}

	if isBuilder {
		err := builder.Validate(name)
		if err != nil {
			return fmt.Errorf("secret keeper %q: %w", name, err)
		}

		return nil
	}

	return fmt.Errorf("secret keeper %q: unable to find the secret keeper factory in context", name)
}

func Exists(ctx context.Context, name string) bool {
	builder, isBuilder := ctx.Value(builderKey{}).(*builderContext)
	if !isBuilder {
		panic("unable to find the secret keeper factory in context")
	}

	return builder.c.Keepers[name] != nil
}

func Decode(ctx context.Context, name string) (any, error) {
	builder, isBuilder := ctx.Value(builderKey{}).(*builderContext)
	if !isBuilder {
		panic("unable to find the secret keeper factory in context")
	}

	return builder.Decode(name)
}

func (mb *builderContext) Decode(name string) (any, error) {
	kc, typBuilder, err := mb.configAndBuilder(name)
	if err != nil {
		return nil, err
	}

	_, err = mb.resolveSecretRefsInMap(kc, true)
	if err != nil {
		return nil, err
	}

	cfg := reflect.New(typBuilder.Config).Interface()
	err = mapstructure.Decode(kc, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to structure configuration for %q: %w", name, err)
	}

	return cfg, nil
}

func (mb *builderContext) Build(name string) (secrets.Keeper, error) {
	kc, typBuilder, err := mb.configAndBuilder(name)
	if err != nil {
		return nil, err
	}

	_, err = mb.resolveSecretRefsInMap(kc, true)
	if err != nil {
		return nil, err
	}

	cfg := reflect.New(typBuilder.Config).Interface()
	err = mapstructure.Decode(kc, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to structure configuration for %q: %w", name, err)
	}

	return typBuilder.Builder(mb.ctx, cfg)
}

func (mb *builderContext) Validate(name string) error {
	kc, typBuilder, err := mb.configAndBuilder(name)
	if err != nil {
		return err
	}

	_, err = mb.resolveSecretRefsInMap(kc, false)
	if err != nil {
		return err
	}

	cfg := reflect.New(typBuilder.Config).Interface()
	err = mapstructure.Decode(kc, cfg)
	if err != nil {
		return fmt.Errorf("unable to structure configuration for %q: %w", name, err)
	}

	if typBuilder.Validator == nil {
		return nil
	}

	return typBuilder.Validator(mb.ctx, cfg)
}

func (mb *builderContext) configAndBuilder(name string) (kc config.KeeperConfig, r plugin.RegisteredConfig, err error) {
	kc = mb.c.Keepers[name]
	if kc == nil {
		err = fmt.Errorf("no configuration for keeper named %q", name)
		return
	}

	typ := kc["type"].(string)
	var hasPlugin bool
	r, hasPlugin = plugin.Get(typ)
	if !hasPlugin {
		err = fmt.Errorf("keeper configuration for %q has incorrect or unregistered type", name, typ)
		return
	}

	return
}

func (mb *builderContext) resolveSecretRefsInMap(kc config.KeeperConfig, lookup bool) (any, error) {
	for k, v := range kc {
		if k == config.SecretRefKey {
			var ref config.SecretRef
			err := mapstructure.Decode(v, &ref)
			if err != nil {
				return nil, err
			}

			if ref.KeeperName == "" {
				return nil, errors.New("malformed secret reference: keeper is empty")
			}

			if mb.c.Keepers[ref.KeeperName] == nil {
				return nil, fmt.Errorf("malformed secret reference: keeper %q does not exist", ref.KeeperName)
			}

			if ref.SecretName == "" {
				return nil, errors.New("malformed secret reference: secret is empty")
			}

			if ref.Field == "" {
				return nil, errors.New("malformed secret reference: field is empty")
			}

			if !lookup {
				return "<secret-placeholder>", nil
			}

			kpr, err := mb.Build(ref.KeeperName)
			if err != nil {
				return nil, fmt.Errorf("unable to perform lookup with keeper %q: %w", ref.KeeperName, err)
			}

			sec, err := kpr.GetSecret(mb, ref.SecretName)
			if err != nil {
				return nil, fmt.Errorf("unable to perform secret lookup with keeper %q and secret %q: %w", ref.KeeperName, ref.SecretName, err)
			}

			switch ref.Field {
			case "id":
				return sec.ID(), nil
			case "username":
				return sec.Username(), nil
			case "password":
				return sec.Password(), nil
			case "type":
				return sec.Type(), nil
			case "url":
				return sec.Url().String(), nil
			default:
				return sec.GetField(ref.Field), nil
			}
		}

		if vMap, isMap := v.(config.KeeperConfig); isMap {
			lVal, err := mb.resolveSecretRefsInMap(vMap, lookup)
			if err != nil {
				return nil, err
			}

			kc[k] = lVal
		}
	}

	return kc, nil
}
