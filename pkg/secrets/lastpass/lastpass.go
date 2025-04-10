package lastpass

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ansd/lastpass-go"
	"github.com/zostay/go-std/set"

	"github.com/zostay/ghost/pkg/secrets"
)

// Client defines the interface required for a LastPass client.
// Normally, this is fulfilled by lastpass.Client, but is handled as an
// interface to make testing possible without an actual LastPass account.
type Client interface {
	// Accounts should list secrets.
	Accounts(ctx context.Context) ([]*lastpass.Account, error)

	// Update updates a single secret.
	Update(ctx context.Context, a *lastpass.Account) error

	// Add creates a new secret.
	Add(ctx context.Context, a *lastpass.Account) error

	// Delete deletes a secret.
	Delete(ctx context.Context, a *lastpass.Account) error
}

func _retry[T any](run func() (T, error)) (t T, err error) {
	backoff := 1 * time.Millisecond
	retriesLeft := 20
	for {
		if retriesLeft <= 0 {
			err = fmt.Errorf("too many retries: %w", err)
			return
		}

		<-time.After(backoff)

		t, err = run()
		if err != nil {
			// this is what the LastPass CLI does...
			if strings.Contains(err.Error(), "429") {
				backoff *= 8
			} else {
				backoff *= 2
			}
			retriesLeft--
			continue
		}

		return
	}
}

// LastPass is a secret Keeper that gets secrets from the LastPass
// password manager service.
type LastPass struct {
	lp Client
}

var _ secrets.Keeper = &LastPass{}

// NewWithCLient constructs a new LastPass Keeper with a custom LastPass
// client. This constructor is mostly intended for use during testing.
func NewWithCLient(
	lp Client,
) (*LastPass, error) {
	return &LastPass{lp}, nil
}

// New constructs and returns a new LastPass Keeper or returns an error
// if there was a problem during construction.
//
// The username and password arguments are used to authenticate with LastPass.
func New(ctx context.Context, username, password string) (*LastPass, error) {
	lp, err := lastpass.NewClient(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return &LastPass{lp}, nil
}

func (l *LastPass) listAccounts(ctx context.Context) ([]*lastpass.Account, error) {
	return _retry[[]*lastpass.Account](func() ([]*lastpass.Account, error) {
		return l.lp.Accounts(ctx)
	})
}

// ListLocations returns a list of LastPass folders.
func (l *LastPass) ListLocations(ctx context.Context) ([]string, error) {
	as, err := l.listAccounts(ctx)
	if err != nil {
		return nil, err
	}

	locations := set.New[string]()
	for _, a := range as {
		locations.Insert(a.Group)
	}

	return locations.Keys(), nil
}

// ListSecrets returns a list of secrets in each folder.
func (l *LastPass) ListSecrets(ctx context.Context, location string) ([]string, error) {
	as, err := l.listAccounts(ctx)
	if err != nil {
		return nil, err
	}

	secs := make([]string, 0, len(as))
	for _, a := range as {
		if a.Group == location {
			secs = append(secs, a.ID)
		}
	}

	return secs, nil
}

func (l *LastPass) getAccount(
	ctx context.Context,
	id string,
) (*lastpass.Account, error) {
	as, err := l.listAccounts(ctx)
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		if a.ID == id {
			return a, nil
		}
	}

	return nil, secrets.ErrNotFound
}

// GetSecret returns the secret from the Lastpass service.
func (l *LastPass) GetSecret(ctx context.Context, id string) (secrets.Secret, error) {
	a, err := l.getAccount(ctx, id)
	if err != nil {
		return nil, err
	}

	return newSecret(a), nil
}

// GetSecretsByName returns all the secrets from the LastPass service with the
// given name.
func (l *LastPass) GetSecretsByName(
	ctx context.Context,
	name string,
) ([]secrets.Secret, error) {
	as, err := l.listAccounts(ctx)
	if err != nil {
		return nil, err
	}

	var secs []secrets.Secret
	for _, a := range as {
		if a.Name == name {
			secs = append(secs, newSecret(a))
		}
	}

	return secs, nil
}

func (l *LastPass) addAccount(ctx context.Context, acc *lastpass.Account) error {
	_, err := _retry[struct{}](func() (struct{}, error) {
		return struct{}{}, l.lp.Add(ctx, acc)
	})
	return err
}

// SetSecret sets the secret value in the LastPass service.
func (l *LastPass) SetSecret(
	ctx context.Context,
	secret secrets.Secret,
) (secrets.Secret, error) {
	a, err := l.getAccount(ctx, secret.ID())
	if err != nil && !errors.Is(err, secrets.ErrNotFound) {
		return nil, err
	}

	if a == nil {
		newSec := fromSecret(secret)
		newSec.Account.ID = ""

		err = l.addAccount(ctx, newSec.Account)

		return newSec, err
	}

	newSec := fromSecret(secret)
	err = l.addAccount(ctx, newSec.Account)
	return newSec, err
}

func (l *LastPass) deleteAccount(ctx context.Context, acc *lastpass.Account) error {
	_, err := _retry[struct{}](func() (struct{}, error) {
		return struct{}{}, l.lp.Delete(ctx, acc)
	})
	return err
}

// DeleteSecret removes the account from LastPass.
func (l *LastPass) DeleteSecret(ctx context.Context, id string) error {
	a, err := l.getAccount(ctx, id)
	if err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			return nil
		}
		return err
	}

	return l.deleteAccount(ctx, a)
}

func (l *LastPass) updateAccount(ctx context.Context, acc *lastpass.Account) error {
	_, err := _retry[struct{}](func() (struct{}, error) {
		return struct{}{}, l.lp.Update(ctx, acc)
	})
	return err
}

// CopySecret copies the account from one group to another in LastPass.
func (l *LastPass) CopySecret(
	ctx context.Context,
	id, grp string,
) (secrets.Secret, error) {
	a, err := l.getAccount(ctx, id)
	if err != nil {
		return nil, err
	}

	newSec := newSecret(a)
	newSec.Account.ID = ""
	newSec.Group = grp
	return newSec, l.updateAccount(ctx, newSec.Account)
}

// MoveSecret copies the account to a new group and deletes the old one.
func (l *LastPass) MoveSecret(
	ctx context.Context,
	id, grp string,
) (secrets.Secret, error) {
	a, err := l.getAccount(ctx, id)
	if err != nil {
		return nil, err
	}

	newSec := newSecret(a)
	newSec.Group = grp
	return newSec, l.updateAccount(ctx, newSec.Account)
}
