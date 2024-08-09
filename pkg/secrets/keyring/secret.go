package keyring

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zostay/ghost/pkg/secrets"
)

// Secret is a secrets.Secret implementation that wraps a keyring entry. This
// implementation is unique in that the ID, name, and username are all treated
// the same.
type Secret struct {
	name  string
	value string

	decoded map[string]string
}

var _ secrets.Secret = &Secret{}

const (
	fldPassword     = "password"
	fldType         = "type"
	fldUrl          = "url"
	fldFieldsPrefix = "field:"
	fldMtime        = "mtime"
)

// FromKeyring creates a new Secret from the given keyring entry.
func FromKeyring(name string, value string) (*Secret, error) {
	dec, err := decodeValue(value)
	if err != nil {
		return nil, err
	}

	return &Secret{
		name:    name,
		value:   value,
		decoded: dec,
	}, nil
}

// FromSecret creates a new Secret from the given secret.
func FromSecret(secret secrets.Secret) (*Secret, error) {
	val, err := encodeSecret(secret)
	if err != nil {
		return nil, err
	}

	s := &Secret{
		name:  secret.Name(),
		value: val,
	}

	s.decoded, err = decodeValue(val)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func encodeSecret(secret secrets.Secret) (string, error) {
	if secret == nil {
		return "", errors.New("empty secrets not permitted")
	}

	if secret.Location() != "" {
		return "", errors.New("secrets with non-empty locations are not permitted")
	}

	if secret.ID() != secret.Name() {
		return "", errors.New("secrets must be stored with matching ID and Name")
	}

	if secret.ID() != secret.Username() {
		return "", errors.New("secrets must be stored with matching ID and Username")
	}

	var mt time.Time
	if secret.LastModified().IsZero() {
		mt = time.Now()
	} else {
		mt = secret.LastModified()
	}
	mtime := strconv.FormatInt(mt.Unix(), 10)
	raw := map[string]string{
		fldPassword: secret.Password(),
		fldType:     secret.Type(),
		fldMtime:    mtime,
		fldUrl:      secret.Url().String(),
	}
	for k, v := range secret.Fields() {
		raw[fldFieldsPrefix+k] = v
	}

	val := &strings.Builder{}
	enc := json.NewEncoder(val)
	enc.SetEscapeHTML(false)
	err := enc.Encode(raw)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

func decodeValue(value string) (map[string]string, error) {
	dec := map[string]string{}
	err := json.Unmarshal([]byte(value), &dec)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

// ID returns the name of the secret.
func (s *Secret) ID() string {
	return s.name
}

// Name returns the name of the secret.
func (s *Secret) Name() string {
	return s.name
}

// Username returns name of the secret.
func (s *Secret) Username() string {
	return s.name
}

// Password returns the password of the secret.
func (s *Secret) Password() string {
	return s.decoded[fldPassword]
}

// Type returns the type of the secret.
func (s *Secret) Type() string {
	return s.decoded[fldType]
}

// Fields returns the fields of the secret.
func (s *Secret) Fields() map[string]string {
	flds := make(map[string]string, len(s.decoded))
	for k, v := range s.decoded {
		if strings.HasPrefix(fldFieldsPrefix, k) {
			flds[strings.TrimPrefix(k, fldFieldsPrefix)] = v
		}
	}
	return flds
}

// GetField returns the value of the given field.
func (s *Secret) GetField(name string) string {
	return s.decoded[fldFieldsPrefix+name]
}

// LastModified returns the last modified time of the secret.
func (s *Secret) LastModified() time.Time {
	ue, _ := strconv.ParseInt(s.decoded[fldMtime], 10, 64)
	return time.Unix(ue, 0)
}

// Url returns the URL of the secret.
func (s *Secret) Url() *url.URL {
	u, _ := url.Parse(s.decoded[fldUrl])
	return u
}

// Location returns the location of the secret.
func (s *Secret) Location() string {
	return ""
}
