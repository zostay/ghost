package http

import (
	"net/url"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zostay/ghost/pkg/secrets"
)

type SecretWrapper struct {
	*Secret
}

var _ secrets.Secret = &SecretWrapper{}

func NewSecretWrapper(s *Secret) *SecretWrapper {
	return &SecretWrapper{
		Secret: s,
	}
}

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

func (s *SecretWrapper) ID() string {
	return s.GetId()
}

func (s *SecretWrapper) Name() string {
	return s.GetName()
}

func (s *SecretWrapper) SetName(name string) {
	s.init()
	s.Secret.Name = name
}

func (s *SecretWrapper) Username() string {
	return s.GetUsername()
}

func (s *SecretWrapper) SetUsername(username string) {
	s.init()
	s.Secret.Username = username
}

func (s *SecretWrapper) Password() string {
	return s.GetPassword()
}

func (s *SecretWrapper) SetPassword(password string) {
	s.init()
	s.Secret.Password = password
}

func (s *SecretWrapper) Type() string {
	return s.GetType()
}

func (s *SecretWrapper) SetType(typ string) {
	s.init()
	s.Secret.Type = typ
}

func (s *SecretWrapper) Fields() map[string]string {
	return s.GetFields()
}

func (s *SecretWrapper) SetFields(flds map[string]string) {
	s.init()
	s.Secret.Fields = flds
}

func (s *SecretWrapper) GetField(name string) string {
	return s.Fields()[name]
}

func (s *SecretWrapper) SetField(name, value string) {
	s.init()
	if s.Secret.Fields == nil {
		s.Secret.Fields = map[string]string{name: value}
		return
	}
	s.Secret.Fields[name] = value
}

func (s *SecretWrapper) DeleteField(name string) {
	s.init()
	if s.Secret.Fields == nil {
		return
	}
	delete(s.Secret.Fields, name)
}

func (s *SecretWrapper) LastModified() time.Time {
	return s.GetLastModified().AsTime()
}

func (s *SecretWrapper) Url() *url.URL {
	u, _ := url.Parse(s.GetUrl())
	return u
}

func (s *SecretWrapper) SetUrl(url *url.URL) {
	s.init()
	s.Secret.Url = url.String()
}

func (s *SecretWrapper) Location() string {
	return s.GetLocation()
}
