package lastpass

import (
	"net/url"
	"strconv"
	"time"

	"github.com/ansd/lastpass-go"

	"github.com/zostay/ghost/pkg/secrets"
)

// Secret is a secrets.Secret implementation that wraps a LastPass account.
type Secret struct {
	*lastpass.Account

	parsed bool
	typ    string
	notes  map[string]string
}

func newSecret(account *lastpass.Account) *Secret {
	return &Secret{
		Account: account,
	}
}

func fromSecret(secret secrets.Secret) *Secret {
	newSec := newSecret(
		&lastpass.Account{
			ID:       secret.ID(),
			Name:     secret.Name(),
			Username: secret.Username(),
			Password: secret.Password(),
			URL:      secret.Url().String(),
			Group:    secret.Location(),
			Notes:    writeNotes(secret.Type(), secret.Fields()),
		},
	)

	if s, isSecret := secret.(*Secret); isSecret && s.parsed {
		newSec.Notes = writeNotes(s.typ, s.notes)
	}

	return newSec
}

// ID returns the LastPass account ID.
func (s *Secret) ID() string {
	return s.Account.ID
}

// Name returns the LastPass account name.
func (s *Secret) Name() string {
	return s.Account.Name
}

// SetName sets the LastPass account name.
func (s *Secret) SetName(name string) {
	s.Account.Name = name
}

// Username returns the LastPass account username.
func (s *Secret) Username() string {
	return s.Account.Username
}

// SetUsername sets the LastPass account username.
func (s *Secret) SetUsername(username string) {
	s.Account.Username = username
}

// Password returns the LastPass account password.
func (s *Secret) Password() string {
	return s.Account.Password
}

// SetPassword sets the LastPass account password.
func (s *Secret) SetPassword(secret string) {
	s.Account.Password = secret
}

// Url returns the LastPass account URL.
func (s *Secret) Url() *url.URL {
	url, _ := url.Parse(s.Account.URL)
	return url
}

// SetUrl sets the LastPass account URL.
func (s *Secret) SetUrl(url *url.URL) {
	s.Account.URL = url.String()
}

// Location returns the LastPass account Group.
func (s *Secret) Location() string {
	return s.Account.Group
}

func (s *Secret) parseNotes() {
	if s.parsed {
		return
	}

	flds := parseNotes(s.Account.Notes)
	s.typ = flds["NoteType"]
	delete(flds, "NoteType")
	s.notes = flds
	s.parsed = true
}

// Type returns the LastPass account NoteType.
func (s *Secret) Type() string {
	s.parseNotes()
	return s.typ
}

// SetType sets the LastPass account NoteType.
func (s *Secret) SetType(typ string) {
	s.parseNotes()
	s.typ = typ
}

// Fields returns the LastPass account fields to be stored in the note field.
func (s *Secret) Fields() map[string]string {
	s.parseNotes()
	return s.notes
}

// GetField gets a LastPass account field from the note field.
func (s *Secret) GetField(name string) string {
	s.parseNotes()
	return s.notes[name]
}

// SetFields sets all LastPass account field in the note field at once.
func (s *Secret) SetFields(fields map[string]string) {
	s.parseNotes()
	s.notes = fields
}

// SetField sets a LastPass account field in the note field.
func (s *Secret) SetField(name, value string) {
	s.parseNotes()
	s.notes[name] = value
}

// LastModified returns the LastPass account LastModifiedGMT.
func (s *Secret) LastModified() time.Time {
	lmSeconds, _ := strconv.ParseInt(s.Account.LastModifiedGMT, 10, 64)
	return time.Unix(lmSeconds, 0)
}
