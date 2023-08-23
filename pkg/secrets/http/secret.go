package http

import (
	"net/url"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zostay/ghost/pkg/secrets"
)

// SecretWrapper is a helper that maps the secrets.Secret interface to the
// protobuf Secret message.
type SecretWrapper struct {
	*Secret
}

var _ secrets.Secret = &SecretWrapper{}

// NewSecretWrapper creates a new secret wrapper for the given protobuf Secret
// message.
func NewSecretWrapper(s *Secret) *SecretWrapper {
	return &SecretWrapper{
		Secret: s,
	}
}

// FromSecret creates a new protobuf Secret message from the given secret.
func FromSecret(s secrets.Secret) *Secret {
	return &Secret{
		Id:           s.ID(),
		Name:         s.Name(),
		Username:     s.Username(),
		Password:     s.Password(),
		Type:         s.Type(),
		Fields:       s.Fields(),
		Url:          s.Url().String(),
		Location:     s.Location(),
		LastModified: timestamppb.New(s.LastModified()),
	}
}

func (s *SecretWrapper) init() {
	if s.Secret == nil {
		s.Secret = &Secret{}
	}
}

// ID returns the ID of the secret.
func (s *SecretWrapper) ID() string {
	return s.GetId()
}

// Name returns the name of the secret.
func (s *SecretWrapper) Name() string {
	return s.GetName()
}

// SetName sets the name of the secret.
func (s *SecretWrapper) SetName(name string) {
	s.init()
	s.Secret.Name = name
}

// Username returns the username of the secret.
func (s *SecretWrapper) Username() string {
	return s.GetUsername()
}

// SetUsername sets the username of the secret.
func (s *SecretWrapper) SetUsername(username string) {
	s.init()
	s.Secret.Username = username
}

// Password returns the password of the secret.
func (s *SecretWrapper) Password() string {
	return s.GetPassword()
}

// SetPassword sets the password of the secret.
func (s *SecretWrapper) SetPassword(password string) {
	s.init()
	s.Secret.Password = password
}

// Type returns the type of the secret.
func (s *SecretWrapper) Type() string {
	return s.GetType()
}

// SetType sets the type of the secret.
func (s *SecretWrapper) SetType(typ string) {
	s.init()
	s.Secret.Type = typ
}

// Fields returns the fields of the secret.
func (s *SecretWrapper) Fields() map[string]string {
	return s.GetFields()
}

// SetFields sets the fields of the secret.
func (s *SecretWrapper) SetFields(flds map[string]string) {
	s.init()
	s.Secret.Fields = flds
}

// GetField returns the value of the field with the given name.
func (s *SecretWrapper) GetField(name string) string {
	return s.Fields()[name]
}

// SetField sets the value of the field with the given name.
func (s *SecretWrapper) SetField(name, value string) {
	s.init()
	if s.Secret.Fields == nil {
		s.Secret.Fields = map[string]string{name: value}
		return
	}
	s.Secret.Fields[name] = value
}

// DeleteField deletes the field with the given name.
func (s *SecretWrapper) DeleteField(name string) {
	s.init()
	if s.Secret.Fields == nil {
		return
	}
	delete(s.Secret.Fields, name)
}

// LastModified returns the last modified date of the secret.
func (s *SecretWrapper) LastModified() time.Time {
	return s.GetLastModified().AsTime()
}

// Url returns the URL of the secret.
func (s *SecretWrapper) Url() *url.URL {
	u, _ := url.Parse(s.GetUrl())
	return u
}

// SetUrl sets the URL of the secret.
func (s *SecretWrapper) SetUrl(url *url.URL) {
	s.init()
	s.Secret.Url = url.String()
}

// Location returns the location of the secret.
func (s *SecretWrapper) Location() string {
	return s.GetLocation()
}
