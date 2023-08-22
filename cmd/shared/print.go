package shared

import (
	"strings"

	"github.com/zostay/go-std/set"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/secrets"
)

func PrintSecret(sec secrets.Secret, showSecret bool, flds ...string) {
	fldSet := set.New[string](slices.Map(flds, strings.ToLower)...)

	Logger.Printf("%s:", sec.Name())
	if fldSet.Len() == 0 || fldSet.Contains("id") {
		Logger.Printf("  ID: %s", sec.ID())
	}
	if fldSet.Len() == 0 || fldSet.Contains("location") {
		Logger.Printf("  Location: %s", sec.Location())
	}
	if fldSet.Len() == 0 || fldSet.Contains("username") {
		Logger.Printf("  Username: %s", sec.Username())
	}
	if fldSet.Len() == 0 || fldSet.Contains("password") {
		pw := "<hidden>"
		if showSecret {
			pw = sec.Password()
		}
		Logger.Printf("  Password: %s", pw)
	}
	if fldSet.Len() == 0 || fldSet.Contains("url") {
		Logger.Printf("  URL: %v", sec.Url())
	}
	if fldSet.Len() == 0 || fldSet.Contains("last-modified") {
		Logger.Printf("  Modified: %v", sec.LastModified())
	}
	if fldSet.Len() == 0 || fldSet.Contains("type") {
		Logger.Printf("  Type: %s", sec.Type())
	}
	printedHeading := false
	for k, v := range sec.Fields() {
		if fldSet.Len() == 0 || fldSet.Contains(strings.ToLower(k)) {
			if !printedHeading {
				Logger.Print("  Fields:")
				printedHeading = true
			}
			Logger.Printf("    %s: %s", k, v)
		}
	}
}
