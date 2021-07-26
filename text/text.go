package text

import (
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func Normalize(s string) (string, error) {
	f := transform.RemoveFunc(func(r rune) bool { return unicode.Is(unicode.Mn, r) })
	t := transform.Chain(norm.NFD, f, norm.NFC)
	r, _, err := transform.String(t, s)
	return r, err
}
