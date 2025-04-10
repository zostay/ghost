package human

import (
	"context"
	"errors"
	"net/url"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/zostay/ghost/pkg/keeper"
	"github.com/zostay/ghost/pkg/secrets"
)

// Question defines the questions to ask for a secret as well as the preset
// values to use to fill in the rest.
type Question struct {
	preset map[string]string
	flds   []string
}

// Human is a secret keeper that asks the user for the secrets.
type Human struct {
	secrets map[string]Question
}

var _ secrets.Keeper = &Human{}

// New creates a new human secrets keeper.
func New() *Human {
	return &Human{
		secrets: make(map[string]Question, 1),
	}
}

// AddQuestion adds a question to the human secrets keeper.
func (h *Human) AddQuestion(
	id string,
	askFor []string,
	preset map[string]string,
) {
	h.secrets[id] = Question{
		preset: preset,
		flds:   askFor,
	}
}

// ListLocations returns the list of locations from the human secrets keeper.
// This always just returns "". As of this writing, you should only use an empty
// location with the human secret keeper.
func (h *Human) ListLocations(_ context.Context) ([]string, error) {
	return []string{""}, nil
}

// ListSecrets returns an empty list.
func (h *Human) ListSecrets(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func setField(
	sec *secrets.Single,
	fld string,
	val string,
) error {
	switch fld {
	case "username":
		sec.SetUsername(val)
	case "password":
		sec.SetPassword(val)
	case "url":
		u, err := url.Parse(val)
		if err != nil {
			return err
		}
		sec.SetUrl(u)
	case "type":
		sec.SetType(val)
	default:
		sec.SetField(fld, val)
	}
	return nil
}

// GetSecret retrieves the secret with the given ID by asking the user for the
// secret information as defined by the identified question configuration.
func (h *Human) GetSecret(_ context.Context, id string) (secrets.Secret, error) {
	title := cases.Title(language.English)
	if def, defExists := h.secrets[id]; defExists {
		sec := secrets.NewSecret(id, "", "",
			secrets.WithID(id),
			secrets.WithLastModified(time.Now()))
		for _, fld := range def.flds {
			v, err := keeper.GetPassword(
				"Enter "+fld,
				"Asking for "+fld+" for "+id,
				title.String(fld),
				"OK",
			)

			if err != nil {
				return nil, err
			}

			if err := setField(sec, fld, v); err != nil {
				return nil, err
			}
		}

		for k, v := range def.preset {
			if err := setField(sec, k, v); err != nil {
				return nil, err
			}
		}

		return sec, nil
	}

	return nil, secrets.ErrNotFound
}

// GetSecretsByName retrieves the secret with the given name by asking the user
// for the secret information as defined by the identified question
// configuration. As ID and name are treated the same by the human secret
// keeper, this is essentially identical to GetSecret.
func (h *Human) GetSecretsByName(ctx context.Context, name string) ([]secrets.Secret, error) {
	sec, err := h.GetSecret(ctx, name)
	if err != nil {
		return nil, err
	}
	return []secrets.Secret{sec}, nil
}

// SetSecret fails with an error.
func (h *Human) SetSecret(_ context.Context, _ secrets.Secret) (secrets.Secret, error) {
	return nil, errors.New("read only")
}

// CopySecret fails with an error.
func (h *Human) CopySecret(_ context.Context, _, _ string) (secrets.Secret, error) {
	return nil, errors.New("read only")
}

// MoveSecret fails with an error.
func (h *Human) MoveSecret(_ context.Context, _, _ string) (secrets.Secret, error) {
	return nil, errors.New("read only")
}

// DeleteSecret fails with an error.
func (h *Human) DeleteSecret(_ context.Context, _ string) error {
	return errors.New("read only")
}
