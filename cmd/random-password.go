package cmd

import (
	"bufio"
	"crypto/rand"
	"math/big"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/zostay/go-std/generic"
	"github.com/zostay/go-std/slices"

	s "github.com/zostay/ghost/cmd/shared"
)

var (
	randomCmd = &cobra.Command{
		Use:   "random-password [flags]",
		Short: "Generate a random password matching your specifications",
		Args:  cobra.NoArgs,
		Run:   RunRandomPassword,
	}

	lc, uc, digits, symbols float32
	length                  int
	chbs                    bool
	dictionary              string

	lcChars     = slices.FromRange[byte]('a', 'z', 1)
	ucChars     = slices.FromRange[byte]('A', 'Z', 1)
	symbolChars = []byte{
		'~', '`', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '+', '=',
		'[', '{', ']', '}', '\\', '|', ';', ':', '\'', '"', ',', '<', '.', '>', '/', '?',
	}
	digitChars = slices.FromRange[byte]('0', '9', 1)
)

func init() {
	randomCmd.Flags().Float32VarP(&lc, "lowercase-weight", "l", 0.4, "lowercase letter weight")
	randomCmd.Flags().Float32VarP(&uc, "uppercase-weight", "u", 0.3, "uppercase letter weight")
	randomCmd.Flags().Float32VarP(&digits, "digit-weight", "d", 0.2, "numeric digit weight")
	randomCmd.Flags().Float32VarP(&symbols, "symbol-weight", "s", 0.1, "wymbol character weight")
	randomCmd.Flags().IntVarP(&length, "length", "n", 20, "length of the password")
	randomCmd.Flags().BoolVarP(&chbs, "correct-horse-battery-staple", "x", false, "generate a password using the XKCD method")
	randomCmd.Flags().StringVarP(&dictionary, "dictionary", "D", "/usr/share/dict/words", "dictionary file to use for the XKCD method")
}

func RunRandomPassword(*cobra.Command, []string) {
	if chbs {
		runCHBS()
		return
	}

	totes := lc + uc + digits + symbols
	lc /= totes
	uc /= totes
	digits /= totes
	symbols /= totes

	lcCount := selectChars(lc, length)
	ucCount := selectChars(uc, length)
	digitCount := selectChars(digits, length)
	symbolCount := selectChars(symbols, length)

	for (lcCount + ucCount + digitCount + symbolCount) > length {
		switch {
		case lcCount > generic.Max(ucCount, generic.Max(digitCount, symbolCount)):
			lcCount--
		case ucCount > generic.Max(lcCount, generic.Max(digitCount, symbolCount)):
			ucCount--
		case digitCount > generic.Max(lcCount, generic.Max(ucCount, symbolCount)):
			digitCount--
		default:
			symbolCount--
		}
	}

	for (lcCount + ucCount + digitCount + symbolCount) < length {
		switch {
		case lcCount > 0 && lcCount < generic.Min(ucCount, generic.Min(digitCount, symbolCount)):
			lcCount++
		case ucCount > 0 && ucCount < generic.Min(lcCount, generic.Min(digitCount, symbolCount)):
			ucCount++
		case digitCount > 0 && digitCount < generic.Min(lcCount, generic.Min(ucCount, symbolCount)):
			digitCount++
		case symbolCount > 0:
			symbolCount++
		}
	}

	pw := sample(lcChars, lcCount)
	pw = append(pw, sample(ucChars, ucCount)...)
	pw = append(pw, sample(digitChars, digitCount)...)
	pw = append(pw, sample(symbolChars, symbolCount)...)

	shuffle(pw)

	s.Logger.Println(string(pw))
}

func selectChars(weight float32, length int) int {
	if weight > 0 {
		return int(generic.Max(1.0, weight*float32(length)))
	}
	return 0
}

var plainWord = regexp.MustCompile(`^\w+$`)

func runCHBS() {
	dr, err := os.Open(dictionary)
	if err != nil {
		s.Logger.Panic(err)
	}
	defer func() { _ = dr.Close() }()

	longWords := make([]string, 0, 1000)
	shortWords := make([]string, 0, 1000)
	dscanner := bufio.NewScanner(dr)
	for dscanner.Scan() {
		word := dscanner.Text()
		if plainWord.MatchString(word) {
			if len(word) > 6 {
				longWords = append(longWords, word)
			} else {
				shortWords = append(shortWords, word)
			}
		}
	}

	shuffle(shortWords)
	shuffle(longWords)

	pw := ""
	for len(pw) < length {
		if len(pw) > 0 {
			pw += " "
		}

		if randomInt(100) < 5 {
			pw += pick(longWords)
		} else {
			pw += pick(shortWords)
		}
	}

	s.Logger.Println(pw)
}

func sample[T any](from []T, count int) []T {
	out := make([]T, 0, count)
	for len(out) < count {
		out = append(out, pick(from))
	}
	return out
}

func pick[T any](from []T) T {
	max := big.NewInt(int64(len(from)))
	p, err := rand.Int(rand.Reader, max)
	if err != nil {
		s.Logger.Panic(err)
	}
	return from[p.Int64()]
}

func randomInt(max int) int {
	p, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		s.Logger.Panic(err)
	}
	return int(p.Int64())
}

func shuffle[T any](in []T) []T {
	for i := range in {
		j := randomInt(len(in))
		in[i], in[j] = in[j], in[i]
	}
	return in
}
