package internal

import (
	"fmt"
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
			runes.Remove(runes.In(unicode.Mn)),
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

	result, _, err := transform.String(chain, v)
	if !ok {
		return "", fmt.Errorf("normalize string %w", err)
	}

	return result, nil
}
