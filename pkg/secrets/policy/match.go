package policy

import (
	"regexp"
	"strings"

	"github.com/gobwas/glob"

	"github.com/zostay/ghost/pkg/secrets"
)

type matchStatus int

const (
	matchMiss matchStatus = iota
	matchYes
	matchNo
)

type matchFunc func(string) bool

type Match struct {
	LocationMatch string
	NameMatch     string
	UsernameMatch string
	TypeMatch     string
	UrlMatch      string
}

var (
	matcherCache = map[string]matchFunc{}
)

func matchToStatus(m bool) matchStatus {
	if m {
		return matchYes
	}
	return matchNo
}

func matchString(match, against string) matchStatus {
	if match == "" {
		return matchMiss
	}

	if matcher, hasMatcher := matcherCache[match]; hasMatcher {
		return matchToStatus(matcher(against))
	}

	if strings.HasPrefix(match, "/") && strings.HasSuffix(match, "/") {
		re := regexp.MustCompile(match[1 : len(match)-1])
		matcherCache[match] = re.MatchString
		return matchToStatus(re.MatchString(against))
	}

	gl := glob.MustCompile(match)
	matcherCache[match] = gl.Match
	return matchToStatus(gl.Match(against))
}

func (m Match) matchLocation(loc string) matchStatus {
	return matchString(m.LocationMatch, loc)
}

func (m Match) matchName(name string) matchStatus {
	return matchString(m.NameMatch, name)
}

func (m Match) matchUsername(username string) matchStatus {
	return matchString(m.UsernameMatch, username)
}

func (m Match) matchType(typ string) matchStatus {
	return matchString(m.TypeMatch, typ)
}

func (m Match) matchUrl(url string) matchStatus {
	return matchString(m.UrlMatch, url)
}

func (m Match) matchSecret(sec secrets.Secret) matchStatus {
	fs := []struct {
		mf func(string) matchStatus
		s  string
	}{
		{m.matchName, sec.Name()},
		{m.matchLocation, sec.Location()},
		{m.matchUsername, sec.Username()},
		{m.matchType, sec.Type()},
		{m.matchUrl, sec.Url().String()},
	}

	yesses := 0
	for _, mp := range fs {
		ms := mp.mf(mp.s)
		if ms == matchYes {
			yesses++
		}

		if ms == matchNo {
			return matchNo
		}
	}

	if yesses > 0 {
		return matchYes
	}

	return matchMiss
}
