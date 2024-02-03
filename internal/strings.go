package internal

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var chains = &sync.Pool{
	New: func() any {
		return transform.Chain(
			norm.NFD,
			runes.Remove(runes.Predicate(func(r rune) bool {
				return !unicode.In(r, unicode.Number, unicode.Letter)
			})),
			norm.NFC,
		)
	},
}

func NormalizeStr(v string) (string, error) {
	chain, ok := chains.Get().(transform.Transformer)
	if !ok {
		return "", fmt.Errorf("invalid type on chains pool %T", chain)
	}
	defer chains.Put(chain)

	v, _, err := transform.String(chain, v)
	if !ok {
		return "", fmt.Errorf("normalize string %w", err)
	}

	v = strings.ToLower(v)

	return v, nil
}

func MustNormalizeStr(v string) string {
	out, err := NormalizeStr(v)
	if err != nil {
		panic("normalize str: " + err.Error())
	}
	return out
}
