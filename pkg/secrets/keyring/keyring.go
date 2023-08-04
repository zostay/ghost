package keyring

import (
	"context"
	"errors"

	"github.com/zalando/go-keyring"

	"github.com/zostay/ghost/pkg/secrets"
)

type Keyring struct {
	serviceName string
}

func New(serviceName string) *Keyring {
	return &Keyring{
		serviceName: serviceName,
	}
}

func (k Keyring) ListLocations(_ context.Context) ([]string, error) {
	return []string{""}, nil
}

func (k Keyring) ListSecrets(context.Context, string) ([]string, error) {
	return nil, errors.New("secrets in keyring cannot be listed")
}

func (k Keyring) GetSecretsByName(_ context.Context, name string) ([]secrets.Secret, error) {
	password, err := keyring.Get(k.serviceName, name)
	if err != nil {
		return nil, err
	}

	sec, err := FromKeyring(name, password)
	if err != nil {
		return nil, err
	}

	return []secrets.Secret{sec}, nil
}

func (k Keyring) GetSecret(_ context.Context, id string) (secrets.Secret, error) {
	password, err := keyring.Get(k.serviceName, id)
	if err != nil {
		return nil, err
	}

	return FromKeyring(id, password)
}

func (k Keyring) SetSecret(_ context.Context, secret secrets.Secret) (secrets.Secret, error) {
	newSec, err := FromSecret(secret)
	if err != nil {
		return nil, err
	}

	err = keyring.Set(k.serviceName, newSec.name, newSec.value)
	if err != nil {
		return nil, err
	}

	return newSec, nil
}

func (k Keyring) CopySecret(context.Context, string, string) (secrets.Secret, error) {
	return nil, errors.New("secrets in keyring cannot be copied")
}

func (k Keyring) MoveSecret(context.Context, string, string) (secrets.Secret, error) {
	return nil, errors.New("secrets in keyring cannot be moved")
}

func (k Keyring) DeleteSecret(_ context.Context, id string) error {
	return keyring.Delete(k.serviceName, id)
}

var _ secrets.Keeper = &Keyring{}
