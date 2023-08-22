package human

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/gopasspw/pinentry"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/zostay/ghost/pkg/secrets"
)

type Question struct {
	preset map[string]string
	flds   []string
}

type Human struct {
	secrets map[string]Question
}

var _ secrets.Keeper = &Human{}

func New() *Human {
	return &Human{
		secrets: make(map[string]Question, 1),
	}
}

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

func (h *Human) ListLocations(_ context.Context) ([]string, error) {
	return []string{""}, nil
}

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

func (h *Human) GetSecret(_ context.Context, id string) (secrets.Secret, error) {
	title := cases.Title(language.English)
	if def, defExists := h.secrets[id]; defExists {
		sec := secrets.NewSecret(id, "", "",
			secrets.WithID(id),
			secrets.WithLastModified(time.Now()))
		for _, fld := range def.flds {
			v, err := pinEntry(
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

func (h *Human) GetSecretsByName(ctx context.Context, name string) ([]secrets.Secret, error) {
	sec, err := h.GetSecret(ctx, name)
	if err != nil {
		return nil, err
	}
	return []secrets.Secret{sec}, nil
}

func (h *Human) SetSecret(_ context.Context, secret secrets.Secret) (secrets.Secret, error) {
	return nil, errors.New("read only")
}

func (h *Human) CopySecret(_ context.Context, id, location string) (secrets.Secret, error) {
	return nil, errors.New("read only")
}

func (h *Human) MoveSecret(_ context.Context, id, location string) (secrets.Secret, error) {
	return nil, errors.New("read only")
}

func (h *Human) DeleteSecret(_ context.Context, id string) error {
	return errors.New("read only")
}

// pinEntry is a tool that makes it easier to display a dialog prompting the
// user for a password.
func pinEntry(title, desc, prompt, ok string) (string, error) {
	pi, err := pinentry.New()
	if err != nil {
		return "", err
	}

	_ = pi.Set("title", title)
	_ = pi.Set("desc", desc)
	_ = pi.Set("prompt", prompt)
	_ = pi.Set("ok", ok)
	x, err := pi.GetPin()
	if err != nil {
		return "", err
	}

	return string(x), nil
}
