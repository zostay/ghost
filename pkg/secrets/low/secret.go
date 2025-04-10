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
	lm := s.LastModified()
	if lm.IsZero() {
		lm = time.Now()
	}

	return map[string]any{
		"Name":         s.Name(),
		"Username":     s.Username(),
		"Password":     s.Password(),
		"Type":         s.Type(),
		"Location":     s.Location(),
		"URL":          secrets.UrlString(s),
		"Fields":       s.Fields(),
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
				s.SetField(key, val)
			}
		}

		if node.Content[i+1].Kind != yaml.ScalarNode {
			continue
		}

		switch node.Content[i].Value {
		case "Name":
			s.SetName(node.Content[i+1].Value)
		case "Username":
			s.SetUsername(node.Content[i+1].Value)
		case "Password":
			s.SetPassword(node.Content[i+1].Value)
		case "Type":
			s.SetType(node.Content[i+1].Value)
		case "Location":
			s.SetLocation(node.Content[i+1].Value)
		case "URL":
			u, err := url.Parse(node.Content[i+1].Value)
			if err != nil {
				return err
			}
			s.SetUrl(u)
		case "LastModified":
			us, err := strconv.ParseInt(node.Content[i+1].Value, 10, 64)
			if err != nil {
				return err
			}
			s.SetLastModified(time.Unix(us, 0))
		}
	}

	return nil
}
