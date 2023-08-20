package human

import (
	"context"
	"fmt"
	"reflect"

	"github.com/zostay/go-std/maps"
	"github.com/zostay/go-std/set"

	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "human"

type QuestionConfig struct {
	ID      string            `mapstructure:"id" yaml:"id"`
	Presets map[string]string `mapstructure:"presets" yaml:"presets"`
	AskFor  []string          `mapstructure:"ask_for" yaml:"ask_for"`
}

type Config struct {
	Questions []QuestionConfig `mapstructure:"questions" yaml:"questions"`
}

func Validator(_ context.Context, c any) error {
	cfg, isHuman := c.(*Config)
	if !isHuman {
		return plugin.ErrConfig
	}

	errs := plugin.NewValidationError()

	for _, q := range cfg.Questions {
		flds := set.New[string](maps.Keys(q.Presets)...)
		for _, f := range q.AskFor {
			if !flds.Contains(f) {
				errs.Append(fmt.Errorf("human question configuration already contains field named %q", f))
			}
			flds.Insert(f)
		}
	}

	return errs.Return()
}

func Builder(_ context.Context, c any) (secrets.Keeper, error) {
	cfg, isHuman := c.(*Config)
	if !isHuman {
		return nil, plugin.ErrConfig
	}

	kpr := New()
	for _, q := range cfg.Questions {
		kpr.AddQuestion(
			q.ID,
			q.AskFor,
			q.Presets,
		)
	}

	return kpr, nil
}

func init() {
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, Validator)
}
