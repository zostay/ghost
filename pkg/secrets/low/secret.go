package low

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/zostay/ghost/pkg/secrets"
)

// Secret is a secrets.Secret implementation that wraps secrets.Single with a
// settable id and with YAML marshalling and unmarshalling.
type Secret struct {
	secrets.Single

	id string
}

var _ yaml.Marshaler = &Secret{}
var _ yaml.Unmarshaler = &Secret{}

// ID returns the secret ID.
func (s *Secret) ID() string {
	if s.id == "" {
		return s.Single.ID()
	}
	return s.id
}

// SetID sets the secret ID.
func (s *Secret) SetID(id string) {
	s.id = id
}

// MarshalYAML marshals the secret to YAML.
func (s *Secret) MarshalYAML() (interface{}, error) {
	lm := s.Single.LastModified()
	if lm.IsZero() {
		lm = time.Now()
	}

	return map[string]any{
		"Name":         s.Single.Name(),
		"Username":     s.Single.Username(),
		"Password":     s.Single.Password(),
		"Type":         s.Single.Type(),
		"Location":     s.Single.Location(),
		"URL":          secrets.UrlString(s),
		"Fields":       s.Single.Fields(),
		"LastModified": lm.Unix(),
	}, nil
}

// UnmarshalYAML unmarshals the secret from YAML.
func (s *Secret) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.New("secret must be a mapping")
	}

	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Kind != yaml.ScalarNode {
			continue
		}

		if node.Content[i].Value == "Fields" &&
			node.Content[i+1].Kind == yaml.MappingNode {
			fNode := node.Content[i+1]
			for j := 0; j < len(fNode.Content); j += 2 {
				key := fNode.Content[j].Value
				val := fNode.Content[j+1].Value
				s.Single.SetField(key, val)
			}
		}

		if node.Content[i+1].Kind != yaml.ScalarNode {
			continue
		}

		switch node.Content[i].Value {
		case "Name":
			s.Single.SetName(node.Content[i+1].Value)
		case "Username":
			s.Single.SetUsername(node.Content[i+1].Value)
		case "Password":
			s.Single.SetPassword(node.Content[i+1].Value)
		case "Type":
			s.Single.SetType(node.Content[i+1].Value)
		case "Location":
			s.Single.SetLocation(node.Content[i+1].Value)
		case "URL":
			url, err := url.Parse(node.Content[i+1].Value)
			if err != nil {
				return err
			}
			s.Single.SetUrl(url)
		case "LastModified":
			us, err := strconv.ParseInt(node.Content[i+1].Value, 10, 64)
			if err != nil {
				return err
			}
			s.Single.SetLastModified(time.Unix(us, 0))
		}
	}

	return nil
}
