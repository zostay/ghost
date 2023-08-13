package set

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/config"
)

var (
	PolicyCmd = &cobra.Command{
		Use:     "policy <keeper-name> [flags]",
		Short:   "Configure a policy enforcement secret keeper",
		Args:    cobra.ExactArgs(1),
		PreRunE: PreRunSetPolicyKeeperConfig,
		Run:     RunSetKeeperConfig,
	}

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

func init() {
	PolicyCmd.Flags().BoolVar(&defaultPolicy, "default", false, "Set the default policy for the keeper")
	PolicyCmd.Flags().BoolVar(&appendPolicy, "append", false, "Add a new rule to the policy")
	PolicyCmd.Flags().IntVar(&insertPolicy, "insert", -1, "Insert a new rule at the specified index")
	PolicyCmd.Flags().IntVar(&replacePolicy, "replace", -1, "Replace the rule at the specified index")
	PolicyCmd.Flags().IntVar(&removePolicy, "remove", -1, "Remove the rule at the specified index")

	PolicyCmd.Flags().StringVar(&acceptance, "acceptance", "", "Set the acceptance policy for the keeper")
	PolicyCmd.Flags().DurationVar(&lifetime, "lifetime", 0, "Set the lifetime policy for the keeper")

	PolicyCmd.Flags().StringVar(&locationMatch, "location", "", "Set the location policy for the keeper")
	PolicyCmd.Flags().StringVar(&nameMatch, "name", "", "Set the name policy for the keeper")
	PolicyCmd.Flags().StringVar(&usernameMatch, "username", "", "Set the username policy for the keeper")
	PolicyCmd.Flags().StringVar(&typeMatch, "type", "", "Set the type policy for the keeper")
	PolicyCmd.Flags().StringVar(&urlMatch, "url", "", "Set the url policy for the keeper")
}

func checkOptions() error {
	if !defaultPolicy && config.ValidAcceptance(acceptance, !defaultPolicy) && lifetime > 0 {
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

	if insertPolicy > len(Replacement.Policy.Rules) {
		return errors.New("insert index out of range")
	}

	if insertPolicy == len(Replacement.Policy.Rules) {
		insertPolicy = -1
		appendPolicy = true
	}

	if replacePolicy >= len(Replacement.Policy.Rules) {
		return errors.New("replace index out of range")
	}

	if removePolicy >= len(Replacement.Policy.Rules) {
		return errors.New("remove index out of range")
	}

	if removePolicy >= 0 && matchers > 0 {
		return errors.New("remove policy must not set match strings")
	}

	hasPolicy := config.ValidAcceptance(acceptance, !defaultPolicy) || lifetime > 0
	if (defaultPolicy || appendPolicy || insertPolicy >= 0 || replacePolicy >= 0) && !hasPolicy {
		return errors.New("must set acceptance or lifetime policy")
	}

	return nil
}

func PreRunSetPolicyKeeperConfig(cmd *cobra.Command, args []string) error {
	if err := checkOptions(); err != nil {
		return err
	}

	if defaultPolicy {
		if acceptance != "" {
			Replacement.Policy.DefaultRule.Acceptance = acceptance
		}
		if lifetime > 0 {
			Replacement.Policy.DefaultRule.Lifetime = lifetime
		}
	}

	if appendPolicy || insertPolicy >= 0 || replacePolicy >= 0 {
		rule := config.PolicyMatchRuleConfig{
			PolicyMatchConfig: config.PolicyMatchConfig{
				LocationMatch: locationMatch,
				NameMatch:     nameMatch,
				UsernameMatch: usernameMatch,
				TypeMatch:     typeMatch,
				UrlMatch:      urlMatch,
			},
			PolicyRuleConfig: config.PolicyRuleConfig{
				Lifetime:   lifetime,
				Acceptance: acceptance,
			},
		}

		if appendPolicy {
			Replacement.Policy.Rules = append(Replacement.Policy.Rules, rule)
		}

		if insertPolicy >= 0 {
			Replacement.Policy.Rules = slices.Insert(Replacement.Policy.Rules, insertPolicy, rule)
		}

		if replacePolicy >= 0 {
			Replacement.Policy.Rules[replacePolicy] = rule
		}
	}

	if removePolicy >= 0 {
		Replacement.Policy.Rules = slices.Delete(Replacement.Policy.Rules, removePolicy)
	}

	return nil
}
