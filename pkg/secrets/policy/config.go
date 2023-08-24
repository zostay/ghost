package policy

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/pflag"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

// ConfigType is the name of the config type for the policy secret keeper.
const ConfigType = "policy"

// RuleConfig configures the action to apply with a rule.
type RuleConfig struct {
	// Lifetime is the maximum lifetime of a secret in the keeper.
	Lifetime time.Duration `mapstructure:"lifetime" yaml:"lifetime"`
	// Acceptance determines whether access to the secret is allowed or denied.
	Acceptance string `mapstructure:"acceptance" yaml:"acceptance"`
}

// MatchConfig configures the matchers for a rule.
type MatchConfig struct {
	// LocationMatch is a matches a rule by location by exact match, glob, or
	// regular expression.
	LocationMatch string `mapstructure:"location" yaml:"location"`
	// NameMatch is a matches a rule by name by exact match, glob, or regular
	// expression.
	NameMatch string `mapstructure:"name" yaml:"name"`
	// UsernameMatch is a matches a rule by username by exact match, glob, or
	// regular expression.
	UsernameMatch string `mapstructure:"username" yaml:"username"`
	// TypeMatch is a matches a rule by secret type by exact match, glob, or
	// regular expression.
	TypeMatch string `mapstructure:"secret_type" yaml:"secret_type"`
	// UrlMatch is a matches a rule by URL by exact match, glob, or regular
	// expression.
	UrlMatch string `mapstructure:"url" yaml:"url"`
}

// MatchRuleConfig configures a rule with matchers.
type MatchRuleConfig struct {
	// MatchConfig configures the matchers for a rule.
	MatchConfig `mapstructure:",squash" yaml:",inline"`
	// RuleConfig configures the action to apply with a rule.
	RuleConfig `mapstructure:",squash" yaml:",inline"`
}

// Config is the configuration for the policy secret keeper.
type Config struct {
	// Keeper is the name of the keeper to wrap.
	Keeper string `mapstructure:"keeper" yaml:"keeper"`
	// DefaultRule configures the default rule for the keeper.
	DefaultRule RuleConfig `mapstructure:",squash" yaml:",inline"`
	// Rules configures the match rules for the keeper.
	Rules []MatchRuleConfig `mapstructure:"rules" yaml:"rules"`
}

var acceptances = map[string]Acceptance{
	"allow":   Allow,
	"deny":    Deny,
	"inherit": InheritAcceptance,
}

// ValidAcceptance returns true if the acceptance string is valid. The values
// "allow" and "deny" are always allowed. The value "inherit" is allowed when
// inheritAllowed is true.
func ValidAcceptance(a string, inheritAllowed bool) bool {
	if inheritAllowed && a == "inherit" {
		return true
	}
	return a == "allow" || a == "deny"
}

// Validate validates the policy configuration.
func Validate(ctx context.Context, c any) error {
	cfg, isPolicy := c.(*Config)
	if !isPolicy {
		return plugin.ErrConfig
	}

	return validate(ctx, cfg)
}

func validate(ctx context.Context, cfg *Config) error {
	errs := plugin.NewValidationError()

	if !keeper.Exists(ctx, cfg.Keeper) {
		errs.Append(fmt.Errorf("policy keeper %q does not exist", cfg.Keeper))
	}

	if !ValidAcceptance(cfg.DefaultRule.Acceptance, false) {
		errs.Append(fmt.Errorf("policy default rule acceptance %q must be allow or deny", cfg.DefaultRule.Acceptance))
	}

	for _, r := range cfg.Rules {
		if !ValidAcceptance(r.Acceptance, true) {
			errs.Append(fmt.Errorf("policy rule acceptance %q must be allow or deny or inherit", r.Acceptance))
		}

		if ValidAcceptance(r.Acceptance, false) && r.Lifetime > 0 {
			errs.Append(fmt.Errorf("policy rule with both lifteime and acceptance settings is not permitted"))
		}

		if !ValidAcceptance(r.Acceptance, false) && r.Lifetime == 0 {
			errs.Append(fmt.Errorf("policy rule with neither lifetime nor acceptance settings is not permitted"))
		}
	}

	return errs.Return()
}

// Builder constructs a new policy secret keeper.
func Builder(ctx context.Context, c any) (secrets.Keeper, error) {
	cfg, isPolicy := c.(*Config)
	if !isPolicy {
		return nil, plugin.ErrConfig
	}

	nextKpr, err := keeper.Build(ctx, cfg.Keeper)
	if err != nil {
		return nil, err
	}

	kpr := New(nextKpr)

	kpr.SetDefaultLifetime(cfg.DefaultRule.Lifetime)
	kpr.SetDefaultAcceptance(acceptances[cfg.DefaultRule.Acceptance])

	for _, r := range cfg.Rules {
		var rule *Rule
		switch {
		case r.Lifetime > 0:
			rule = NewLifetimeRule(r.Lifetime)
		case ValidAcceptance(r.Acceptance, true):
			rule = NewAcceptanceRule(acceptances[r.Acceptance])
		}

		match := &Match{m: r.MatchConfig}

		kpr.AddRule(&MatchRule{match, rule})
	}

	return kpr, nil
}

