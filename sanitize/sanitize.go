package sanitize

import (
	"fmt"
	"strings"
)

// ErrEmptyName is returned if an empty string is provided for sanitising.
var ErrEmptyName = invalidName("DNS name can not be empty")

// MaxDNSNameLength is the limit for some resources Name fields where they can
// be used as DNS names per RFC 1035.
const MaxDNSNameLength = 63

// MaxK8sValueLength is the limit for names that can be used as DNS subdomain
// values per RFC 1123.
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names
const MaxK8SValueLength = 253

// InvalidNameError is returned when a name can't be sanitized.
type InvalidNameError struct {
	msg string
}

func (m InvalidNameError) Error() string {
	return m.msg
}

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
		return "", invalidNamef("DNS name %q does not start with a valid character", name)
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
		return "", invalidNamef("DNS name %q exceeded maximum length of %v", name, MaxDNSNameLength)
	}

	if len(output) == 0 {
		return "", invalidNamef("DNS name %q sanitized is empty", name)
	}

	return output, nil
}

// DNS subdomains are DNS labels separated by '.', max 253 characters.
func SanitizeDNSDomain(name string) (string, error) {
	if name == "" {
		return "", ErrEmptyName
	}
	dnsSegments := strings.Split(name, ".")

	firstSegment := true
	output := ""
	for _, segment := range dnsSegments {
		sanitized, err := SanitizeDNSName(segment)
		if err != nil {
			return "", err
		}
		if firstSegment {
			output += sanitized
			firstSegment = false
			continue
		}
		output += "." + sanitized
	}

	if len(output) > MaxK8SValueLength {
		return "", invalidNamef("DNS name %q exceeded maximum length of %v", name, MaxK8SValueLength)
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

func invalidName(s string) InvalidNameError {
	return InvalidNameError{msg: s}
}
func invalidNamef(format string, a ...any) InvalidNameError {
	return InvalidNameError{msg: fmt.Sprintf(format, a...)}
}
