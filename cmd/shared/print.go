package shared

import (
	"strings"

	"github.com/zostay/go-std/set"
	"github.com/zostay/go-std/slices"

	"github.com/zostay/ghost/pkg/secrets"
)

func PrintSecret(sec secrets.Secret, showSecret bool, flds ...string) {
	fldSet := set.New[string](slices.Map(flds, strings.ToLower)...)

	Printer.Printf("%s:", sec.Name())
	if (fldSet.Len() == 0 || fldSet.Contains("id")) && sec.ID() != "" {
		Printer.Printf("  ID: %s", sec.ID())
	}
	if (fldSet.Len() == 0 || fldSet.Contains("location")) && sec.Location() != "" {
		Printer.Printf("  Location: %s", sec.Location())
	}
	if (fldSet.Len() == 0 || fldSet.Contains("username")) && sec.Username() != "" {
		Printer.Printf("  Username: %s", sec.Username())
	}
	if (fldSet.Len() == 0 || fldSet.Contains("password")) && sec.Password() != "" {
		pw := "<hidden>"
		if showSecret {
			pw = sec.Password()
		}
		Printer.Printf("  Password: %s", pw)
	}
	if (fldSet.Len() == 0 || fldSet.Contains("url")) && sec.Url() != nil && sec.Url().String() != "" {
		Printer.Printf("  URL: %v", sec.Url())
	}
	if (fldSet.Len() == 0 || fldSet.Contains("last-modified")) && !sec.LastModified().IsZero() {
		Printer.Printf("  Modified: %v", sec.LastModified())
	}
	if (fldSet.Len() == 0 || fldSet.Contains("type")) && sec.Type() != "" {
		Printer.Printf("  Type: %s", sec.Type())
	}
	printedHeading := false
	for k, v := range sec.Fields() {
		if fldSet.Len() == 0 || fldSet.Contains(strings.ToLower(k)) {
			if !printedHeading {
				Printer.Print("  Fields:")
				printedHeading = true
			}
			Printer.Printf("    %s: %s", k, v)
		}
	}
}
