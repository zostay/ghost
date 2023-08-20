package policy

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/plugin"
	"github.com/zostay/ghost/pkg/secrets"
)

const ConfigType = "policy"

type RuleConfig struct {
	Lifetime   time.Duration `mapstructure:"lifetime" yaml:"lifetime"`
	Acceptance string        `mapstructure:"acceptance" yaml:"acceptance"`
}

type MatchConfig struct {
	LocationMatch string `mapstructure:"location" yaml:"location"`
	NameMatch     string `mapstructure:"name" yaml:"name"`
	UsernameMatch string `mapstructure:"username" yaml:"username"`
	TypeMatch     string `mapstructure:"secret_type" yaml:"secret_type"`
	UrlMatch      string `mapstructure:"url" yaml:"url"`
}

type MatchRuleConfig struct {
	MatchConfig `mapstructure:",squash" yaml:",inline"`
	RuleConfig  `mapstructure:",squash" yaml:",inline"`
}

type Config struct {
	Keeper      string            `mapstructure:"keeper" yaml:"keeper"`
	DefaultRule RuleConfig        `mapstructure:",squash" yaml:",inline"`
	Rules       []MatchRuleConfig `mapstructure:"rules" yaml:"rules"`
}

var acceptances = map[string]Acceptance{
	"allow":   Allow,
	"deny":    Deny,
	"inherit": InheritAcceptance,
}

func ValidAcceptance(a string, inheritAllowed bool) bool {
	if inheritAllowed && a == "inherit" {
		return true
	}
	return a == "allow" || a == "deny"
}

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

		match := &Match{
			LocationMatch: r.LocationMatch,
			NameMatch:     r.NameMatch,
			UsernameMatch: r.UsernameMatch,
			TypeMatch:     r.TypeMatch,
			UrlMatch:      r.UrlMatch,
		}

		kpr.AddRule(&MatchRule{match, rule})
	}

	return kpr, nil
}

func init() {
	plugin.Register(ConfigType, reflect.TypeOf(Config{}), Builder, Validate)
}
