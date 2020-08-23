package updater

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFunctions(t *testing.T) {
	funcTests := []struct {
		name  string
		input []byte
		want  []byte
		f     ContentUpdater
	}{
		{"replace contents", []byte("input"), []byte("output"), ReplaceContents([]byte("output"))},
		{"update yaml key", []byte("input:\n  value: test\n"), []byte("input:\n  value: new\n"), UpdateYAML("input.value", "new")},
	}

	for _, tt := range funcTests {
		t.Run(tt.name, func(rt *testing.T) {
			got, err := tt.f(tt.input)

			if err != nil {
				rt.Errorf("got an error updating the content: %v", err)
				return
			}

			if diff := cmp.Diff(string(tt.want), string(got)); diff != "" {
				rt.Errorf("returned body failed:\n%s", diff)
			}
		})
	}
}
