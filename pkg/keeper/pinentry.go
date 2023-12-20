package keeper

import (
	"fmt"
	"os"

	"github.com/ncruces/zenity"
	"golang.org/x/term"
)

// GetPassword is a tool that makes it easier to display a dialog prompting the
// user for a password.
func GetPassword(title, desc, prompt, ok string) (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return getTermPassword(title, desc, prompt, ok)
	}

	return getGUIPassword(title, desc, prompt, ok)
}

func getGUIPassword(title, desc, prompt, ok string) (string, error) {
	_, x, err := zenity.Password(
		zenity.Title(title),
		zenity.EntryText(desc+"\n\n"+prompt),
		zenity.OKLabel(ok),
	)

	if err != nil {
		return "", err
	}

	return x, nil
}

func getTermPassword(title, desc, prompt, ok string) (string, error) { //nolint:unparam
	fmt.Print(prompt + ": ")
	x, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	fmt.Print("\n")

	return string(x), nil
}
