package set

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zostay/ghost/pkg/config"
	"github.com/zostay/ghost/pkg/secrets/keepass"
)

var (
	KeepassCmd = &cobra.Command{
		Use:    "keepass <keeper-name> [flags]",
		Short:  "Configure a Keepass secret keeper",
		Args:   cobra.ExactArgs(1),
		PreRun: PreRunSetKeepassKeeperConfig,
		RunE:   RunSetKeeperConfig,
	}

	keepassPath         string
	keepassPathSecret   string
	keepassMaster       string
	keepassMasterSecret string
)

func init() {
	KeepassCmd.Flags().StringVar(&keepassPath, "path", "", "Path to the Keepass database")
	KeepassCmd.Flags().StringVar(&keepassPathSecret, "path-secret", "", "Use a secret to lookup the path to the Keepass database")
	KeepassCmd.Flags().StringVar(&keepassMaster, "master", "", "The master password to use to unlock the Keepass database")
	KeepassCmd.Flags().StringVar(&keepassMasterSecret, "master-secret", "", "Use a secret to lookup and set the master password to use to unlock the Keepass database")
}

func PreRunSetKeepassKeeperConfig(cmd *cobra.Command, args []string) error {
	if keepassPath != "" && keepassPathSecret != "" {
		return errors.New("cannot use both --path and --path-secret")
	}

	if keepassMaster != "" && keepassMasterSecret != "" {
		return errors.New("cannot use both --master and --master-secret")
	}

	Replacement = map[string]any{
		"type": keepass.ConfigType,
	}

	if keepassPath != "" {
		Replacement["path"] = keepassPath
	}

	if keepassPathSecret != "" {
		var err error
		Replacement["path"], err = decodeSecretLookup(keepassPathSecret)
		if err != nil {
			return fmt.Errorf("error decoding --path-secret: %w", err)
		}
	}

	if keepassMaster != "" {
		Replacement["master"] = keepassMaster
	}

	if keepassMasterSecret != "" {
		var err error
		Replacement["master"], err = decodeSecretLookup(keepassMasterSecret)
		if err != nil {
			return fmt.Errorf("error decoding --master-secret: %w", err)
		}
	}

	return nil
}

func decodeSecretLookup(secret string) (map[string]any, error) {
	parts := strings.SplitN(secret, ":", 3)
	if len(parts) != 3 {
		return nil, errors.New("secret lookups must be in the form of <keeper>:<secret>:<field-name>")
	}

	keeperName, secretName, fieldName := parts[0], parts[1], parts[2]

	c := config.Instance()
	if _, hasKeeper := c.Keepers[keeperName]; !hasKeeper {
		return nil, fmt.Errorf("secret lookup names keeper %q which does not exist", keeperName)
	}

	if secretName == "" {
		return nil, fmt.Errorf("empty secret identifier named")
	}

	if fieldName == "" {
		return nil, fmt.Errorf("empty field name given")
	}

	return map[string]any{
		"keeper": keeperName,
		"secret": secretName,
		"field":  fieldName,
	}, nil
}
