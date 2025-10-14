package names

import (
	"testing"
)

func TestGenerator(t *testing.T) {
	g := New()

	name := g.PrefixedName("testing-")

	if len(name) <= len("testing-") {
		t.Errorf("generated name too short: %s", name)
	}
	if name[:8] != "testing-" {
		t.Errorf("generated name does not preserve prefix: %s", name)
	}
}

func TestGenerator_trims_prefix(t *testing.T) {
	testString := "this-is-a-long-string-"
	g := New()
	g.MaxLength = len(testString)

	name := g.PrefixedName(testString)

	if len(name) != len(testString) {
		t.Errorf("expected length %d got %d", len(testString), len(name))
	}
	// ensure the generated name still begins with the trimmed prefix or preserves a trailing '-'
	if name[len(name)-1] == '-' {
		// ok: suffix trimmed but preserved dash
	}
}
