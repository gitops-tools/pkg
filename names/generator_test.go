package names

import (
	"math/rand"
	"testing"
)

func TestGenerator(t *testing.T) {
	g := New(rand.New(rand.NewSource(100)))

	name := g.PrefixedName("testing-")

	if name != "testing-DlPsU" {
		t.Fatalf("got %v, want %v", name, "testing-DlPsU")
	}
}

func TestGenerator_trims_prefix(t *testing.T) {
	testString := "this-is-a-long-string-"
	g := New(rand.New(rand.NewSource(100)))
	g.MaxLength = len(testString)

	name := g.PrefixedName(testString)

	want := "this-is-a-long-s-DlPsU"
	if name != want {
		t.Fatalf("got %s want %s", name, want)
	}
}
