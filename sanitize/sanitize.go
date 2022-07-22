package sanitize

import (
	"errors"
	"fmt"
	"strings"
)

// ErrEmptyName is returned if an empty string is provided for sanitising.
var ErrEmptyName = errors.New("DNS name can not be empty")

// MaxDNSNameLength is the limit for some resources Name fields where they can
// be used as DNS names.
const MaxDNSNameLength = 63

// SanitizeDNSName sanitizes a string suitable for use in K8s resources that
// require a DNS 1035 compatible name.
//
// The name returned from here must conform to following rules (as per RFC 1035):
//  - length must be <= 63 characters
//  - must be all lower case alphanumeric characters or '-'
//  - must start with an alphabet
//  - must end with an alphanumeric character
func SanitizeDNSName(name string) (string, error) {
	if name == "" {
		return "", ErrEmptyName
	}
	runes := []rune(strings.ToLower(name))
	start := findIndex(runes, isAlpha)
	if start == len(runes) {
		return "", fmt.Errorf("DNS name %q does not start with a valid character", name)
	}

	end := max(start, len(runes)-1)
	for end > start && !isAlphanumeric(runes[end]) {
		end--
	}

	output := ""
	for i := start; i <= end; i++ {
		if isAllowedDNS(runes[i]) {
			output += string(runes[i])
		}
	}

	if len(output) > MaxDNSNameLength {
		return "", fmt.Errorf("DNS name %q exceeded maximum length of %v", name, MaxDNSNameLength)
	}

	if len(output) == 0 {
		return "", fmt.Errorf("DNS name %q sanitized is empty", name)
	}

	return output, nil
}

func isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isAlphanumeric(c rune) bool {
	return isAlpha(c) || (c >= '0' && c <= '9')
}

func isAllowedDNS(c rune) bool {
	return isAlphanumeric(c) || c == '-'
}

func findIndex(runes []rune, pred func(c rune) bool) int {
	for i := 0; i < len(runes); i++ {
		if pred(runes[i]) {
			return i
		}
	}
	return 0
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
