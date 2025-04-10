package policy

import "time"

type Acceptance int

const (
	Deny              Acceptance = iota // secret is not accessible
	Allow                               // secret is accessible
	InheritAcceptance                   // secret inherits the policy default
)

// Rule is a policy rule that applies to secrets.
type Rule struct {
	lifetime   time.Duration
	acceptance Acceptance
}

// NewLifetimeRule creates a new rule with the given lifetime and inherit acceptance.
func NewLifetimeRule(l time.Duration) *Rule {
	return &Rule{
		lifetime:   l,
		acceptance: InheritAcceptance,
	}
}

// NewAcceptanceRule creates a new rule with the given acceptance and no
// lifetime.
func NewAcceptanceRule(a Acceptance) *Rule {
	return &Rule{
		lifetime:   -1,
		acceptance: a,
	}
}
