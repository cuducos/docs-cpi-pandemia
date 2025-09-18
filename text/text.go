package text

import (
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func Normalize(s string) (string, error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	r, _, err := transform.String(t, s)
	if err != nil {
		return "", err
	}
	if !utf8.ValidString(r) {
		r, err = charmap.ISO8859_1.NewDecoder().String(r)
		if err != nil {
			return "", err
		}
	}
	return r, err
}