func init() {
	var (
		defaultPolicy bool
		appendPolicy  bool
		insertPolicy  int
		replacePolicy int
		removePolicy  int

		acceptance string
		lifetime   time.Duration

		locationMatch string
		nameMatch     string
		usernameMatch string
		typeMatch     string
		urlMatch      string
	)

	checkOptions := func(kc config.KeeperConfig) error {
		if !defaultPolicy && ValidAcceptance(acceptance, !defaultPolicy) && lifetime > 0 {
			return errors.New("cannot set both acceptance and lifetime policies")
		}

		if !defaultPolicy && !appendPolicy && insertPolicy == 0 && replacePolicy == 0 && removePolicy == 0 {
			return errors.New("you must specify the operation to perform")
		}

		matchers := 0
		if locationMatch != "" {
			matchers++
		}
		if nameMatch != "" {
			matchers++
		}
		if usernameMatch != "" {
			matchers++
		}
		if typeMatch != "" {
			matchers++
		}
		if urlMatch != "" {
			matchers++
		}

		if defaultPolicy && matchers > 0 {
			return errors.New("default policy has no match strings")
		}

		if (appendPolicy || insertPolicy >= 0 || replacePolicy >= 0) && matchers == 0 {
			return errors.New("you must specify at least one matcher")
		}

		rules := kc["rules"].([]map[string]any)

		if insertPolicy > len(rules) {
			return errors.New("insert index out of range")
		}

		if insertPolicy == len(rules) {
			insertPolicy = -1
			appendPolicy = true
		}

		if replacePolicy >= len(rules) {
			return errors.New("replace index out of range")
		}

		if removePolicy >= len(rules) {
			return errors.New("remove index out of range")
		}

		if removePolicy >= 0 && matchers > 0 {
			return errors.New("remove policy must not set match strings")
		}

		hasPolicy := ValidAcceptance(acceptance, !defaultPolicy) || lifetime > 0
		if (defaultPolicy || appendPolicy || insertPolicy >= 0 || replacePolicy >= 0) && !hasPolicy {
			return errors.New("must set acceptance or lifetime policy")
		}

		return nil
	}

	cmd := plugin.CmdConfig{
		Short: "Configure a policy enforcement secret keeper",
		FlagInit: func(flags *pflag.FlagSet) error {
			flags.BoolVar(&defaultPolicy, "default", false, "Set the default policy for the keeper")
			flags.BoolVar(&appendPolicy, "append", false, "Add a new rule to the policy")
			flags.IntVar(&insertPolicy, "insert", -1, "Insert a new rule at the specified index")
			flags.IntVar(&replacePolicy, "replace", -1, "Replace the rule at the specified index")
			flags.IntVar(&removePolicy, "remove", -1, "Remove the rule at the specified index")

			flags.StringVar(&acceptance, "acceptance", "", "Set the acceptance policy for the keeper")
			flags.DurationVar(&lifetime, "lifetime", 0, "Set the lifetime policy for the keeper")

			flags.StringVar(&locationMatch, "location", "", "Set the location policy for the keeper")
			flags.StringVar(&nameMatch, "name", "", "Set the name policy for the keeper")
			flags.StringVar(&usernameMatch, "username", "", "Set the username policy for the keeper")
			flags.StringVar(&typeMatch, "type", "", "Set the type policy for the keeper")
			flags.StringVar(&urlMatch, "url", "", "Set the url policy for the keeper")

			return nil
		},
		Run: func(keeperName string, fields map[string]any) (config.KeeperConfig, error) {
			c := config.Instance()
			kc := c.Keepers[keeperName]
			if kc == nil {
				kc = map[string]any{
					"type":  ConfigType,
					"rules": []map[string]any{},
				}
			}

			if err := checkOptions(kc); err != nil {
				return nil, err
			}

			if defaultPolicy {
				if acceptance != "" {
					kc["acceptance"] = acceptance
				}
				if lifetime > 0 {
					kc["lifetime"] = lifetime
				}
			}

			if appendPolicy || insertPolicy >= 0 || replacePolicy >= 0 {
				rule := map[string]any{
					"location":    locationMatch,
					"name":        nameMatch,
					"username":    usernameMatch,
					"secret_type": typeMatch,
					"url":         urlMatch,

					"acceptance": acceptance,
					"lifetime":   lifetime,
				}

				rules := kc["rules"].([]map[string]any)
				if rules == nil {
					rules = make([]map[string]any, 0, 1)
				}

				if appendPolicy {
					rules = append(rules, rule)
				}

				if insertPolicy >= 0 {
					rules = slices.Insert(rules, insertPolicy, rule)
				}

				if replacePolicy >= 0 {
					rules[replacePolicy] = rule
				}

				kc["rules"] = rules
			}

			if removePolicy >= 0 {
				rules := kc["rules"].([]map[string]any)
				kc["rules"] = slices.Delete(rules, removePolicy)
			}

			return kc, nil
		},
	}

	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, Validate, cmd)
}
