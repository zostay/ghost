package keeper

import "github.com/gopasspw/pinentry"

// PinEntry is a tool that makes it easier to display a dialog prompting the
// user for a password.
func PinEntry(title, desc, prompt, ok string) (string, error) {
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
