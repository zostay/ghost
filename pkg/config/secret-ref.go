package config

const SecretRefKey = "__SECRET__"

type SecretRef struct {
	KeeperName string `mapstructure:"keeper"`
	SecretName string `mapstructure:"secret"`
	Field      string `mapstructure:"field"`
}
