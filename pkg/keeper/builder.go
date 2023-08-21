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

	err = mb.resolveSecretRefsInMap(kc, true)
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

	err = mb.resolveSecretRefsInMap(kc, true)
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

	err = mb.resolveSecretRefsInMap(kc, false)
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

func (mb *builderContext) resolveSecretRefsInMap(kc config.KeeperConfig, lookup bool) error {
	for k, v := range kc {
		if k == config.SecretRefKey {
			var ref config.SecretRef
			err := mapstructure.Decode(v, &ref)
			if err != nil {
				return err
			}

			if ref.KeeperName == "" {
				return errors.New("malformed secret reference: keeper is empty")
			}

			if mb.c.Keepers[ref.KeeperName] == nil {
				return fmt.Errorf("malformed secret reference: keeper %q does not exist", ref.KeeperName)
			}

			if ref.SecretName == "" {
				return errors.New("malformed secret reference: secret is empty")
			}

			if ref.Field == "" {
				return errors.New("malformed secret reference: field is empty")
			}

			if lookup {
				kc[k] = "<secret-placeholder>"
				continue
			}
		}

		if vMap, isMap := v.(map[string]any); isMap {
			err := mb.resolveSecretRefsInMap(vMap, lookup)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
