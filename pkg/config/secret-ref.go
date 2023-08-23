package config

// SecretRefKey is the key used to store the secret reference in the
// keeper config.
const SecretRefKey = "__SECRET__"

// SecretRef is a reference to a secret in another keeper.
type SecretRef struct {
	KeeperName string `mapstructure:"keeper"`
	SecretName string `mapstructure:"secret"`
	Field      string `mapstructure:"field"`
}
