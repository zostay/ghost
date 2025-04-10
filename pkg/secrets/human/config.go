package human

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/spf13/pflag"
	"github.com/zostay/go-std/maps"
	"github.com/zostay/go-std/set"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the type name for the human secrets keeper.
const ConfigType = "human"

// QuestionConfig is the configuration for a single question. Each question
// behaves as a secret. The contents of the secret are determined by this
// configuration and user input.
type QuestionConfig struct {
	// ID is the unique identifier for the question. It is used to reference
	// the question in the configuration.
	ID string `mapstructure:"id" yaml:"id"`

	// Presets are the values that will be used to populate the parts of the
	// secret that are not queried directly from the user via pinentry.
	Presets map[string]string `mapstructure:"presets" yaml:"presets"`

	// AskFor is the list of fields that will be queried from the user via
	// pinentry.
	AskFor []string `mapstructure:"ask_for" yaml:"ask_for"`
}

// Config is the configuration of the human secrets keeper.
type Config struct {
	// Questions is the list of questions that will be asked of the user.
	Questions []QuestionConfig `mapstructure:"questions" yaml:"questions"`
}

// Validator checks that the configuration is correct for the human secrets
// keeper. It will check that every question has at least one field to ask
// for and that no question has a preset that is already being asked for.
func Validator(_ context.Context, c any) error {
	cfg, isHuman := c.(*Config)
	if !isHuman {
		return plugin.ErrConfig
	}

	errs := plugin.NewValidationError()

	for _, q := range cfg.Questions {
		if len(q.AskFor) == 0 {
			errs.Append(fmt.Errorf("human question %q asks for nothing", q.ID))
		}

		flds := set.New[string](maps.Keys(q.Presets)...)
		for _, f := range q.AskFor {
			if flds.Contains(f) {
				errs.Append(fmt.Errorf("human question %q configuration already contains field named %q", q.ID, f))
			}
			flds.Insert(f)
		}
	}

	return errs.Return()
}

// Builder creates a new human secrets keeper from the given configuration.
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

// Print prints the configuration of the human secrets keeper.
func Print(c any, w io.Writer) error {
	cfg, isHuman := c.(*Config)
	if !isHuman {
		return plugin.ErrConfig
	}

	fmt.Fprintln(w, "questions:")
	for _, q := range cfg.Questions {
		fmt.Fprintln(w, "- id:", q.ID)
		if len(q.Presets) > 0 {
			fmt.Fprintln(w, "  presets:")
			for k, v := range q.Presets {
				fmt.Fprintln(w, "   ", k, "=", v)
			}
		}
		fmt.Fprintln(w, "asking for:", strings.Join(q.AskFor, ", "))
	}
	return nil
}

func init() {
	var (
		setQuestion, removeQuestion string
		askFor                      []string
		presets                     map[string]string
	)

	cmd := plugin.CmdConfig{
		Short: "Configure a human secret keeper",
		Fields: map[string]string{
			"preset-username": "Set a preset username",
			"preset-password": "Set a preset password",
		},
		FlagInit: func(flags *pflag.FlagSet) error {
			flags.StringSliceVar(&askFor, "ask-for", []string{}, "Ask for a secret value")
			flags.StringToStringVar(&presets, "preset", map[string]string{}, "Set a preset value")
			flags.StringVar(&setQuestion, "set", "", "Add or update a secret value with the given ID")
			flags.StringVar(&removeQuestion, "remove", "", "Remove a secret value with the given ID")
			return nil
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			if setQuestion != "" && removeQuestion != "" {
				return nil, fmt.Errorf("cannot set and remove a secret value in the same step")
			}

			if setQuestion == "" && removeQuestion == "" {
				return nil, fmt.Errorf("you must set or remove a secret value with this command")
			}

			if removeQuestion != "" && (len(askFor) > 0 || len(presets) > 0 || fields["preset-username"] != nil || fields["preset-password"] != nil) {
				return nil, fmt.Errorf("--remove is incompatible with --ask-for and preset flags")
			}

			if setQuestion != "" && len(askFor) == 0 {
				return nil, fmt.Errorf("--set requires --ask-for")
			}

			c := config.Instance()
			kc := c.Keepers[keeperName]
			if kc == nil {
				kc = map[string]any{
					"type":      ConfigType,
					"questions": []any{},
				}
			}

			presetFields := make(map[string]any, len(presets)+len(fields))
			for k, v := range presets {
				presetFields[k] = v
			}
			for k, v := range fields {
				presetFields[strings.TrimPrefix(k, "preset-")] = v
			}

			if removeQuestion != "" {
				RemoveQuestion(kc, removeQuestion)
			}

			SetQuestion(kc, setQuestion, presetFields, askFor)

			return kc, nil
		},
	}

	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, Validator, Print, cmd)
}

// SetQuestion adds or updates a question in the given keeper configuration.
// This is used by the command-line interface to help build up the configuration
// incrementally.
func SetQuestion(
	kc config.KeeperConfig,
	id string,
	presets map[string]any,
	askFor []string,
) {
	qs := kc["questions"].([]map[string]any)
	if qs == nil {
		qs = []map[string]any{}
	}

	qs = append(qs, map[string]any{
		"id":      id,
		"presets": presets,
		"ask_for": askFor,
	})

	kc["questions"] = qs
}

// RemoveQuestion removes a question from the given keeper configuration. This
// is used by the command-line interface to help remove configuration
// incrementally.
func RemoveQuestion(kc config.KeeperConfig, id string) {
	qs := kc["questions"].([]map[string]any)
	if qs == nil {
		return
	}

	i := slices.FirstIndex(qs, func(q map[string]any) bool {
		return q["id"] == id
	})
	if i >= 0 {
		qs = slices.Delete(qs, i)
	}

	kc["questions"] = qs
}
