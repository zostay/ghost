package flag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zostay/ghost/pkg/config"
)

type Secret struct {
	*config.SecretRef
}

func (s *Secret) Set(value string) error {
	parts := strings.SplitN(value, ":", 3)
	if len(parts) != 3 {
		return errors.New("secret lookups must be in the form of <keeper>:<secret>:<field-name>")
	}

	s.KeeperName, s.SecretName, s.Field = parts[0], parts[1], parts[2]

	c := config.Instance()
	if _, hasKeeper := c.Keepers[s.KeeperName]; !hasKeeper {
		return fmt.Errorf("secret lookup names keeper %q which does not exist", s.KeeperName)
	}

	if s.SecretName == "" {
		return fmt.Errorf("empty secret identifier named")
	}

	if s.Field == "" {
		return fmt.Errorf("empty field name given")
	}

	return nil
}

func (s *Secret) String() string {
	return fmt.Sprintf("%s:%s:%s", s.KeeperName, s.SecretName, s.Field)
}

func (s *Secret) Type() string {
	return "keeper:secret:field"
}
