package names

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	branchMaxLength  = 100
	defaultSuffixLen = 5
)

// RandomGenerator generates a random name suffix.
type RandomGenerator struct {
	MaxLength int
	SuffixLen int
}

// New creates and returns a RandomGenerator.
// The generator uses crypto/rand for non-predictable suffixes.
func New() *RandomGenerator {
	return &RandomGenerator{MaxLength: branchMaxLength, SuffixLen: defaultSuffixLen}
}

// PrefixedName generates a name from the prefix with an additional random set
// of alphabetic characters.
//
// If the prefix + the length of the random set of characters would exceed the
// MaxLength, then the prefix will be trimmed to accommodate the random string.
//
// If the prefix ends with "-" this will be preserved in the trimmed string,
// before adding the suffix.
func (g RandomGenerator) PrefixedName(prefix string) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	suffixLen := g.SuffixLen
	if suffixLen <= 0 {
		suffixLen = defaultSuffixLen
	}

	b := make([]byte, suffixLen)
	charsetLen := int64(len(charset))
	for i := 0; i < suffixLen; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(charsetLen))
		if err != nil {
			// fallback: use 'x' on error (extremely unlikely)
			b[i] = 'x'
			continue
		}
		b[i] = charset[n.Int64()]
	}

	lastChar := ""
	if len(prefix)+suffixLen > g.MaxLength {
		trimPoint := g.MaxLength - suffixLen
		if trimPoint < 0 {
			trimPoint = 0
		}
		if len(prefix) > 0 && prefix[len(prefix)-1:] == "-" && trimPoint > 0 {
			lastChar = "-"
			trimPoint -= 1
		}
		if trimPoint < len(prefix) {
			prefix = prefix[:trimPoint]
		}
	}
	return fmt.Sprintf("%s%s%s", prefix, lastChar, string(b))
}
