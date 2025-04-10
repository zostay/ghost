package onepassword

import (
	"net/url"
	"strings"
	"time"

	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/google/uuid"

	"github.com/zostay/ghost/pkg/secrets"
)

var hiddenFields = map[string]struct{}{
	"username": {},
	"password": {},
}

type Single struct {
	item *onepassword.Item
}

func newSecret(item *onepassword.Item) *Single {
	return &Single{
		item: item,
	}
}

func fromSecret(secret secrets.Secret) *Single {
	s := newSecret(
		&onepassword.Item{
			ID:    secret.ID(),
			Title: secret.Name(),
			URLs: []onepassword.ItemURL{{
				Label:   "website",
				URL:     secrets.UrlString(secret),
				Primary: true,
			}},
			Vault: onepassword.ItemVault{
				ID: secret.Location(),
			},
		},
	)

	for field, v := range secret.Fields() {
		s.setField(field, v)
	}

	return s
}

func (s *Single) fieldIndex(field string) (int, string, string) {
	var section string
	if strings.Contains(field, ".") {
		parts := strings.SplitN(field, ".", 2)
		section, field = parts[0], parts[1]
	}

	for i, f := range s.item.Fields {
		if section != "" {
			if f.Section == nil {
				continue
			}
			if f.Section.Label != section {
				continue
			}
		}

		if f.ID == field {
			return i, section, field
		}
	}

	return -1, section, field
}

func (s *Single) getField(field string) *onepassword.ItemField {
	idx, _, _ := s.fieldIndex(field)
	if idx >= 0 {
		return s.item.Fields[idx]
	}

	return &onepassword.ItemField{}
}

func (s *Single) setField(field, value string) {
	var (
		idx                int
		section, sectionId string

		typ     = onepassword.FieldTypeString
		purpose = onepassword.FieldPurposeNotes
	)

	idx, section, field = s.fieldIndex(field)
	if idx >= 0 {
		s.item.Fields[idx].Value = value
		return
	}

	if section != "" {
		for _, s := range s.item.Sections {
			if s.Label == section {
				sectionId = s.ID
				break
			}
		}

		if sectionId == "" {
			sectionId = uuid.NewString()
			s.item.Sections = append(s.item.Sections, &onepassword.ItemSection{
				ID:    sectionId,
				Label: section,
			})
		}
	}

	switch field {
	case "username":
		purpose = onepassword.FieldPurposeUsername
	case "password":
		purpose = onepassword.FieldPurposePassword
		typ = onepassword.FieldTypeConcealed
	}

	s.item.Fields = append(s.item.Fields, &onepassword.ItemField{
		ID:      field,
		Label:   field,
		Value:   value,
		Type:    typ,
		Purpose: purpose,
		Section: &onepassword.ItemSection{
			Label: section,
			ID:    sectionId,
		},
	})
}

func (s *Single) ID() string {
	return s.item.ID
}

func (s *Single) Name() string {
	return s.item.Title
}

func (s *Single) SetName(name string) {
	s.item.Title = name
}

func (s *Single) Username() string {
	return s.getField("username").Value
}

func (s *Single) SetUsername(username string) {
	s.setField("username", username)
}

func (s *Single) Password() string {
	return s.getField("password").Value
}

func (s *Single) SetPassword(password string) {
	s.setField("password", password)
}

func (s *Single) Type() string {
	return ""
}

func (s *Single) SetType(_ string) {}

func (s *Single) Fields() map[string]string {
	flds := make(map[string]string, len(s.item.Fields))
	for _, f := range s.item.Fields {
		if _, hidden := hiddenFields[f.ID]; hidden {
			continue
		}

		flds[f.ID] = f.Value
	}
	return flds
}

func (s *Single) GetField(field string) string {
	if _, hidden := hiddenFields[field]; hidden {
		return ""
	}

	return s.getField(field).Value
}

func (s *Single) SetField(field, value string) {
	s.setField(field, value)
}

func (s *Single) SetFields(fields map[string]string) {
	for k, v := range fields {
		s.setField(k, v)
	}
}

func (s *Single) LastModified() time.Time {
	return s.item.UpdatedAt
}

func (s *Single) Url() *url.URL {
	if len(s.item.URLs) == 0 {
		return nil
	}

	for _, u := range s.item.URLs {
		if u.Primary {
			ur, err := url.Parse(u.URL)
			if err != nil {
				return nil
			}

			return ur
		}
	}

	u, err := url.Parse(s.item.URLs[0].URL)
	if err != nil {
		return nil
	}

	return u
}

func (s *Single) SetUrl(nu *url.URL) {
	if len(s.item.URLs) == 0 {
		s.item.URLs = []onepassword.ItemURL{{}}
	}

	for _, u := range s.item.URLs {
		if u.Primary {
			u.Label = "website"
			u.URL = nu.String()
			return
		}
	}

	s.item.URLs[0].Label = "website"
	s.item.URLs[0].URL = nu.String()
	s.item.URLs[0].Primary = true
}

func (s *Single) Location() string {
	return s.item.Vault.ID
}

var _ secrets.Secret = (*Single)(nil)
