package policy

import (
	"time"

	"github.com/zostay/ghost/pkg/secrets"
)

type MatchRule struct {
	*Match
	*Rule
}

func (mr MatchRule) matchLocationAndAcceptable(loc string) matchStatus {
	ms := mr.matchLocation(loc)
	if ms == matchMiss || ms == matchNo {
		return matchMiss
	}

	switch mr.acceptance {
	case Allow:
		return matchYes
	case InheritAcceptance:
		return matchMiss
	case Deny:
		return matchNo
	}
	return matchNo
}

func (mr MatchRule) matchSecretAndAccessible(defRule *Rule, sec secrets.Secret) matchStatus {
	ms := mr.matchSecret(sec)
	if ms == matchMiss || ms == matchNo {
		return matchMiss
	}

	if mr.acceptance == Deny {
		return matchNo
	}

	allowed := mr.acceptance == Allow || (mr.acceptance == InheritAcceptance && defRule.acceptance == Allow)

	if allowed {
		return matchYes
	}

	return matchMiss
}

func (mr MatchRule) matchSecretAndLifetime(sec secrets.Secret) (matchStatus, time.Duration) {
	ms := mr.matchSecret(sec)
	if ms == matchYes {
		return matchYes, mr.lifetime
	}

	return matchMiss, 0
}
