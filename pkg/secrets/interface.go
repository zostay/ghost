package secrets

import (
	"net/url"
	"time"
)

// Secret is the interface for a secret.
type Secret interface {
	// ID returns the unique ID of the secret.
	ID() string

	// Name returns the name of the secret.
	Name() string

	// Username returns the username for the secret.
	Username() string

	// Secret returns the secret value.
	Password() string

	// Type returns the type of the secret.
	Type() string

	// Fields returns the fields for the secret.
	Fields() map[string]string

	// GetField returns the value of the named field.
	GetField(string) string

	// LastModified returns the last modified time for the secret.
	LastModified() time.Time

	// Url returns the URL for the secret.
	Url() *url.URL

	// Location returns the location for the secret.
	Location() string
}

// SettableName is the interface for a secret that can have its name set.
type SettableName interface {
	// SetName sets the name of the secret.
	SetName(string)
}

// SettableUsername is the interface for a secret that can have its username
// set.
type SettableUsername interface {
	// SetUsername sets the username for the secret.
	SetUsername(string)
}

// SettablePassword is the interface for a secret that can have its password
// value set.
type SettablePassword interface {
	// SetPassword sets the secret value.
	SetPassword(string)
}

// SettableType is the interface for a secret that can have its type set.
type SettableType interface {
	// SetType sets the type of the secret.
	SetType(string)
}

// SettableFields is the interface for a secret that can have its fields set.
type SettableFields interface {
	SetField(string, string)
	DeleteField(string)
}

// SettableLastModified is the interface for a secret that can have its last
// modified time set.
type SettableLastModified interface {
	// SetLastModified sets the last modified time for the secret.
	SetLastModified(time.Time)
}

// SettableUrl is the interface for a secret that can have its URL set.
type SettableUrl interface {
	// SetUrl sets the URL for the secret.
	SetUrl(*url.URL)
}

// UrlString is a helper that returns the string for a URL. If the URL is set,
// it returns the value returned by calling the String method on it. If not, it
// returns an empty string.
func UrlString(sec Secret) string {
	if sec.Url() == nil {
		return ""
	}
	return sec.Url().String()
}
