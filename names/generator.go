package names

import (
	"fmt"
	"math/rand"
)

const (
	branchMaxLength  = 100
	defaultSuffixLen = 5
)

// RandomGenerator generates a random name prefix.
type RandomGenerator struct {
	rand      *rand.Rand
	MaxLength int
	SuffixLen int
}

// New creates and returns a RandomGenerator.
func New(r *rand.Rand) *RandomGenerator {
	return &RandomGenerator{rand: r, MaxLength: branchMaxLength, SuffixLen: defaultSuffixLen}
}

// PrefixedName generates a name from the prefix with an additional random set
// of alphabetic characters.
//
// If the prefix + the length of the random set of characters would exceed the
// MaxLength, then the prefix will be trimmed to accomodate the random string.
//
// If the prefix ends with "-" this will be preserved in the trimmed string,
// before adding the prefix.
func (g RandomGenerator) PrefixedName(prefix string) string {
	charset := "abcdefghijklmnopqrstuvwyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, g.SuffixLen)
	for i := range b {
		b[i] = charset[g.rand.Intn(len(charset))]
	}
	lastChar := ""
	if len(prefix)+g.SuffixLen > g.MaxLength {
		trimPoint := g.MaxLength - g.SuffixLen
		if lc := prefix[len(prefix)-1:]; lc == "-" {
			lastChar = lc
			trimPoint -= 1
		}
		prefix = prefix[:trimPoint]
	}
	return fmt.Sprintf("%s%s%s", prefix, lastChar, b)
}
