package policy

import "time"

type Acceptance int

const (
	Deny Acceptance = iota
	Allow
	InheritAcceptance
)

type Rule struct {
	lifetime   time.Duration
	acceptance Acceptance
}

func NewLifetimeRule(l time.Duration) *Rule {
	return &Rule{
		lifetime:   l,
		acceptance: InheritAcceptance,
	}
}

func NewAcceptanceRule(a Acceptance) *Rule {
	return &Rule{
		lifetime:   -1,
		acceptance: a,
	}
}
