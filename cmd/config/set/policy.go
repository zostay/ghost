package set

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/policy"
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

func checkOptions(kc config.KeeperConfig) error {
	if !defaultPolicy && policy.ValidAcceptance(acceptance, !defaultPolicy) && lifetime > 0 {
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

	hasPolicy := policy.ValidAcceptance(acceptance, !defaultPolicy) || lifetime > 0
	if (defaultPolicy || appendPolicy || insertPolicy >= 0 || replacePolicy >= 0) && !hasPolicy {
		return errors.New("must set acceptance or lifetime policy")
	}

	return nil
}

func PreRunSetPolicyKeeperConfig(cmd *cobra.Command, args []string) error {
	keeperName := args[0]
	c := config.Instance()
	kc := c.Keepers[keeperName]
	if kc == nil {
		kc = map[string]any{
			"type":  policy.ConfigType,
			"rules": []map[string]any{},
		}
		Replacement = kc
	}

	if err := checkOptions(kc); err != nil {
		return err
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

	return nil
}
