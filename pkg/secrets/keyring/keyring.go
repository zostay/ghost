package keyring

import (
	"context"
	"errors"

	"github.com/zalando/go-keyring"

	"github.com/zostay/ghost/pkg/secrets"
)

// Keyring is a Keeper with access to the system keyring.
type Keyring struct {
	serviceName string
}

var _ secrets.Keeper = &Keyring{}

// New creates a new Keyring Keeper and returns it.
func New(serviceName string) *Keyring {
	return &Keyring{
		serviceName: serviceName,
	}
}

// ListLocations returns a list of locations that this Keeper can access.
func (k Keyring) ListLocations(_ context.Context) ([]string, error) {
	return []string{""}, nil
}

// ListSecrets returns a list of secrets in the given location.
func (k Keyring) ListSecrets(context.Context, string) ([]string, error) {
	return nil, errors.New("secrets in keyring cannot be listed")
}

// GetSecretsByName returns a list of secrets in the given location with the
// given name.
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

// GetSecret returns the secret with the given ID.
func (k Keyring) GetSecret(_ context.Context, id string) (secrets.Secret, error) {
	password, err := keyring.Get(k.serviceName, id)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, secrets.ErrNotFound
		}

		return nil, err
	}

	return FromKeyring(id, password)
}

// SetSecret sets the secret with the given ID.
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

// CopySecret copies the secret with the given ID to the new name.
func (k Keyring) CopySecret(context.Context, string, string) (secrets.Secret, error) {
	return nil, errors.New("secrets in keyring cannot be copied")
}

// MoveSecret moves the secret with the given ID to the new name.
func (k Keyring) MoveSecret(context.Context, string, string) (secrets.Secret, error) {
	return nil, errors.New("secrets in keyring cannot be moved")
}

// DeleteSecret deletes the secret with the given ID.
func (k Keyring) DeleteSecret(_ context.Context, id string) error {
	return keyring.Delete(k.serviceName, id)
}
