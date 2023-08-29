package keepass

import (
	"fmt"
	"net/url"
	"time"

	keepass "github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
	"github.com/zostay/go-std/set"

	"github.com/zostay/ghost/pkg/secrets"
)

const (
	keyTitle    = "Title"
	keyUsername = "Username"
	keySecret   = "Password"
	keyType     = "Type"
	keyURL      = "URL"
)

var stdKeys = set.New(keyTitle, keyUsername, keySecret, keyType, keyURL)

// Secret is a secrets.Secret implementation that wraps a keepass.Entry.
type Secret struct {
	db  *keepass.Database
	e   *keepass.Entry
	dir string

	newFields   map[string]string
	delFields   set.Set[string]
	newUrl      *url.URL
	newLocation *string
}

func newSecret(
	db *keepass.Database,
	e *keepass.Entry,
	dir string,
) *Secret {
	return &Secret{
		db:        db,
		e:         e,
		dir:       dir,
		newFields: map[string]string{},
		delFields: set.New[string](),
	}
}

func fromSecret(
	db *keepass.Database,
	secret secrets.Secret,
	keepID bool,
) *Secret {
	if eSec, isESec := secret.(*Secret); isESec {
		cp := eSec.e.Clone()
		retSec := &Secret{
			db:        db,
			e:         &cp,
			dir:       eSec.dir,
			newFields: map[string]string{},
			delFields: set.New[string](),
		}

		if keepID {
			retSec.e.UUID = eSec.e.UUID
		}

		retSec.applyChanges(secret)
		return retSec
	}

	var uuid keepass.UUID
	if keepID {
		uuid, _ = makeUUID(secret.ID())
	} else {
		uuid = keepass.NewUUID()
	}

	eSec := &Secret{
		db: db,
		e: &keepass.Entry{
			UUID:   uuid,
			Values: make([]keepass.ValueData, 0, len(secret.Fields())+stdKeys.Len()),
		},
		dir: secret.Location(),
	}

	eSec.applyChanges(secret)
	return eSec
}

// setEntryValue replaces a value in an entry or adds the value to the entry.
func (s *Secret) setEntryValue(key, value string, protected bool) {
	// update existing
	for k, v := range s.e.Values {
		if v.Key == key {
			s.e.Values[k].Value.Content = value
			return
		}
	}

	// create new
	newValue := keepass.ValueData{
		Key: key,
		Value: keepass.V{
			Content:   value,
			Protected: w.NewBoolWrapper(protected),
		},
	}
	s.e.Values = append(s.e.Values, newValue)
}

func (s *Secret) applyChanges(secret secrets.Secret) {
	for k, v := range secret.Fields() {
		s.setEntryValue(k, v, false)
	}

	s.setEntryValue(keyTitle, secret.Name(), false)
	s.setEntryValue(keyUsername, secret.Username(), false)
	s.setEntryValue(keySecret, secret.Password(), true)
	s.setEntryValue(keyType, secret.Type(), false)
	if secret.Url() != nil {
		s.setEntryValue(keyURL, secret.Url().String(), false)
	}
}

func makeID(id keepass.UUID) string {
	t, _ := id.MarshalText()
	return string(t)
}

func makeUUID(id string) (keepass.UUID, error) {
	var uuid keepass.UUID
	err := uuid.UnmarshalText([]byte(id))
	return uuid, err
}

func (s *Secret) set(key, value string) {
	if s.newFields == nil {
		s.newFields = map[string]string{}
	}
	s.newFields[key] = value
	s.delFields.Delete(key)
}

// ID returns the UUID of the Keepass entry.
func (s *Secret) ID() string {
	return makeID(s.e.UUID)
}

// Name returns the Title of the Keepass entry.
func (s *Secret) Name() string {
	if title, hasNewTitle := s.newFields[keyTitle]; hasNewTitle {
		return title
	}
	return s.e.GetTitle()
}

// SetName sets the Title of the Keepass entry.
func (s *Secret) SetName(name string) {
	s.set(keyTitle, name)
}

// Username returns the Username of the Keepass entry.
func (s *Secret) Username() string {
	if username, hasNewUsername := s.newFields[keyUsername]; hasNewUsername {
		return username
	}
	return s.e.GetContent(keyUsername)
}

// SetUsername sets the Username of the Keepass entry.
func (s *Secret) SetUsername(username string) {
	s.set(keyUsername, username)
}

func (s *Secret) whileUnlocked(run func()) {
	err := s.db.UnlockProtectedEntries()
	if err != nil {
		panic(fmt.Errorf("failed to unlock protected entries: %w", err))
	}
	defer func() {
		err := s.db.LockProtectedEntries()
		if err != nil {
			panic(fmt.Errorf("failed to lock protected entries: %w", err))
		}
	}()
	run()
}

// Password returns the Password of the Keepass entry.
func (s *Secret) Password() string {
	if secret, hasNewSecret := s.newFields[keySecret]; hasNewSecret {
		return secret
	}

	var secret string
	s.whileUnlocked(func() {
		secret = s.e.GetPassword()
	})
	return secret
}

// SetPassword sets the Password of the Keepass entry.
func (s *Secret) SetPassword(secret string) {
	s.set(keySecret, secret)
}

// Type returns the Type of the Keepass entry.
func (s *Secret) Type() string {
	if typ, hasNewType := s.newFields[keyType]; hasNewType {
		return typ
	}
	return s.e.GetContent(keyType)
}

// SetType sets the Type of the Keepass entry.
func (s *Secret) SetType(typ string) {
	s.set(keyType, typ)
}

// Fields returns the fields of the Keepass entry.
func (s *Secret) Fields() map[string]string {
	flds := make(map[string]string, len(s.e.Values))
	for _, val := range s.e.Values {
		if stdKeys.Contains(val.Key) {
			continue
		}
		if newValue, hasNewValue := s.newFields[val.Key]; hasNewValue {
			flds[val.Key] = newValue
			continue
		}
		flds[val.Key] = val.Value.Content
	}
	return flds
}

// GetField	returns the value of the field with the given key.
func (s *Secret) GetField(key string) string {
	if stdKeys.Contains(key) {
		return ""
	}

	if newValue, hasNewValue := s.newFields[key]; hasNewValue {
		return newValue
	}
	return s.e.GetContent(key)
}

// SetField sets the value of the field with the given key.
func (s *Secret) SetField(key, value string) {
	if key == keySecret {
		s.SetPassword(value)
	}
	s.set(key, value)
}

// DeleteField removes the field with the given key.
func (s *Secret) DeleteField(key string) {
	s.delFields.Insert(key)
}

// LastModified returns the last modification time of the Keepass entry.
func (s *Secret) LastModified() time.Time {
	return s.e.Times.LastModificationTime.Time
}

// Url returns the URL of the Keepass entry.
func (s *Secret) Url() *url.URL {
	if s.newUrl != nil {
		return s.newUrl
	}
	urlStr := s.e.GetContent(keyURL)
	u, _ := url.Parse(urlStr)
	return u
}

// SetUrl sets the URL of the Keepass entry.
func (s *Secret) SetUrl(u *url.URL) {
	s.newUrl = u
}

// Location returns the full path of the location.
func (s *Secret) Location() string {
	if s.newLocation != nil {
		return *s.newLocation
	}
	return s.dir
}

// SetLocation sets the location of the full path of the secret.
func (s *Secret) SetLocation(loc string) {
	s.newLocation = &loc
}
